package compile

import (
	"dandelion/ast"
	"dandelion/errs"
	"dandelion/parser"
	"dandelion/transform"
	"dandelion/types"
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"math"
	"reflect"
	"strings"
)

type PointerEnv map[string]map[string]value.Value

func (e PointerEnv) Set(fName string, vName string, val value.Value) {
	funEnv, ok := e[fName]
	if !ok {
		e[fName] = make(map[string]value.Value)
		funEnv = e[fName]
	}

	funEnv[vName] = val
}

func (e PointerEnv) Get(fName string, vName string) (value.Value, bool) {
	funEnv, ok := e[fName]
	if !ok {
		return nil, false
	}

	val, stillOk := funEnv[vName]
	return val, stillOk
}

type CoroState struct {
	Cleanup *ir.Block
	Suspend *ir.Block
	Promise value.Value
}

type Compiler struct {
	currBlock  *ir.Block
	currFun    *ir.Func
	currCoro   *CoroState
	mod        *ir.Module
	PEnv       PointerEnv
	Types      map[ast.NodeHash]types.Type
	FEnv       map[string]*CFunc
	TypeDefs   map[string]lltypes.Type
	LabelNo    int
	prog       *ast.Program
	onBreak    *ir.Block
	onContinue *ir.Block
	typeTable  TypeTable
	bailBlock  bool
}

type CFunc struct {
	Func      *ir.Func
	RetPtr    value.Value
	RetBlock  *ir.Block
	RetBlocks map[*ir.Block]bool // We don't want to overwrite returns that have already been setup, keep track of which blocks we've already returned from
}

var StrType lltypes.Type = lltypes.NewStruct(lltypes.I64, lltypes.I8Ptr)

// AnyType is type tag, pointer to data (if reference), int data (if int)
var AnyType lltypes.Type = lltypes.NewStruct(lltypes.I32, lltypes.I8Ptr, IntType)
var CoroType lltypes.Type = lltypes.NewStruct(lltypes.I1, lltypes.I8Ptr)
var LenType lltypes.Type
var CapType = lltypes.I32
var IntType = lltypes.I32
var ByteType = lltypes.I8
var BoolType = lltypes.I1
var FloatType = lltypes.Float
var Zero = constant.NewInt(IntType, 0)
var One = constant.NewInt(IntType, 1)

func (c *Compiler) getLabel(label string) string {
	c.LabelNo++
	return fmt.Sprintf("%s_%d", label, c.LabelNo)
}

func (c *Compiler) llType(myType types.Type) lltypes.Type {
	switch t := myType.(type) {
	case types.BoolType:
		return BoolType
	case types.IntType:
		return IntType
	case types.ByteType:
		return ByteType
	case types.FloatType:
		return FloatType
	case types.StringType:
		return lltypes.NewPointer(StrType)
	case types.VoidType:
		return lltypes.Void
	case types.FuncType:
		retType := c.llType(t.RetType)
		argTypes := make([]lltypes.Type, 0)
		for _, arg := range t.ArgTypes {
			argTypes = append(argTypes, c.llType(arg))
		}
		return lltypes.NewPointer(lltypes.NewFunc(retType, argTypes...))
	case types.ArrayType:
		subtype := c.llType(t.Subtype)
		arrPtr := lltypes.NewPointer(subtype)
		return lltypes.NewPointer(lltypes.NewStruct(LenType, CapType, arrPtr))
	case types.CoroutineType:
		return lltypes.I8Ptr
	case types.StructType:
		typeDef, ok := c.TypeDefs[t.Name]
		if ok {
			return typeDef
		}

		structDef := c.prog.Struct(t.Name)
		memberTypes := make([]lltypes.Type, len(structDef.Members))
		for i, member := range structDef.Members {
			memberTypes[i] = c.llType(member.Type)
		}
		return lltypes.NewPointer(lltypes.NewStruct(memberTypes...))
	case types.TupleType:
		elemTypes := make([]lltypes.Type, len(t.Types))
		for i, elem := range t.Types {
			elemTypes[i] = c.llType(elem)
		}

		return lltypes.NewPointer(lltypes.NewStruct(elemTypes...))
	case types.AnyType:
		return lltypes.NewPointer(AnyType)
	default:
		panic(fmt.Sprintf("Unknown type: %v", reflect.TypeOf(myType)))
	}
}

func (c *Compiler) PromiseType(coroutineType types.CoroutineType) *lltypes.StructType {
	return lltypes.NewStruct(
		c.llType(coroutineType.Yields), lltypes.I32)
}

func (c *Compiler) Type(node ast.Node) types.Type {
	return c.Types[ast.HashNode(node)]
}

func (c *Compiler) SetType(node ast.Node, ty types.Type) {
	c.Types[ast.HashNode(node)] = ty
}

func (c *Compiler) SetupTypes(prog *ast.Program) {
	StrType = c.mod.NewTypeDef("str", StrType)
	LenType = c.mod.NewTypeDef("len_t", lltypes.NewInt(32))

	for i := 0; i < prog.StructCount(); i++ {
		structDef := prog.StructNo(i)
		structType := lltypes.NewStruct()
		c.TypeDefs[structDef.Type.Name] = lltypes.NewPointer(c.mod.NewTypeDef(structDef.Type.Name, structType))

		for _, member := range structDef.Members {
			structType.Fields = append(structType.Fields, c.llType(member.Type))
		}
	}
}

