package main

import (
	"ahead/ast"
	"ahead/parser"
	"fmt"
	"testing"
)

func TestWhile(t *testing.T) {
	prog := `
x = 12;
while x < 20 {
	x = x + 5;
	p(x);
};`

	output := `17
22`
	if !CompareOutput(prog, output) {
		t.Fatal("Output not equal")
	}
}

func TestFunction(t *testing.T) {
	prog := `
func = f(a, b, c){
	a + b + c;
};
p(func(4, 5, 6));`

	output := "15"
	if !CompareOutput(prog, output) {
		t.Fatal("Output not equal")
	}
}

func TestBlockScope(t *testing.T) {
	prog := `
x = 5;
func = f(a, b, c) {
	x * a + b + c;
};
x = 10;
p(func(3, 2, 1));`

	output := "33"
	if !CompareOutput(prog, output) {
		t.Fatal("Output not equal")
	}
}

func TestArray(t *testing.T) {
	prog := `
x = 7;
arr = [1, 2, x, 4];
p(arr);`

	output := "[1, 2, 7, 4]"
	if !CompareOutput(prog, output) {
		t.Fatal("Output not equal")
	}
}

func TestSlice(t *testing.T) {
	prog := `
arr = [1, 2, 3, 4];
p(arr[2]);`

	output := "3"
	if !CompareOutput(prog, output) {
		t.Fatal("Output not equal")
	}
}

func TestString(t *testing.T) {
	prog := `
str = "hello, world!";
p(str);`
	output := "hello, world!"
	if !CompareOutput(prog, output) {
		t.Fatal("Output not equal")
	}
}

func TestApplyFunc(t *testing.T) {
	prog := `
f{
	x = 1 + 6;
	y = "bob" * x;
};
`

	parsed := parser.ParseProgram(prog)

	newAst := ast.ApplyFunc(parsed.MainFunc, func(node ast.Node) ast.Node {
		switch t := node.(type) {
		case *ast.AddSub:
			return &ast.AddSub{t.Right, t.Left, t.Op}
		}

		return nil
	})

	if parsed.MainFunc.String() == fmt.Sprintf("%s", newAst) {
		t.Fatal("Transformed AST equals un-transformed AST")
	}
}
