package compile

import (
	"ahead/ast"
	"ahead/parser"
	"ahead/transform"
	"ahead/typecheck"
	"ahead/types"
	"bytes"
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

type Compiler struct {
	currBlock *ir.Block
	currFun   *ir.Func
	mod       *ir.Module
	PEnv      PointerEnv
	Types     map[ast.NodeHash]types.Type
	FEnv      map[string]*CFunc
	TypeDefs  map[string]lltypes.Type
	LabelNo   int
}

type CFunc struct {
	Func     *ir.Func
	RetPtr   value.Value
	RetBlock *ir.Block
}

var StrType lltypes.Type = lltypes.NewStruct(lltypes.I64, lltypes.I8Ptr)
var LenType lltypes.Type
var IntType = lltypes.I32
var ByteType = lltypes.I8
var BoolType = lltypes.I1
var FloatType = lltypes.Float
var Zero = constant.NewInt(IntType, 0)
var InitTrampoline value.Value
var AdjustTrampoline value.Value
var AllocClo value.Value

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
	case types.StructType:
		typeDef, ok := c.TypeDefs[t.Name]
		if ok {
			return typeDef
		}

		memberTypes := make([]lltypes.Type, len(t.MemberTypes))
		for i, member := range t.MemberTypes {
			memberTypes[i] = c.typeToLLType(member)
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

func (c *Compiler) GetType(node ast.Node) types.Type {
	return c.Types[ast.HashNode(node)]
}

func (c *Compiler) SetupTypes(prog *ast.Program) {
	StrType = c.mod.NewTypeDef("str", StrType)
	LenType = c.mod.NewTypeDef("len_t", lltypes.NewInt(32))

	for _, structDef := range prog.Structs {
		newDefPtr := c.typeToLLType(structDef.Type)
		newDef := newDefPtr.(*lltypes.PointerType).ElemType
		c.TypeDefs[structDef.Type.Name] = lltypes.NewPointer(c.mod.NewTypeDef(structDef.Type.Name, newDef))
	}
}

func (c *Compiler) SetupFuncs(prog *ast.Program) {
	InitTrampoline = c.mod.NewFunc(
		"llvm.init.trampoline",
		lltypes.Void,
		ir.NewParam("tramp", lltypes.I8Ptr),
		ir.NewParam("func", lltypes.I8Ptr),
		ir.NewParam("nval", lltypes.I8Ptr))
	AdjustTrampoline = c.mod.NewFunc(
		"llvm.adjust.trampoline",
		lltypes.I8Ptr,
		ir.NewParam("tramp", lltypes.I8Ptr))
	AllocClo = c.mod.NewFunc(
		"alloc_clo",
		lltypes.I8Ptr)

	abs := c.mod.NewFunc("abs", lltypes.I32, ir.NewParam("x", lltypes.I32))
	c.FEnv["abs"] = &CFunc{abs, nil, nil}

	for name, fun := range prog.Funcs {
		llRetType := c.typeToLLType(c.GetType(fun).(types.FuncType).RetType)
		params := make([]*ir.Param, 0)
		for i := 0; i < len(fun.Args); i++ {
			argName := fun.Args[i].(*ast.Ident).Value
			argType := c.typeToLLType(c.GetType(fun.Args[i]))
			newParam := ir.NewParam(argName, argType)
			// TODO do a better job of detecting the closure argument
			if strings.HasSuffix(argName, ".arg") {
				newParam.Attrs = append(newParam.Attrs, enum.ParamAttrNest)
			}
			params = append(params, newParam)
		}

		funPtr := c.mod.NewFunc(name, llRetType, params...)
		c.FEnv[name] = &CFunc{funPtr, nil, nil}
	}
}

func (c *Compiler) CompileFunc(name string, fun *ast.FunDef) {
	cFun, ok := c.FEnv[name]
	if !ok {
		panic("Function " + name + " not defined")
	}
	c.currFun = cFun.Func
	c.currBlock = c.currFun.NewBlock(c.getLabel(name + "_entry"))

	_, isVoid := c.GetType(fun).(types.FuncType).RetType.(types.NullType)

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
	retType := c.typeToLLType(c.GetType(fun).(types.FuncType).RetType)
	cFun.RetBlock = cFun.Func.NewBlock(c.getLabel(name + "_ret"))
	if !isVoid {
		retPtr := c.currBlock.NewAlloca(retType)
		cFun.RetPtr = retPtr
		cFun.RetBlock.NewRet(cFun.RetBlock.NewLoad(retPtr))
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
		if lineNo == len(fun.Body.Lines)-1 {
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

	c.mod = ir.NewModule()

	// Create type defs
	c.SetupTypes(prog)

	// Initialize all function pointers ahead of time
	c.SetupFuncs(prog)

	// Compile function bodies
	for name, fun := range prog.Funcs {
		c.CompileFunc(name, fun)
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
		callee := c.CompileNode(node.Fun)

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
			retVal = c.currBlock.NewLoad(ptr)
		}
	case *ast.CompNode:
		retVal = c.currBlock.NewICmp(node.LLPred(), c.CompileNode(node.Left), c.CompileNode(node.Right))
	case *ast.ReturnExp:
		cFun := c.FEnv[c.currFun.Name()]
		c.currBlock.NewStore(c.CompileNode(node.Target), cFun.RetPtr)
		c.currBlock.NewBr(cFun.RetBlock)
	case *ast.Closure:
		tuplePtr := c.CompileNode(node.ArgTup)
		execMem := c.currBlock.NewCall(AllocClo)
		sourceFuncPtr := c.CompileNode(node.Target)

		// Cast all ptr types
		sourceFuncBytePtr := c.currBlock.NewBitCast(sourceFuncPtr, lltypes.I8Ptr)
		tupleBytePtr := c.currBlock.NewBitCast(tuplePtr, lltypes.I8Ptr)

		c.currBlock.NewCall(InitTrampoline, execMem, sourceFuncBytePtr, tupleBytePtr)
		adjustedTrampPtr := c.currBlock.NewCall(AdjustTrampoline, execMem)

		newFunType := c.typeToLLType(c.GetType(node.NewFunc))
		castTrampPtr := c.currBlock.NewBitCast(adjustedTrampPtr, newFunType)
		retVal = castTrampPtr
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
		lenPtr := c.currBlock.NewGetElementPtr(strPtr, Zero, Zero)
		c.currBlock.NewStore(constant.NewInt(lltypes.I64, int64(len(node.Value))), lenPtr)

		// Store actual string pointer
		charPtr := c.currBlock.NewGetElementPtr(constArr, Zero, Zero)
		charPtrDest := c.currBlock.NewGetElementPtr(strPtr, Zero, constant.NewInt(IntType, 1))
		c.currBlock.NewStore(charPtr, charPtrDest)
		retVal = strPtr
	case *ast.ArrayLiteral:
		listType := c.GetType(node).(types.ArrayType)
		llListType := c.typeToLLType(listType).(*lltypes.PointerType).ElemType
		llSubtype := c.typeToLLType(listType.Subtype)
		list := c.currBlock.NewAlloca(llListType)

		// Set list length
		lenPtr := c.currBlock.NewGetElementPtr(list, constant.NewInt(IntType, 0), constant.NewInt(IntType, 0))
		c.currBlock.NewStore(constant.NewInt(IntType, int64(node.Length)), lenPtr)

		// Get array start ptr
		arr := c.currBlock.NewAlloca(lltypes.NewArray(uint64(node.Length), llSubtype))
		arrStart := c.currBlock.NewGetElementPtr(arr, constant.NewInt(IntType, 0), constant.NewInt(IntType, 0))

		// Set arr start pointer in list
		arrPtr := c.currBlock.NewGetElementPtr(list, constant.NewInt(IntType, 0), constant.NewInt(IntType, 1))
		c.currBlock.NewStore(arrStart, arrPtr)

		// Set all arr elements
		for i, val := range node.Exprs {
			compVal := c.CompileNode(val)
			elemPtr := c.currBlock.NewGetElementPtr(arr, constant.NewInt(IntType, int64(0)), constant.NewInt(IntType, int64(i)))
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
		fmt.Println(structType.Fields[0].Name())

		if structType.Fields[0].Name() == "len_t" {
			// Array type
			elemPtr := c.getListElemPtr(sliceable, index)
			retVal = c.currBlock.NewLoad(elemPtr)
		} else {
			// Tuple type
			elemPtr := c.currBlock.NewGetElementPtr(sliceable, Zero, index)
			retVal = c.currBlock.NewLoad(elemPtr)
		}
	case *ast.TupleLiteral:
		tupleType := c.typeToLLType(c.GetType(node)).(*lltypes.PointerType).ElemType
		tuplePtr := c.currBlock.NewAlloca(tupleType)

		for i, elem := range node.Exprs {
			elemPtr := c.CompileNode(elem)
			tupleElemPtr := c.currBlock.NewGetElementPtr(tuplePtr, Zero, constant.NewInt(lltypes.I32, int64(i)))
			c.currBlock.NewStore(elemPtr, tupleElemPtr)
		}

		retVal = tuplePtr
	case *ast.StructInstance:
		structType := c.typeToLLType(node.DefRef.Type).(*lltypes.PointerType).ElemType
		structPtr := c.currBlock.NewAlloca(structType)

		for i, member := range node.Values {
			valuePtr := c.CompileNode(member)
			memberPtr := c.currBlock.NewGetElementPtr(structPtr, Zero, constant.NewInt(lltypes.I32, int64(i)))
			c.currBlock.NewStore(valuePtr, memberPtr)
		}

		retVal = structPtr
	case *ast.StructAccess:
		structPtr := c.CompileNode(node.Target)

		structType := c.GetType(node.Target).(types.StructType)
		structOffset := structType.Offset(node.Field.(*ast.Ident).Value)
		memberPtr := c.currBlock.NewGetElementPtr(structPtr, Zero, constant.NewInt(IntType, int64(structOffset)))
		retVal = c.currBlock.NewLoad(memberPtr)
	default:
		panic("No compilation step defined for node of type: " + reflect.TypeOf(node).String())
	}

	return retVal
}

func (c *Compiler) compilePipeline(node *ast.Pipeline) value.Value {

	return nil
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

			targetAddr = c.currBlock.NewAlloca(targetLLType)
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
		structOffset := structType.Offset(target.Field.(*ast.Ident).Value)
		destPtr := c.currBlock.NewGetElementPtr(structPtr, Zero, constant.NewInt(lltypes.I32, int64(structOffset)))
		c.currBlock.NewStore(expPtr, destPtr)
	}

	return retVal
}

func (c *Compiler) getListElemPtr(list value.Value, index value.Value) value.Value {
	// Load the pointer to the array from the struct
	arrPtr := c.currBlock.NewGetElementPtr(list, constant.NewInt(IntType, 0), constant.NewInt(IntType, 1))
	// Load the pointer itself
	arrStart := c.currBlock.NewLoad(arrPtr)
	// Get the pointer for the specific element
	elemPtr := c.currBlock.NewGetElementPtr(arrStart, index)

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
