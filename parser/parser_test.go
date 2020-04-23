package parser

import (
	"dandelion/ast"
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

func TestParsePipeline(t *testing.T) {
	src := `
[1, 2, 3] -> f{ e = e + 1; } -> p;
`

	fmt.Println(ParseProgram(src))
}

func TestDesugarIterFor(t *testing.T) {
	fmt.Println(ForIterToFor(&ast.Block{[]ast.Node{&ast.Ident{"line", 0}}}, &ast.Ident{"iter()", 0}, &ast.Ident{"item", 0}))
}
