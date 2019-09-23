package main

import (
	"ahead/parser"
	"fmt"
)

func main() {
	prog := parser.ParseProgram(`
x = 12;
f{
	while x < 100 {
		x = x + 5;
		p(x);
	};
}();
`)

	fmt.Println(prog.MainFunc)
}