func (c *Compiler) CompileLoopBody(block *ast.Block, onBreak *ir.Block, onContinue *ir.Block) {
	currBreak := c.onBreak
	currContinue := c.onContinue
	c.onBreak = onBreak
	c.onContinue = onContinue

	c.CompileBlock(block)

	c.onBreak = currBreak
	c.onContinue = currContinue
}

func (c *Compiler) SetupFuncs(prog *ast.Program) {
	c.setupIntrinsics()

	abs := c.mod.NewFunc("abs", lltypes.I32, ir.NewParam("x", lltypes.I32))
	c.FEnv["abs"] = &CFunc{abs, nil, nil, nil}

	for name, fun := range prog.Funcs {
		llRetType := c.llType(c.Type(fun).(types.FuncType).RetType)
		params := make([]*ir.Param, 0)
		for i := 0; i < len(fun.Args); i++ {
			argName := fun.Args[i].(*ast.Ident).Value
			argType := c.llType(c.Type(fun.Args[i]))
			newParam := ir.NewParam(argName, argType)
			// TODO do a better job of detecting the closure argument
			if transform.IsCloArg(fun.Args[i]) || strings.HasPrefix(argName, "__this") {
				newParam.Attrs = append(newParam.Attrs, enum.ParamAttrNest)
			}
			if transform.IsCloArg(fun.Args[i]) {
				newParam.Typ = lltypes.I8Ptr
			}
			params = append(params, newParam)
		}

		funPtr := c.mod.NewFunc(name, llRetType, params...)
		c.FEnv[name] = &CFunc{funPtr, nil, nil, make(map[*ir.Block]bool)}
	}
}

func (c *Compiler) CompileFunc(name string, fun *ast.FunDef) {
	cFun, ok := c.FEnv[name]
	if !ok {
		panic("Function " + name + " not defined")
	}
	c.currFun = cFun.Func
	c.currBlock = c.currFun.NewBlock("entry")
	funType := c.Type(fun).(types.FuncType)
	if *fun.IsCoro {
		c.currBlock = c.SetupCoro(c.currBlock, c.currFun, funType.RetType.(types.CoroutineType))
	}

	_, isVoid := funType.RetType.(types.VoidType)
	if isVoid && len(fun.Body.Lines) == 0 {
		c.currBlock.NewRet(nil)
	}

	// Bind function args
	for i, arg := range fun.Args {
		argName := arg.(*ast.Ident).Value

		var storePtr value.Value = c.currBlock.Parent.Params[i]
		var argType lltypes.Type
		if transform.IsCloArg(arg) {
			// If the arg is the closure value, get the type for the related tuple and cast it to that type
			cloTupIdent := &ast.Ident{transform.CloArgToTupName(arg.(*ast.Ident).Value), ast.NoID}
			cloTupType := c.llType(c.Type(cloTupIdent))
			castTupPtr := c.currBlock.NewBitCast(c.currBlock.Parent.Params[i], cloTupType)
			argType = c.llType(c.Type(cloTupIdent))
			storePtr = castTupPtr
		} else {
			argType = c.llType(c.Type(arg))
		}

		argPtr := c.currBlock.NewAlloca(argType)
		c.currBlock.NewStore(storePtr, argPtr)

		c.PEnv.Set(c.currFun.Name(), argName, argPtr)
	}
	// Allocate space for return value & setup return block
	// If the return value is null, return void
	retType := c.llType(funType.RetType)
	cFun.RetBlock = cFun.Func.NewBlock(c.getLabel(name + "_ret"))
	if !isVoid {
		retPtr := c.currBlock.NewAlloca(retType)
		cFun.RetPtr = retPtr
		cFun.RetBlock.NewRet(NewLoad(cFun.RetBlock, retPtr))
	} else {
		cFun.RetBlock.NewRet(nil)
	}

	if name == "main" {
		c.currBlock.NewStore(constant.NewInt(IntType, 0), cFun.RetPtr)
	}
	for lineNo, line := range fun.Body.Lines {
		lastVal := c.CompileNode(line)
		if c.bailBlock {
			c.bailBlock = false
			c.currBlock.NewBr(cFun.RetBlock)
			break
		}

		// TODO support multiple returns & returns that aren't at the end of the block
		if lineNo == len(fun.Body.Lines)-1 && !*fun.IsCoro {
			if lastVal != nil && name != "main" && !isVoid {
				// Only auto-return when it's an expression
				c.currBlock.NewStore(lastVal, cFun.RetPtr)
			}
			c.currBlock.NewBr(cFun.RetBlock)
		}
	}
	if *fun.IsCoro {
		suspendRes := c.currBlock.NewCall(CoroSuspend, constant.None, constant.True)
		c.currBlock.NewSwitch(
			suspendRes,
			c.currCoro.Suspend,
			ir.NewCase(constant.NewInt(lltypes.I8, 1), c.currCoro.Cleanup))
	}
}

