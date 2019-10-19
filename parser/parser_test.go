package parser

import (
	"fmt"
	"testing"
)

func TestTypedFunc(t *testing.T) {
	src := `
f(string name, int number) int {
	name;
};
`

	prog := ParseProgram(src)
	fmt.Println(prog)
}

func TestTypedFunc2(t *testing.T) {
	src := `
f(string[] names, int[] numbers) f(int, int) int {
	f(int a, int b) int {
		a + b;
	};
};
`

	prog := ParseProgram(src)
	fmt.Println(prog)
}
