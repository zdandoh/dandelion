package compile

import (
	"bytes"
	"dandelion/ast"
	"dandelion/parser"
	"dandelion/transform"
	"dandelion/typecheck"
	"dandelion/types"
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"syscall"
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
	currBlock *ir.Block
	currFun   *ir.Func
	currCoro  *CoroState
	mod       *ir.Module
	PEnv      PointerEnv
	Types     map[ast.NodeHash]types.Type
	FEnv      map[string]*CFunc
	TypeDefs  map[string]lltypes.Type
	LabelNo   int
	prog      *ast.Program
}

type CFunc struct {
	Func      *ir.Func
	RetPtr    value.Value
	RetBlock  *ir.Block
	RetBlocks map[*ir.Block]bool // We don't want to overwrite returns that have already been setup, keep track of which blocks we've already returned from
}

var StrType lltypes.Type = lltypes.NewStruct(lltypes.I64, lltypes.I8Ptr)
var LenType lltypes.Type
var IntType = lltypes.I32
var ByteType = lltypes.I8
var BoolType = lltypes.I1
var FloatType = lltypes.Float
var Zero = constant.NewInt(IntType, 0)

func (c *Compiler) getLabel(label string) string {
	c.LabelNo++
	return fmt.Sprintf("%s_%d", label, c.LabelNo)
}

func (c *Compiler) typeToLLType(myType types.Type) lltypes.Type {
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
	case types.NullType:
		return lltypes.Void
	case types.FuncType:
		retType := c.typeToLLType(t.RetType)
		argTypes := make([]lltypes.Type, 0)
		for _, arg := range t.ArgTypes {
			argTypes = append(argTypes, c.typeToLLType(arg))
		}
		return lltypes.NewPointer(lltypes.NewFunc(retType, argTypes...))
	case types.ArrayType:
		subtype := c.typeToLLType(t.Subtype)
		arrPtr := lltypes.NewPointer(subtype)
		return lltypes.NewPointer(lltypes.NewStruct(LenType, arrPtr))
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
			memberTypes[i] = c.typeToLLType(member.Type)
		}
		return lltypes.NewPointer(lltypes.NewStruct(memberTypes...))
	case types.TupleType:
		elemTypes := make([]lltypes.Type, len(t.Types))
		for i, elem := range t.Types {
			elemTypes[i] = c.typeToLLType(elem)
		}

		return lltypes.NewPointer(lltypes.NewStruct(elemTypes...))
	default:
		panic(fmt.Sprintf("Unknown type: %v", reflect.TypeOf(myType)))
	}
}

func (c *Compiler) PromiseType(coroutineType types.CoroutineType) *lltypes.StructType {
	return lltypes.NewStruct(
		c.typeToLLType(coroutineType.Yields), lltypes.I32) // c.typeToLLType(coroutineType.Reads))
}

func (c *Compiler) GetType(node ast.Node) types.Type {
	return c.Types[ast.HashNode(node)]
}

func (c *Compiler) SetupTypes(prog *ast.Program) {
	StrType = c.mod.NewTypeDef("str", StrType)
	LenType = c.mod.NewTypeDef("len_t", lltypes.NewInt(32))

	for _, structDef := range prog.Structs {
		structType := lltypes.NewStruct()
		c.TypeDefs[structDef.Type.Name] = lltypes.NewPointer(c.mod.NewTypeDef(structDef.Type.Name, structType))

		for _, member := range structDef.Members {
			structType.Fields = append(structType.Fields, c.typeToLLType(member.Type))
		}
	}
}

