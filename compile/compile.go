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
	currFun *ir.Func
	mod *ir.Module
	PEnv map[string]value.Value
	TEnv map[string]types.Type
}

var StrType = lltypes.NewStruct(lltypes.I64, lltypes.I8Ptr)
var IntType = lltypes.I64
var BoolType = lltypes.I8

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
		return lltypes.NewFunc(retType, argTypes...)
	default:
		panic("Unknown type: " + reflect.TypeOf(myType).String())
	}
}

func Compile(prog *ast.Program, TEnv map[string]types.Type) string {
	c := Compiler{}
	c.PEnv = make(map[string]value.Value)
	c.TEnv = TEnv

	c.mod = ir.NewModule()

	c.currFun = c.mod.NewFunc("main", lltypes.I32)
	c.currBlock = c.currFun.NewBlock("")

	for name, fun := range prog.Funcs {
		llRetType := typeToLLType(fun.Type.RetType)
		params := make([]*ir.Param, 0)
		for i := 0; i < len(fun.Args); i++ {
			argName := fun.Args[i].(*ast.Ident).Value
			newParam := ir.NewParam(argName, typeToLLType(fun.Type.ArgTypes[i]))
			params = append(params, newParam)
		}

		c.mod.NewFunc(name, llRetType, params...)
	}

	for _, line := range prog.MainFunc.Body.Lines {
		c.CompileNode(line)
	}
	c.currBlock.NewRet(constant.NewInt(lltypes.I32, 0))

	return c.mod.String()
}

func (c *Compiler) CompileNode(astNode ast.Node) value.Value {
	var retVal value.Value

	switch node := astNode.(type) {
	case *ast.Num:
		retVal = constant.NewInt(IntType, node.Value)
	case *ast.AddSub:
		retVal = c.currBlock.NewAdd(c.CompileNode(node.Left), c.CompileNode(node.Right))
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
	case *ast.Ident:
		ptr, ok := c.PEnv[node.Value]
		if !ok {
			panic("Unbound identifier: " + node.Value)
		}

		retVal = c.currBlock.NewLoad(ptr)
	default:
		panic("No compilation step defined for node of type: " + reflect.TypeOf(node).String())
	}

	return retVal
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