func Compile(prog *ast.Program, Types map[ast.NodeHash]types.Type) string {
	c := Compiler{}
	c.PEnv = make(PointerEnv)
	c.FEnv = make(map[string]*CFunc)
	c.TypeDefs = make(map[string]lltypes.Type)
	c.Types = Types
	c.prog = prog

	c.mod = ir.NewModule()

	// Create type defs
	c.SetupTypes(prog)

	// Init compiler type table
	c.SetupTypeTable()

	// Initialize all function pointers ahead of time
	c.SetupFuncs(prog)

	// Compile function bodies
	for name, fun := range prog.Funcs {
		c.CompileFunc(name, fun)
	}

	// Reorder allocas
	for _, fun := range c.mod.Funcs {
		c.reorderAllocas(fun)
	}
	return c.mod.String()
}

func (c *Compiler) CompileNode(astNode ast.Node) value.Value {
	var retVal value.Value

	switch node := astNode.(type) {
	case *ast.ParenExp:
		retVal = c.CompileNode(node.Exp)
	case *ast.Num:
		retVal = constant.NewInt(IntType, node.Value)
	case *ast.ByteExp:
		retVal = constant.NewInt(ByteType, int64(node.Value))
	case *ast.FloatExp:
		retVal = constant.NewFloat(FloatType, node.Value)
	case *ast.BoolExp:
		retVal = constant.NewBool(node.Value)
	case *ast.NullExp:
		nullType := c.llType(c.Type(node))
		retVal = constant.NewNull(nullType.(*lltypes.PointerType))
	case *ast.AddSub:
		rightNode := c.CompileNode(node.Right)
		leftNode := c.CompileNode(node.Left)
		addType := rightNode.Type()

		if addType.Equal(IntType) {
			switch node.Op {
			case "+":
				retVal = c.currBlock.NewAdd(leftNode, rightNode)
			case "-":
				retVal = c.currBlock.NewSub(leftNode, rightNode)
			}
		} else if addType.Equal(FloatType) {
			switch node.Op {
			case "+":
				retVal = c.currBlock.NewFAdd(leftNode, rightNode)
			case "-":
				retVal = c.currBlock.NewFSub(leftNode, rightNode)
			}
		} else if addType.Equal(c.llType(types.StringType{})) {
			retVal = c.strConcat(leftNode, rightNode)
		}

	case *ast.MulDiv:
		rightNode := c.CompileNode(node.Right)
		leftNode := c.CompileNode(node.Left)
		addType := rightNode.Type()

		if addType.Equal(IntType) {
			switch node.Op {
			case "*":
				retVal = c.currBlock.NewMul(leftNode, rightNode)
			case "/":
				retVal = c.currBlock.NewSDiv(leftNode, rightNode)
			}
		} else if addType.Equal(FloatType) {
			switch node.Op {
			case "*":
				retVal = c.currBlock.NewFMul(leftNode, rightNode)
			case "/":
				retVal = c.currBlock.NewFDiv(leftNode, rightNode)
			}
		}
	case *ast.Mod:
		compLeft := c.CompileNode(node.Left)
		compRight := c.CompileNode(node.Right)
		retVal = c.currBlock.NewSRem(compLeft, compRight)
	case *ast.Assign:
		retVal = c.compileAssign(node)
	case *ast.FunApp:
		baseType, methodName, isBaseMethod := c.checkBaseMethod(node.Fun)
		if isBaseMethod {
			retVal = c.compileBaseMethod(baseType, methodName, node.Args)
			break
		}

		var callee value.Value
		if node.Extern {
			callee = ir.NewGlobal(node.Fun.(*ast.Ident).Value, c.llType(c.Type(node.Fun)).(*lltypes.PointerType).ElemType)
		} else {
			callee = c.CompileNode(node.Fun)
		}

		argVals := make([]value.Value, 0)
		for _, arg := range node.Args {
			argVals = append(argVals, c.CompileNode(arg))
		}

		retVal = c.currBlock.NewCall(callee, argVals...)
	case *ast.Ident:
		inFEnv := false
		ptr, ok := c.PEnv.Get(c.currFun.Name(), node.Value)
		if !ok {
			var cFun *CFunc
			cFun, inFEnv = c.FEnv[node.Value]
			if !inFEnv {
				errs.Error(errs.ErrorValue, node, "unbound identifier")
				errs.CheckExit()
			}
			ptr = cFun.Func
		}

		if inFEnv {
			retVal = ptr
		} else {
			retVal = NewLoad(c.currBlock, ptr)
		}
	case *ast.CompNode:
		compLeft := c.CompileNode(node.Left)
		compRight := c.CompileNode(node.Right)
		retVal = c.currBlock.NewICmp(node.LLPred(), compLeft, compRight)
	case *ast.ReturnExp:
		cFun := c.FEnv[c.currFun.Name()]

		_, returned := cFun.RetBlocks[c.currBlock]
		if returned {
			break
		}

		cFun.RetBlocks[c.currBlock] = true
		storeVal := c.CompileNode(node.Target)
		c.currBlock.NewStore(storeVal, cFun.RetPtr)
		c.currBlock.NewBr(cFun.RetBlock)
		c.bailBlock = true
	case *ast.YieldExp:
		prevContinuation := c.currBlock.Term
		resumeBlock := c.currFun.NewBlock(c.getLabel("resume"))
		resumeBlock.Term = prevContinuation

		yieldValPtr := NewGetElementPtr(c.currBlock, c.currCoro.Promise, Zero, Zero)
		srcPtr := c.CompileNode(node.Target)
		c.currBlock.NewStore(srcPtr, yieldValPtr)
		suspendRes := c.currBlock.NewCall(CoroSuspend, constant.None, constant.NewBool(false))

		c.currBlock.NewSwitch(
			suspendRes,
			c.currCoro.Suspend,
			ir.NewCase(constant.NewInt(lltypes.I8, 0), resumeBlock),
			ir.NewCase(constant.NewInt(lltypes.I8, 1), c.currCoro.Cleanup))

		c.currBlock = resumeBlock
	case *ast.Closure:
		tuplePtr := c.CompileNode(node.ArgTup)
		sourceFuncPtr := c.CompileNode(node.Target)
		newFunType := c.llType(c.Type(node.NewFunc))

		retVal = c.extractFirstArg(c.currBlock, sourceFuncPtr, tuplePtr, newFunType)
	case *ast.If:
		prevContinuation := c.currBlock.Term

		cond := c.CompileNode(node.Cond)

		ifBody := c.currFun.NewBlock(c.getLabel("ifbody"))
		postIf := c.currFun.NewBlock(c.getLabel("postif"))
		ifBody.NewBr(postIf)
		postIf.Term = prevContinuation

		c.currBlock.NewCondBr(cond, ifBody, postIf)

		c.currBlock = ifBody
		c.CompileBlock(node.Body)

		c.currBlock = postIf
	case *ast.While:
		prevContinuation := c.currBlock.Term

		whileCondBlock := c.currFun.NewBlock(c.getLabel("whilecond"))
		c.currBlock.NewBr(whileCondBlock)

		c.currBlock = whileCondBlock
		cond := c.CompileNode(node.Cond)

		whileBody := c.currFun.NewBlock(c.getLabel("whilebody"))
		whileBody.NewBr(whileCondBlock)
		postWhile := c.currFun.NewBlock(c.getLabel("postwhile"))
		postWhile.Term = prevContinuation

		whileCondBlock.NewCondBr(cond, whileBody, postWhile)

		c.currBlock = whileBody
		c.CompileLoopBody(node.Body, postWhile, whileCondBlock)

		c.currBlock = postWhile
	case *ast.For:
		prevContinuation := c.currBlock.Term

		c.CompileNode(node.Init)

		forCondBlock := c.currFun.NewBlock(c.getLabel("forcond"))
		c.currBlock.NewBr(forCondBlock)

		c.currBlock = forCondBlock
		cond := c.CompileNode(node.Cond)

		forStep := c.currFun.NewBlock(c.getLabel("forstep"))
		c.currBlock = forStep
		c.CompileNode(node.Step)
		c.currBlock.NewBr(forCondBlock)

		forBody := c.currFun.NewBlock(c.getLabel("forbody"))
		forBody.NewBr(forStep)

		postFor := c.currFun.NewBlock(c.getLabel("postfor"))
		postFor.Term = prevContinuation

		forCondBlock.NewCondBr(cond, forBody, postFor)

		c.currBlock = forBody
		c.CompileLoopBody(node.Body, postFor, forStep)

		c.currBlock = postFor
	case *ast.ForIter:
		var compNode ast.Node

		iterType := c.Type(node.Iter)
		compNode, typeMap := parser.DesugarForIter(node.Body, node.Iter, node.Item, iterType)
		for newNode, newType := range typeMap {
			c.SetType(newNode, newType)
		}

		retVal = c.CompileNode(compNode)
	case *ast.Pipeline:
		compNode, typeMap := parser.DesugarPipeline(node, c.Type)
		for newNode, newType := range typeMap {
			c.SetType(newNode, newType)
		}

		retVal = c.CompileNode(compNode)
	case *ast.FlowControl:
		cFun := c.FEnv[c.currFun.Name()]

		_, returned := cFun.RetBlocks[c.currBlock]
		if returned {
			break
		}

		cFun.RetBlocks[c.currBlock] = true
		if node.Type == ast.FlowBreak {
			c.currBlock.NewBr(c.onBreak)
		} else if node.Type == ast.FlowContinue {
			c.currBlock.NewBr(c.onContinue)
		}
		c.bailBlock = true
	case *ast.BlockExp:
		prevContinuation := c.currBlock.Term
		block := c.currFun.NewBlock(c.getLabel("newblock"))
		block.Term = prevContinuation
		c.currBlock.NewBr(block)
		c.currBlock = block

		c.CompileBlock(node.Block)
	case *ast.StrExp:
		strPtr := c.currBlock.NewAlloca(StrType)

		constArr := c.mod.NewGlobalDef(c.getLabel("strconst"), constant.NewCharArrayFromString(node.Value))

		// Store string length
		lenPtr := NewGetElementPtr(c.currBlock, strPtr, Zero, Zero)
		c.currBlock.NewStore(constant.NewInt(lltypes.I64, int64(len(node.Value))), lenPtr)

		// Store actual string pointer
		charPtr := NewGetElementPtr(c.currBlock, constArr, Zero, Zero)
		charPtrDest := NewGetElementPtr(c.currBlock, strPtr, Zero, constant.NewInt(IntType, 1))
		c.currBlock.NewStore(charPtr, charPtrDest)
		retVal = strPtr
	case *ast.ArrayLiteral:
		listType := c.Type(node).(types.ArrayType)
		llListType := c.llType(listType).(*lltypes.PointerType).ElemType
		llSubtype := c.llType(listType.Subtype)
		list := MallocType(c.currBlock, llListType)

		// Set list length
		lenVal := constant.NewInt(IntType, int64(node.Length))
		c.setArrLen(list, lenVal)

		// Set list cap
		minCap := int64(math.Max(float64(node.Length), 8))
		capVal := constant.NewInt(lltypes.I32, minCap)
		c.setArrCap(list, capVal)

		// Get array start ptr
		subtypeSize := GetSize(c.currBlock, llSubtype)
		arrSize := c.currBlock.NewMul(subtypeSize, capVal)

		arr := c.currBlock.NewCall(Malloc, arrSize)
		arrStart := c.currBlock.NewBitCast(arr, lltypes.NewPointer(llSubtype))

		// Set arr start pointer in list
		c.setArrData(list, arrStart)

		// Set all arr elements
		for i, val := range node.Exprs {
			compVal := c.CompileNode(val)
			elemPtr := NewGetElementPtr(c.currBlock, arrStart, constant.NewInt(IntType, int64(i)))
			c.currBlock.NewStore(compVal, elemPtr)
		}

		retVal = list
	case *ast.SliceNode:
		sliceable := c.CompileNode(node.Arr)
		index := c.CompileNode(node.Index)

		targType := c.Type(node.Arr)
		_, isTup := targType.(types.TupleType)
		_, isList := targType.(types.ArrayType)
		_, isStr := targType.(types.StringType)

		if isList {
			// Setup bounds check
			len := c.arrLen(sliceable)
			c.setupBoundsCheck(len, index)

			elemPtr := c.getListElemPtr(sliceable, index)
			retVal = NewLoad(c.currBlock, elemPtr)
		} else if isTup || transform.IsCloArg(node.Arr) {
			elemPtr := NewGetElementPtr(c.currBlock, sliceable, Zero, index)
			retVal = NewLoad(c.currBlock, elemPtr)
		} else if isStr {
			dataPtrPtr := NewGetElementPtr(c.currBlock, sliceable, Zero, One)
			dataPtr := NewLoad(c.currBlock, dataPtrPtr)
			elemPtr := NewGetElementPtr(c.currBlock, dataPtr, index)
			retVal = NewLoad(c.currBlock, elemPtr)
		} else {
			panic("Unknown slice target: " + node.Arr.String())
		}
	case *ast.TupleAccess:
		tup := c.CompileNode(node.Tup)
		index := constant.NewInt(IntType, int64(node.Index))

		elemPtr := NewGetElementPtr(c.currBlock, tup, Zero, index)
		retVal = NewLoad(c.currBlock, elemPtr)
	case *ast.TupleLiteral:
		tupleType := c.llType(c.Type(node)).(*lltypes.PointerType).ElemType
		tuplePtr := MallocType(c.currBlock, tupleType)

		for i, elem := range node.Exprs {
			elemPtr := c.CompileNode(elem)
			tupleElemPtr := NewGetElementPtr(c.currBlock, tuplePtr, Zero, constant.NewInt(lltypes.I32, int64(i)))
			c.currBlock.NewStore(elemPtr, tupleElemPtr)
		}

		retVal = tuplePtr
	case *ast.BuiltinExp:
		retVal = c.compileBuiltin(node)
	case *ast.BeginExp:
		var lastVal value.Value
		for _, subNode := range node.Nodes {
			lastVal = c.CompileNode(subNode)
		}

		retVal = lastVal
	case *ast.TypeAssert:
		compTarg := c.CompileNode(node.Target)

		sourceType := c.Type(node.Target)
		_, isAny := sourceType.(types.AnyType)
		if !isAny {
			errs.Error(errs.ErrorValue, node, "can only use type assertion on 'any' type")
			errs.CheckExit()
		}

		typeTagPtr := NewGetElementPtr(c.currBlock, compTarg, Zero, Zero)
		typeTag := NewLoad(c.currBlock, typeTagPtr)
		targetTypeNo := constant.NewInt(lltypes.I32, int64(c.typeTable.GetNo(node.TargetType)))
		areEq := c.currBlock.NewICmp(enum.IPredEQ, typeTag, targetTypeNo)

		contBlock := c.currFun.NewBlock(c.getLabel("assertcont"))
		failBlock := c.currFun.NewBlock(c.getLabel("assertfail"))
		failBlock.NewCall(ThrowEx, constant.NewInt(lltypes.I32, 1))
		failBlock.NewUnreachable()

		contBlock.Term = c.currBlock.Term
		c.currBlock.NewCondBr(areEq, contBlock, failBlock)

		c.currBlock = contBlock

		// Actually convert the value to the correct type
		var valPtr value.Value
		targetLLType := c.llType(node.TargetType)
		_, isPtr := targetLLType.(*lltypes.PointerType)
		if isPtr {
			valPtr = NewGetElementPtr(c.currBlock, compTarg, Zero, constant.NewInt(lltypes.I32, 1))
		} else {
			valPtr = NewGetElementPtr(c.currBlock, compTarg, Zero, constant.NewInt(lltypes.I32, 2))
		}

		sourcePtr := c.currBlock.NewBitCast(valPtr, lltypes.NewPointer(targetLLType))
		retVal = c.currBlock.NewLoad(targetLLType, sourcePtr)
	case *ast.IsExp:
		checkTypeNo := c.typeTable.GetNo(node.CheckType)
		checkNodeType := c.Type(node.CheckNode)

		_, isCheckNodeAny := checkNodeType.(types.AnyType)
		if !isCheckNodeAny {
			checkNodeTypeNo := c.typeTable.GetNo(checkNodeType)
			if checkTypeNo == checkNodeTypeNo {
				return constant.True
			} else {
				return constant.False
			}
		}

		targetAny := c.CompileNode(node.CheckNode)
		tagPtr := NewGetElementPtr(c.currBlock, targetAny, Zero, Zero)
		tagVal := NewLoad(c.currBlock, tagPtr)
		return c.currBlock.NewICmp(enum.IPredEQ, tagVal, constant.NewInt(lltypes.I32, int64(checkTypeNo)))

	case *ast.StructInstance:
		structType := c.llType(node.DefRef.Type).(*lltypes.PointerType).ElemType
		structPtr := MallocType(c.currBlock, structType)

		for i, member := range node.Values {
			valuePtr := c.CompileNode(member)
			memberPtr := NewGetElementPtr(c.currBlock, structPtr, Zero, constant.NewInt(lltypes.I32, int64(i)))
			c.currBlock.NewStore(valuePtr, memberPtr)
		}

		retVal = structPtr
	case *ast.StructAccess:
		structPtr := c.CompileNode(node.Target)
		structType, isStructType := c.Type(node.Target).(types.StructType)
		if !isStructType {
			errs.Error(errs.ErrorValue, node, "can't use base method '%s' outside of call", node.Field)
			errs.CheckExit()
		}

		var structDef *ast.StructDef
		for i := 0; i < c.prog.StructCount(); i++ {
			s := c.prog.StructNo(i)
			if s.Type.Name == structType.Name {
				structDef = s
				break
			}
		}

		method := structDef.Method(node.Field.(*ast.Ident).Value)
		if method != nil {
			// Method handling
			targFun := c.FEnv[method.TargetName].Func
			structPtr := c.CompileNode(node.Target)
			finalFunType := c.llType(c.Type(node))
			retVal = c.extractFirstArg(c.currBlock, targFun, structPtr, finalFunType)
		} else {
			// Member handling
			structOffset := structDef.Offset(node.Field.(*ast.Ident).Value)
			memberPtr := NewGetElementPtr(c.currBlock, structPtr, Zero, constant.NewInt(IntType, int64(structOffset)))
			retVal = NewLoad(c.currBlock, memberPtr)
		}
	default:
		panic("No compilation step defined for node of type: " + reflect.TypeOf(node).String())
	}

	return retVal
}

