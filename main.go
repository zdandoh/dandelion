package main

import (
	"fmt"
)

func main() {
	prog := ParseProgram(`
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
