package compile

import (
	"ahead/ast"
	"ahead/parser"
	"ahead/transform"
	"ahead/typecheck"
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	lltypes "github.com/llir/llvm/ir/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

type Compiler struct {
	regCount int
}

func Compile(prog *ast.Program) string {
	m := ir.NewModule()

	main := m.NewFunc("main", lltypes.I32)
	block := main.NewBlock("")
	block.NewRet(constant.NewInt(lltypes.I32, 32))

	return fmt.Sprintf("%s", m)
}

func (c *Compiler) CompileNode(astNode ast.Node) {

}

func CompileOutput(progText string, output string) bool {
	prog := parser.ParseProgram(progText)
	transform.TransformAst(prog)

	_, err := typecheck.TypeCheck(prog)
	if err != nil {
		log.Fatal("Program doesn't type check: " + err.Error())
	}

	llvm_ir := Compile(prog)
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