func (c *Compiler) compileBuiltin(node *ast.BuiltinExp) value.Value {
	var retVal value.Value

	switch node.Type {
	case ast.BuiltinLen:
		retVal = c.compileLen(node)
	case ast.BuiltinNext:
		coroType := c.Type(node.Args[0]).(types.CoroutineType)
		targetCoro := c.CompileNode(node.Args[0])
		c.currBlock.NewCall(CoroResume, targetCoro)
		voidPromise := c.currBlock.NewCall(CoroPromise, targetCoro, constant.NewInt(lltypes.I32, 4), constant.False)
		promiseStruct := c.currBlock.NewBitCast(voidPromise, lltypes.NewPointer(c.PromiseType(coroType)))
		yieldPtr := NewGetElementPtr(c.currBlock, promiseStruct, Zero, Zero)
		retVal = NewLoad(c.currBlock, yieldPtr)
	case ast.BuiltinAny:
		target := node.Args[0]
		compTarget := c.CompileNode(target)
		targType := c.Type(target)
		_, isTargAny := targType.(types.AnyType)
		if isTargAny {
			retVal = compTarget
			break
		}

		anyPtr := MallocType(c.currBlock, AnyType)

		tagPtr := NewGetElementPtr(c.currBlock, anyPtr, Zero, Zero)
		typeNo := constant.NewInt(lltypes.I32, int64(c.typeTable.GetNo(targType)))
		c.currBlock.NewStore(typeNo, tagPtr)

		_, targetIsPtr := c.llType(targType).(*lltypes.PointerType)
		var valStorePtr value.Value
		if targetIsPtr {
			valStorePtr = NewGetElementPtr(c.currBlock, anyPtr, Zero, constant.NewInt(lltypes.I32, 1))
			compTarget = c.currBlock.NewBitCast(compTarget, lltypes.I8Ptr)
		} else {
			valStorePtr = NewGetElementPtr(c.currBlock, anyPtr, Zero, constant.NewInt(lltypes.I32, 2))
		}
		c.currBlock.NewStore(compTarget, valStorePtr)

		retVal = anyPtr
	case ast.BuiltinType:
		retVal = constant.NewInt(IntType, int64(c.typeTable.GetNo(c.Type(node))))
	case ast.BuiltinDone:
		handle := c.CompileNode(node.Args[0])
		retVal = c.currBlock.NewCall(CoroDone, handle)
	case ast.BuiltinStr:
		byteArr := c.CompileNode(node.Args[0])
		len := c.arrLen(byteArr)
		dataPtr := c.arrData(byteArr)
		retVal = c.createString(len, dataPtr)
	default:
		panic("No compilation step defined for builtin " + node.Type)
	}

	return retVal
}

