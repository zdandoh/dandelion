package interp

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

	walker := &struct{ ast.BaseWalker }{}
	walker.WalkN = func(node ast.Node) ast.Node {
		switch t := node.(type) {
		case *ast.AddSub:
			return &ast.AddSub{t.Right, t.Left, t.Op}
		}

		return nil
	}
	walker.WalkB = func(b *ast.Block) *ast.Block {
		return nil
	}
	newAst := ast.WalkAst(parsed.MainFunc, walker)

	if parsed.MainFunc.String() == fmt.Sprintf("%s", newAst) {
		t.Fatal("Transformed AST equals un-transformed AST")
	}
}

func TestPipe(t *testing.T) {
	src := `
func = f(i, e, a) {
	p(e);
};
[10, 11, 12, 13, 14, 15] -> func;
`

	output := `
10
11
12
13
14
15
`
	if !CompareOutput(src, output) {
		t.FailNow()
	}
}

func TestPipeInline(t *testing.T) {
	src := `
[4, 5, 6, 7] -> f{p(e);};
`

	output := `
4
5
6
7
`
	if !CompareOutput(src, output) {
		t.FailNow()
	}
}

func TestPipeline(t *testing.T) {
	src := `
[1, 2, 3, 4, 5] -> f{ e + 1; } -> f{ e * 2; } -> f{ p(e); };
`

	output := `
4
6
8
10
12
`

	if !CompareOutput(src, output) {
		t.FailNow()
	}
}

func TestCommandExec(t *testing.T) {
	src := "`printf hello` -> f{p(e);};"

	if !CompareOutput(src, "hello") {
		t.FailNow()
	}
}

func TestEmptyPipe(t *testing.T) {
	src := "[] -> f{p(e);};"

	if !CompareOutput(src, "") {
		t.FailNow()
	}
}

func TestMod(t *testing.T) {
	src := `
val = 40 % 32;
p(val);
`

	if !CompareOutput(src, "8") {
		t.FailNow()
	}
}

func TestReturn(t *testing.T) {
	src := `
value = f(){
	var = 30;
	if var == 30 {
		return 5;
	};
	p("didnt return");
	7;
}();
p(value);
`

	if !CompareOutput(src, "5") {
		t.FailNow()
	}
}

func TestClosure(t *testing.T) {
	src := `
clo1 = f(x){
	x = x + 1;
	f(){
		x * 5;
	};
};

clo2 = clo1(6);
clo3 = clo1(2);
p(clo2());
p(clo3());
`

	output := `
35
15
`
	if !CompareOutput(src, output) {
		t.FailNow()
	}
}
