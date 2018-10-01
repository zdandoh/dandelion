package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ahead [filename]")
		return
	}

	code, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Couldn't read file '%s'", os.Args[1])
		return
	}

	tokens := tokenizeCode(string(code))
	main := compileTokens(tokens)
	main.Run()
}