func (c *Compiler) arrLen(arr value.Value) value.Value {
	lenPtr := NewGetElementPtr(c.currBlock, arr, Zero, Zero)
	lenVal := NewLoad(c.currBlock, lenPtr)
	return lenVal
}

func (c *Compiler) setArrLen(arr value.Value, newLen value.Value) {
	lenPtr := NewGetElementPtr(c.currBlock, arr, Zero, Zero)
	c.currBlock.NewStore(newLen, lenPtr)
}

func (c *Compiler) arrCap(arr value.Value) value.Value {
	capPtr := NewGetElementPtr(c.currBlock, arr, Zero, One)
	capVal := NewLoad(c.currBlock, capPtr)
	return capVal
}

func (c *Compiler) setArrCap(arr value.Value, newCap value.Value) {
	capPtr := NewGetElementPtr(c.currBlock, arr, Zero, One)
	c.currBlock.NewStore(newCap, capPtr)
}

func (c *Compiler) arrData(arr value.Value) value.Value {
	dataPtr := NewGetElementPtr(c.currBlock, arr, Zero, constant.NewInt(lltypes.I32, 2))
	dataVal := NewLoad(c.currBlock, dataPtr)
	return dataVal
}

func (c *Compiler) setArrData(arr value.Value, newData value.Value) {
	dataPtr := NewGetElementPtr(c.currBlock, arr, Zero, constant.NewInt(lltypes.I32, 2))
	c.currBlock.NewStore(newData, dataPtr)
}

