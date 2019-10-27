package compile

import (
	"ahead/ast"
	"ahead/parser"
	"ahead/transform"
	"ahead/typecheck"
	"ahead/types"
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
)

type Compiler struct {
	currBlock *ir.Block
	currFun   *ir.Func
	mod       *ir.Module
	PEnv      map[string]value.Value
	TEnv      map[string]types.Type
	FEnv      map[string]*CFunc
}

type CFunc struct {
	Func     *ir.Func
	RetPtr   value.Value
	RetBlock *ir.Block
}

var StrType = lltypes.NewStruct(lltypes.I64, lltypes.I8Ptr)
var IntType = lltypes.I32
var BoolType = lltypes.I1

func pointerType(t types.Type) bool {
	switch t.(type) {
	case *types.FuncType:
		return true
	}

	return false
}

func typeToLLType(myType types.Type) lltypes.Type {
	switch t := myType.(type) {
	case types.BoolType:
		return BoolType
	case types.IntType:
		return IntType
	case types.StringType:
		return StrType
	case types.ArrayType:
		subtype := typeToLLType(t.Subtype)
		return lltypes.NewStruct(lltypes.I64, lltypes.I64, lltypes.NewVector(0, subtype))
	case types.NullType:
		return lltypes.Void
	case *types.FuncType:
		retType := typeToLLType(t.RetType)
		argTypes := make([]lltypes.Type, 0)
		for _, arg := range t.ArgTypes {
			argTypes = append(argTypes, typeToLLType(arg))
		}
		return lltypes.NewPointer(lltypes.NewFunc(retType, argTypes...))
	default:
		panic("Unknown type: " + reflect.TypeOf(myType).String())
	}
}

func (c *Compiler) SetupFuncs(prog *ast.Program) {
	abs := c.mod.NewFunc("abs", lltypes.I32, ir.NewParam("x", lltypes.I32))
	c.FEnv["abs"] = &CFunc{abs, nil, nil}

	for name, fun := range prog.Funcs {
		llRetType := typeToLLType(fun.Type.RetType)
		params := make([]*ir.Param, 0)
		for i := 0; i < len(fun.Args); i++ {
			argName := fun.Args[i].(*ast.Ident).Value
			newParam := ir.NewParam(argName, typeToLLType(fun.Type.ArgTypes[i]))
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
	c.currBlock = c.currFun.NewBlock("")

	_, isVoid := fun.Type.RetType.(types.NullType)

	// Bind function args
	for i, arg := range fun.Args {
		argName := arg.(*ast.Ident).Value
		argType := typeToLLType(c.TEnv[argName])
		argPtr := c.currBlock.NewAlloca(argType)
		c.currBlock.NewStore(c.currBlock.Parent.Params[i], argPtr)
		c.PEnv[argName] = argPtr
	}
	// Allocate space for return value & setup return block
	retPtr := c.currBlock.NewAlloca(typeToLLType(fun.Type.RetType))
	cFun.RetPtr = retPtr
	cFun.RetBlock = cFun.Func.NewBlock("")
	cFun.RetBlock.NewRet(cFun.RetBlock.NewLoad(retPtr))

	for lineNo, line := range fun.Body.Lines {
		fmt.Printf("Compiling line %d of %s\n", lineNo+1, name)
		lastVal := c.CompileNode(line)
		// TODO support multiple returns & returns that aren't at the end of the block
		if lineNo == len(fun.Body.Lines)-1 && !isVoid {
			if name == "main" && c.currBlock.Term == nil {
				// Special case to return 0 from main if we didn't return anything
				c.currBlock.NewStore(constant.NewInt(IntType, 0), retPtr)
			} else {
				if lastVal != nil {
					// Only auto-return when it's an expression
					c.currBlock.NewStore(lastVal, retPtr)
				}
			}
			c.currBlock.NewBr(cFun.RetBlock)
		}
	}
}

func Compile(prog *ast.Program, TEnv map[string]types.Type) string {
	c := Compiler{}
	c.PEnv = make(map[string]value.Value)
	c.FEnv = make(map[string]*CFunc)
	c.TEnv = TEnv

	c.mod = ir.NewModule()

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
	case *ast.Num:
		retVal = constant.NewInt(IntType, node.Value)
	case *ast.AddSub:
		switch node.Op {
		case "+":
			retVal = c.currBlock.NewAdd(c.CompileNode(node.Left), c.CompileNode(node.Right))
		case "-":
			retVal = c.currBlock.NewSub(c.CompileNode(node.Left), c.CompileNode(node.Right))
		}
	case *ast.MulDiv:
		switch node.Op {
		case "*":
			retVal = c.currBlock.NewMul(c.CompileNode(node.Left), c.CompileNode(node.Right))
		case "/":
			retVal = c.currBlock.NewSDiv(c.CompileNode(node.Left), c.CompileNode(node.Right))
		}
	case *ast.Assign:
		targetName := node.Target.(*ast.Ident).Value
		targetAddr, ok := c.PEnv[targetName]
		if !ok {
			targetType, ok := c.TEnv[targetName]
			if !ok {
				panic("Identifier not in type environment: " + targetName)
			}
			targetLLType := typeToLLType(targetType)

			targetAddr = c.currBlock.NewAlloca(targetLLType)
			c.PEnv[targetName] = targetAddr
		}
		c.currBlock.NewStore(c.CompileNode(node.Expr), targetAddr)
	case *ast.FunApp:
		callee := c.CompileNode(node.Fun)

		argVals := make([]value.Value, 0)
		for _, arg := range node.Args {
			argVals = append(argVals, c.CompileNode(arg))
		}

		retVal = c.currBlock.NewCall(callee, argVals...)
	case *ast.Ident:
		inFEnv := false
		ptr, ok := c.PEnv[node.Value]
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
	case *ast.If:
		cond := c.CompileNode(node.Cond)

		ifBlock := c.currFun.NewBlock("")
		newBlock := c.currFun.NewBlock("")

		c.currBlock.NewCondBr(cond, ifBlock, newBlock)

		c.currBlock = ifBlock
		c.CompileBlock(node.Body)

		if c.currBlock.Term == nil {
			c.currBlock.NewBr(newBlock)
		}

		c.currBlock = newBlock
	default:
		panic("No compilation step defined for node of type: " + reflect.TypeOf(node).String())
	}

	return retVal
}

func (c *Compiler) CompileBlock(block *ast.Block) {
	for lineNo, line := range block.Lines {
		fmt.Printf("Compiling line %d of block\n", lineNo)
		c.CompileNode(line)
	}
}

func CompileOutput(progText string, output string) bool {
	prog := parser.ParseProgram(progText)
	transform.TransformAst(prog)

	fmt.Println(prog)
	tEnv, err := typecheck.TypeCheck(prog)
	if err != nil {
		log.Fatal("Program doesn't type check: " + err.Error())
	}

	llvm_ir := Compile(prog, tEnv)
	err = ioutil.WriteFile("llvm_ir.ll", []byte(llvm_ir), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	cmd := exec.Command("bash", "-c", `cat llvm_ir.ll | lli`)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))

	return false
}