func (c *Compiler) SetupFuncs(prog *ast.Program) {
	c.setupIntrinsics()

	abs := c.mod.NewFunc("abs", lltypes.I32, ir.NewParam("x", lltypes.I32))
	c.FEnv["abs"] = &CFunc{abs, nil, nil, nil}

	for name, fun := range prog.Funcs {
		llRetType := c.typeToLLType(c.GetType(fun).(types.FuncType).RetType)
		params := make([]*ir.Param, 0)
		for i := 0; i < len(fun.Args); i++ {
			argName := fun.Args[i].(*ast.Ident).Value
			argType := c.typeToLLType(c.GetType(fun.Args[i]))
			newParam := ir.NewParam(argName, argType)
			// TODO do a better job of detecting the closure argument
			if strings.HasSuffix(argName, ".arg") || strings.HasPrefix(argName, "__this") {
				newParam.Attrs = append(newParam.Attrs, enum.ParamAttrNest)
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
	funType := c.GetType(fun).(types.FuncType)
	if *fun.IsCoro {
		c.currBlock = c.SetupCoro(c.currBlock, c.currFun, funType.RetType.(types.CoroutineType))
	}

	_, isVoid := funType.RetType.(types.NullType)

	// Bind function args
	for i, arg := range fun.Args {
		argName := arg.(*ast.Ident).Value
		argType := c.typeToLLType(c.GetType(arg))
		argPtr := c.currBlock.NewAlloca(argType)
		c.currBlock.NewStore(c.currBlock.Parent.Params[i], argPtr)
		c.PEnv.Set(c.currFun.Name(), argName, argPtr)
	}
	// Allocate space for return value & setup return block
	// If the return value is null, return void
	retType := c.typeToLLType(funType.RetType)
	cFun.RetBlock = cFun.Func.NewBlock(c.getLabel(name + "_ret"))
	if !isVoid {
		retPtr := c.currBlock.NewAlloca(retType)
		cFun.RetPtr = retPtr
		cFun.RetBlock.NewRet(NewLoad(cFun.RetBlock, retPtr))
	} else {
		cFun.RetBlock.NewRet(nil)
	}

	for lineNo, line := range fun.Body.Lines {
		if lineNo == 0 && name == "main" {
			// Store 0 in the main return by default
			c.currBlock.NewStore(constant.NewInt(IntType, 0), cFun.RetPtr)
		}

		fmt.Printf("Compiling line %d of %s\n", lineNo+1, name)
		lastVal := c.CompileNode(line)
		// TODO support multiple returns & returns that aren't at the end of the block
		if lineNo == len(fun.Body.Lines)-1 && !*fun.IsCoro {
			if lastVal != nil && name != "main" && !isVoid {
				// Only auto-return when it's an expression
				c.currBlock.NewStore(lastVal, cFun.RetPtr)
			}
			c.currBlock.NewBr(cFun.RetBlock)
		}
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
		nullType := c.typeToLLType(c.GetType(node))
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
		retVal = c.currBlock.NewSRem(c.CompileNode(node.Left), c.CompileNode(node.Right))
	case *ast.Assign:
		retVal = c.compileAssign(node)
	case *ast.Pipeline:
		retVal = c.compilePipeline(node)
	case *ast.FunApp:
		var callee value.Value
		if node.Extern {
			callee = ir.NewGlobal(node.Fun.(*ast.Ident).Value, c.typeToLLType(c.GetType(node.Fun)).(*lltypes.PointerType).ElemType)
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
				panic("Unbound identifier: " + node.Value)
			}
			ptr = cFun.Func
		}

		if inFEnv {
			retVal = ptr
		} else {
			retVal = NewLoad(c.currBlock, ptr)
		}
	case *ast.CompNode:
		retVal = c.currBlock.NewICmp(node.LLPred(), c.CompileNode(node.Left), c.CompileNode(node.Right))
	case *ast.ReturnExp:
		cFun := c.FEnv[c.currFun.Name()]

		_, returned := cFun.RetBlocks[c.currBlock]
		if returned {
			break
		}

		cFun.RetBlocks[c.currBlock] = true
		c.currBlock.NewStore(c.CompileNode(node.Target), cFun.RetPtr)
		c.currBlock.NewBr(cFun.RetBlock)
	case *ast.YieldExp:
		prevContinuation := c.currBlock.Term
		resumeBlock := c.currFun.NewBlock(c.getLabel("resume"))
		resumeBlock.Term = prevContinuation

		yieldValPtr := NewGetElementPtr(c.currBlock, c.currCoro.Promise, Zero, Zero)
		c.currBlock.NewStore(c.CompileNode(node.Target), yieldValPtr)
		c.currBlock.NewCall(Print, c.CompileNode(node.Target))
		suspendRes := c.currBlock.NewCall(CoroSuspend, constant.None, constant.NewBool(false))

		c.currBlock.NewSwitch(
			suspendRes,
			c.currCoro.Suspend,
			ir.NewCase(constant.NewInt(lltypes.I8, 0), resumeBlock),
			ir.NewCase(constant.NewInt(lltypes.I8, 1), c.currCoro.Cleanup))

		c.currBlock = resumeBlock
	case *ast.NextExp:
		coroType := c.GetType(node.Target).(types.CoroutineType)
		targetCoro := c.CompileNode(node.Target)
		c.currBlock.NewCall(CoroResume, targetCoro)
		voidPromise := c.currBlock.NewCall(CoroPromise, targetCoro, constant.NewInt(lltypes.I32, 4), constant.False)
		promiseStruct := c.currBlock.NewBitCast(voidPromise, lltypes.NewPointer(c.PromiseType(coroType)))
		yieldPtr := NewGetElementPtr(c.currBlock, promiseStruct, Zero, Zero)
		retVal = NewLoad(c.currBlock, yieldPtr)
	case *ast.Closure:
		tuplePtr := c.CompileNode(node.ArgTup)
		sourceFuncPtr := c.CompileNode(node.Target)
		newFunType := c.typeToLLType(c.GetType(node.NewFunc))

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
		c.CompileBlock(node.Body)

		c.currBlock = postWhile
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
		listType := c.GetType(node).(types.ArrayType)
		llListType := c.typeToLLType(listType).(*lltypes.PointerType).ElemType
		llSubtype := c.typeToLLType(listType.Subtype)
		list := c.currBlock.NewAlloca(llListType)

		// Set list length
		lenPtr := NewGetElementPtr(c.currBlock, list, constant.NewInt(IntType, 0), constant.NewInt(IntType, 0))
		c.currBlock.NewStore(constant.NewInt(IntType, int64(node.Length)), lenPtr)

		// Get array start ptr
		arr := CallMalloc(c.currBlock, lltypes.NewArray(uint64(node.Length), llSubtype))
		arrStart := NewGetElementPtr(c.currBlock, arr, constant.NewInt(IntType, 0), constant.NewInt(IntType, 0))

		// Set arr start pointer in list
		arrPtr := NewGetElementPtr(c.currBlock, list, constant.NewInt(IntType, 0), constant.NewInt(IntType, 1))
		c.currBlock.NewStore(arrStart, arrPtr)

		// Set all arr elements
		for i, val := range node.Exprs {
			compVal := c.CompileNode(val)
			elemPtr := NewGetElementPtr(c.currBlock, arr, constant.NewInt(IntType, int64(0)), constant.NewInt(IntType, int64(i)))
			c.currBlock.NewStore(compVal, elemPtr)
		}

		retVal = list
	case *ast.SliceNode:
		sliceable := c.CompileNode(node.Arr)
		index := c.CompileNode(node.Index)

		// Do some hacky stuff here to allow for slicing different types without explicit ast info
		// If the type is a struct that starts with len_t, it's a list, otherwise it's a tuple
		ptrType := sliceable.Type().(*lltypes.PointerType).ElemType
		structType := ptrType.(*lltypes.StructType)

		if structType.Fields[0].Name() == "len_t" {
			// Array type
			elemPtr := c.getListElemPtr(sliceable, index)
			retVal = NewLoad(c.currBlock, elemPtr)
		} else {
			// Tuple type
			elemPtr := NewGetElementPtr(c.currBlock, sliceable, Zero, index)
			retVal = NewLoad(c.currBlock, elemPtr)
		}
	case *ast.TupleLiteral:
		tupleType := c.typeToLLType(c.GetType(node)).(*lltypes.PointerType).ElemType
		tuplePtr := c.currBlock.NewAlloca(tupleType)

		for i, elem := range node.Exprs {
			elemPtr := c.CompileNode(elem)
			tupleElemPtr := NewGetElementPtr(c.currBlock, tuplePtr, Zero, constant.NewInt(lltypes.I32, int64(i)))
			c.currBlock.NewStore(elemPtr, tupleElemPtr)
		}

		retVal = tuplePtr
	case *ast.StructInstance:
		structType := c.typeToLLType(node.DefRef.Type).(*lltypes.PointerType).ElemType
		structPtr := CallMalloc(c.currBlock, structType)

		for i, member := range node.Values {
			valuePtr := c.CompileNode(member)
			memberPtr := NewGetElementPtr(c.currBlock, structPtr, Zero, constant.NewInt(lltypes.I32, int64(i)))
			c.currBlock.NewStore(valuePtr, memberPtr)
		}

		retVal = structPtr
	case *ast.StructAccess:
		structPtr := c.CompileNode(node.Target)
		structType := c.GetType(node.Target).(types.StructType)

		var structDef *ast.StructDef
		for _, s := range c.prog.Structs {
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
			finalFunType := c.typeToLLType(c.GetType(node))
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

func (c *Compiler) compilePipeline(node *ast.Pipeline) value.Value {

	return nil
}

func NewLoad(block *ir.Block, ptr value.Value) value.Value {
	return block.NewLoad(ptr.Type().(*lltypes.PointerType).ElemType, ptr)
}

func NewGetElementPtr(block *ir.Block, src value.Value, indicies ...value.Value) value.Value {
	return block.NewGetElementPtr(src.Type().(*lltypes.PointerType).ElemType, src, indicies...)
}

func CallMalloc(block *ir.Block, typ lltypes.Type) value.Value {
	sizePtr := NewGetElementPtr(block, constant.NewNull(lltypes.NewPointer(typ)), constant.NewInt(lltypes.I32, 1))
	size := block.NewPtrToInt(sizePtr, lltypes.I64)
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
			targetLLType := c.typeToLLType(targetType)

			ptr, isPtr := targetLLType.(*lltypes.PointerType)
			isFunc := false
			if isPtr {
				_, isFunc = ptr.ElemType.(*lltypes.FuncType)
			}
			if isPtr && !isFunc {
				targetAddr = CallMalloc(c.currBlock, targetLLType)
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

		elemPtr := c.getListElemPtr(list, index)
		c.currBlock.NewStore(c.CompileNode(node.Expr), elemPtr)
	case *ast.StructAccess:
		structPtr := c.CompileNode(target.Target)
		expPtr := c.CompileNode(node.Expr)

		structType := c.GetType(target.Target).(types.StructType)
		structDef := c.prog.Struct(structType.Name)
		structOffset := structDef.Offset(target.Field.(*ast.Ident).Value)
		destPtr := NewGetElementPtr(c.currBlock, structPtr, Zero, constant.NewInt(lltypes.I32, int64(structOffset)))
		c.currBlock.NewStore(expPtr, destPtr)
	}

	return retVal
}

func (c *Compiler) getListElemPtr(list value.Value, index value.Value) value.Value {
	// Load the pointer to the array from the struct
	arrPtr := NewGetElementPtr(c.currBlock, list, constant.NewInt(IntType, 0), constant.NewInt(IntType, 1))
	// Load the pointer itself
	arrStart := NewLoad(c.currBlock, arrPtr)
	// Get the pointer for the specific element
	elemPtr := NewGetElementPtr(c.currBlock, arrStart, index)

	return elemPtr
}

func (c *Compiler) CompileBlock(block *ast.Block) {
	for lineNo, line := range block.Lines {
		fmt.Printf("Compiling line %d of block\n", lineNo)
		c.CompileNode(line)
	}
}

func CompileSource(progText string) string {
	prog := parser.ParseProgram(progText)
	transform.TransformAst(prog)

	progTypes := typecheck.Infer(prog)
	llvmIr := Compile(prog, progTypes)

	return llvmIr
}

func ExecIR(llvmIr string) error {
	cmd := exec.Command("lli")
	buffer := bytes.NewBufferString(llvmIr)

	cmd.Stdin = buffer
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		log.Fatalf(err.Error())
	}

	exitStatus := 0
	err = cmd.Wait()
	if err != nil {
		exitCode, ok := err.(*exec.ExitError)
		if ok {
			status, ok := exitCode.Sys().(syscall.WaitStatus)
			if ok {
				exitStatus = status.ExitStatus()
			}
		}
	}

	os.Exit(exitStatus)
	return nil
}

func CompileCheckExit(progText string, code int) bool {
	prog := parser.ParseProgram(progText)
	fmt.Println(prog)
	transform.TransformAst(prog)

	progTypes := typecheck.Infer(prog)

	llvm_ir := Compile(prog, progTypes)
	fmt.Println(llvm_ir)
	err := ioutil.WriteFile("llvm_ir.ll", []byte(llvm_ir), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	output, err := exec.Command("bash", "-i", "tester.sh").Output()
	if err != nil {
		log.Println(string(output))
		log.Fatalf(err.Error())
	}

	outputStr := strings.TrimSpace(string(output))
	exitCode, err := strconv.Atoi(outputStr)
	if err != nil {
		log.Fatalln(outputStr, err)
	}

	if exitCode != code {
		log.Println(outputStr)
		return false
	}

	return true
}