func (c *Compiler) compileLen(node *ast.BuiltinExp) value.Value {
	var retVal value.Value

	targetType := c.Type(node.Args[0])
	switch ty := targetType.(type) {
	case types.ArrayType:
		targetArr := c.CompileNode(node.Args[0])
		retVal = c.arrLen(targetArr)
	case types.TupleType:
		retVal = constant.NewInt(IntType, int64(len(ty.Types)))
	case types.StringType:
		targetString := c.CompileNode(node.Args[0])
		lenPtr := NewGetElementPtr(c.currBlock, targetString, constant.NewInt(IntType, 0), constant.NewInt(IntType, 0))
		sizeVal := c.currBlock.NewLoad(lltypes.I64, lenPtr)
		retVal = c.currBlock.NewTrunc(sizeVal, lltypes.I32)
	default:
		panic("builtin function len not applicable to type " + reflect.TypeOf(targetType).String())
	}

	return retVal
}

func (c *Compiler) extractFirstArg(block *ir.Block, sourceFun value.Value, argVal value.Value, finalType lltypes.Type) value.Value {
	execMem := block.NewCall(AllocClo)

	// Cast all ptr types
	sourceFuncBytePtr := block.NewBitCast(sourceFun, lltypes.I8Ptr)
	tupleBytePtr := block.NewBitCast(argVal, lltypes.I8Ptr)

	block.NewCall(InitTrampoline, execMem, sourceFuncBytePtr, tupleBytePtr)
	adjustedTrampPtr := block.NewCall(AdjustTrampoline, execMem)

	castTrampPtr := c.currBlock.NewBitCast(adjustedTrampPtr, finalType)
	return castTrampPtr
}

