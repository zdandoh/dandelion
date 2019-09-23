package main

import (
	"ahead/interp"
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

	i := interp.NewInterpreter()
	i.Interp(prog)
	fmt.Println(prog.MainFunc)
}
