package main

import (
	"dandelion/compile"
	"dandelion/typecheck"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func parseArgs() (string, string, string, int) {
	outIR := flag.String("irfile", "", "Save the LLVM IR to the specified file rather than executing it")
	optLevel := flag.String("O", "1", "Optimization level to use (0-3)")
	output := flag.String("o", "", "Output the program as a binary with the specified name rather than executing it")
	flag.Parse()

	opt, err := strconv.Atoi(*optLevel)
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	var sourceFile string
	if flag.NArg() == 0 {
		sourceFile = ""
	} else {
		sourceFile = flag.Arg(0)
	}

	if *output != "" {
		*output, _ = filepath.Abs(*output)
	}

	return sourceFile, *output, *outIR, opt
}

func main() {
	typecheck.DebugTypeInf = false
	sourceFile, outBin, outIR, optLevel := parseArgs()

	var src []byte
	var err error
	if sourceFile == "" {
		src, err = ioutil.ReadAll(os.Stdin)
	} else {
		src, err = ioutil.ReadFile(flag.Arg(0))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading input:", err)
		os.Exit(1)
	}

	llvmIr := compile.CompileSource(string(src), optLevel)
	if outIR != "" {
		err = ioutil.WriteFile(outIR, []byte(llvmIr), os.ModePerm)
		return
	}

	binPath := compile.MakeBinary(llvmIr, optLevel, outBin)
	if binPath == "" {
		os.Exit(1)
	}

	if outBin != "" {
		os.Exit(0)
	}

	syscall.Exec(binPath, []string{}, []string{})
}
