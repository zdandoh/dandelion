package main

import (
	"dandelion/compile"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	outfile := flag.String("irfile", "", "Save the LLVM IR to the specified file rather than executing it")
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

	llvmIr := compile.CompileSource(string(src))

	if *outfile != "" {
		err = ioutil.WriteFile(*outfile, []byte(llvmIr), os.ModePerm)
	} else {
		err = compile.ExecIR(llvmIr)
	}

	if err != nil {
		fmt.Println(err)
	}
}
