package main

import (
	"dandelion/compile"
	"dandelion/typecheck"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	typecheck.DebugTypeInf = false
	outIR := flag.String("irfile", "", "Save the LLVM IR to the specified file rather than executing it")
	flag.Parse()

	var src []byte
	var err error
	if flag.NArg() == 0 {
		src, err = ioutil.ReadAll(os.Stdin)
	} else {
		src, err = ioutil.ReadFile(flag.Arg(0))
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	llvmIr := compile.CompileSource(string(src), 1)

	if *outIR != "" {
		err = ioutil.WriteFile(*outIR, []byte(llvmIr), os.ModePerm)
	} else {
		err = compile.ExecIR(llvmIr)
	}

	if err != nil {
		fmt.Println(err)
	}
}
