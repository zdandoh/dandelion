package main

import (
	"ahead/interp"
	"ahead/parser"
	"ahead/transform"
	"fmt"
	"io/ioutil"
	"os"
)

func RunProgram(src string) {
	prog := parser.ParseProgram(src)
	transform.TransformAst(prog)

	i := interp.NewInterpreter()
	i.Interp(prog)
}

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

	RunProgram(string(src))
}
