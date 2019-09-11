package main

import (
	"fmt"
	"time"
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

	fmt.Println(prog.mainFunc)

	n := time.Now()
	prog.interp()
	fmt.Println("TIME:", time.Since(n))
}
