package main

import (
	"fmt"
	"io/ioutil"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func main() {
	// Create convenience types and constants.
	i32 := types.I32
	// Create a new LLVM IR module.
	m := ir.NewModule()

	rand := m.NewFunc("main", i32)

	// Create an unnamed entry basic block and append it to the `rand` function.
	entry := rand.NewBlock("")

	// Create instructions and append them to the entry basic block.
	tmp1 := constant.NewInt(i32, 5)
	tmp2 := constant.NewInt(i32, 65)
	temp3 := entry.NewAdd(tmp1, tmp2)

	entry.NewRet(temp3)

	// Print the LLVM IR assembly of the module.
	fmt.Println(m)

	ioutil.WriteFile("test.ll", []byte(m.String()), 0777)
}