func NewLoad(block *ir.Block, ptr value.Value) value.Value {
	return block.NewLoad(ptr.Type().(*lltypes.PointerType).ElemType, ptr)
}

func NewGetElementPtr(block *ir.Block, src value.Value, indicies ...value.Value) value.Value {
	return block.NewGetElementPtr(src.Type().(*lltypes.PointerType).ElemType, src, indicies...)
}

func GetSize(block *ir.Block, typ lltypes.Type) value.Value {
	sizePtr := NewGetElementPtr(block, constant.NewNull(lltypes.NewPointer(typ)), constant.NewInt(lltypes.I32, 1))
	size := block.NewPtrToInt(sizePtr, lltypes.I64)
	return size
}

func MallocType(block *ir.Block, typ lltypes.Type) value.Value {
	size := GetSize(block, typ)
	mem := block.NewCall(Malloc, size)
	castMem := block.NewBitCast(mem, lltypes.NewPointer(typ))
	return castMem
}

func (c *Compiler) reorderAllocas(fun *ir.Func) {
	allocas := make([]ir.Instruction, 0)
	for _, block := range fun.Blocks {
		newInsts := make([]ir.Instruction, 0)
		for _, inst := range block.Insts {
			alloca, isAlloca := inst.(*ir.InstAlloca)
			if isAlloca {
				allocas = append(allocas, alloca)
			} else {
				newInsts = append(newInsts, inst)
			}
		}
		block.Insts = newInsts
	}

	if len(fun.Blocks) > 0 {
		fun.Blocks[0].Insts = append(allocas, fun.Blocks[0].Insts...)
	}
}

