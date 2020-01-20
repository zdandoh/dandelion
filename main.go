package main

import (
	"ahead/compile"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	var src []byte
	var err error
	if len(os.Args) < 2 {
		src, err = ioutil.ReadAll(os.Stdin)
	} else {
		src, err = ioutil.ReadFile(os.Args[1])
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	llvmIr := compile.CompileSource(string(src))
	err = compile.ExecIR(llvmIr)
	fmt.Println(err)
}