func (c *Compiler) compileAssign(node *ast.Assign) value.Value {
	var retVal value.Value

	switch target := node.Target.(type) {
	case *ast.Ident:
		targetName := target.Value
		targetAddr, ok := c.PEnv.Get(c.currFun.Name(), targetName)
		if !ok {
			targetType, ok := c.Types[ast.HashNode(node.Target)]
			if !ok {
				panic("Identifier not in type environment: " + targetName)
			}
			targetLLType := c.llType(targetType)

			ptr, isPtr := targetLLType.(*lltypes.PointerType)
			isFunc := false
			if isPtr {
				_, isFunc = ptr.ElemType.(*lltypes.FuncType)
			}
			if isPtr && !isFunc {
				targetAddr = MallocType(c.currBlock, targetLLType)
			} else {
				targetAddr = c.currBlock.NewAlloca(targetLLType)
			}

			targetAddr.(value.Named).SetName(targetName)
			c.PEnv.Set(c.currFun.Name(), targetName, targetAddr)
		}

		compiledExpr := c.CompileNode(node.Expr)
		c.currBlock.NewStore(compiledExpr, targetAddr)
	case *ast.SliceNode:
		index := c.CompileNode(target.Index)
		list := c.CompileNode(target.Arr)

		var elemPtr value.Value
		arrType := c.Type(target.Arr)
		_, isTup := arrType.(types.TupleType)
		_, isArr := arrType.(types.ArrayType)
		if isTup {
			elemPtr = NewGetElementPtr(c.currBlock, list, Zero, index)
			retVal = NewLoad(c.currBlock, elemPtr)
		} else if isArr {
			// Setup bounds check
			len := c.arrLen(list)
			c.setupBoundsCheck(len, index)

			elemPtr = c.getListElemPtr(list, index)
		}

		srcPtr := c.CompileNode(node.Expr)
		c.currBlock.NewStore(srcPtr, elemPtr)
	case *ast.TupleAccess:
		index := constant.NewInt(IntType, int64(target.Index))
		list := c.CompileNode(target.Tup)

		var elemPtr value.Value
		arrType := c.Type(target.Tup)
		_, isTup := arrType.(types.TupleType)
		_, isArr := arrType.(types.ArrayType)
		if isTup {
			elemPtr = NewGetElementPtr(c.currBlock, list, Zero, index)
			retVal = NewLoad(c.currBlock, elemPtr)
		} else if isArr {
			// Setup bounds check
			len := c.arrLen(list)
			c.setupBoundsCheck(len, index)

			elemPtr = c.getListElemPtr(list, index)
		}

		srcPtr := c.CompileNode(node.Expr)
		c.currBlock.NewStore(srcPtr, elemPtr)
	case *ast.StructAccess:
		structPtr := c.CompileNode(target.Target)
		expPtr := c.CompileNode(node.Expr)

		structType := c.Type(target.Target).(types.StructType)
		structDef := c.prog.Struct(structType.Name)
		structOffset := structDef.Offset(target.Field.(*ast.Ident).Value)
		destPtr := NewGetElementPtr(c.currBlock, structPtr, Zero, constant.NewInt(lltypes.I32, int64(structOffset)))
		c.currBlock.NewStore(expPtr, destPtr)
	}

	return retVal
}

func (c *Compiler) setupBoundsCheck(len value.Value, index value.Value) {
	contBlock := c.currFun.NewBlock(c.getLabel("slicecont"))
	contBlock.Term = c.currBlock.Term

	failBlock := c.currFun.NewBlock(c.getLabel("slicefail"))
	failBlock.NewCall(IndexError, index)
	failBlock.NewUnreachable()

	boundCheck := c.currBlock.NewICmp(enum.IPredULT, index, len)
	c.currBlock.NewCondBr(boundCheck, contBlock, failBlock)

	c.currBlock = contBlock
}

func (c *Compiler) getListElemPtr(list value.Value, index value.Value) value.Value {
	// Load the pointer to the array from the struct
	arrStart := c.arrData(list)
	// Get the pointer for the specific element
	elemPtr := NewGetElementPtr(c.currBlock, arrStart, index)

	return elemPtr
}

func (c *Compiler) CompileBlock(block *ast.Block) {
	for _, line := range block.Lines {
		c.CompileNode(line)
		if c.bailBlock {
			c.bailBlock = false
			break
		}
	}
}
