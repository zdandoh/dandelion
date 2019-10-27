package transform

import (
	"ahead/parser"
	"testing"
)

func TestRename(t *testing.T) {
	src := `
f(){
	x = 5;
	if x > 10 {
		y = 30;
		x = x + 1;
		lol = f(a, b, c) {
			d = a + b + c;
		};
	};
	y = y + 1;
};
`

	destSrc := `
f(){
	x_1 = 5;
	if x_1 > 10 {
		y_1 = 30;
		x_1 = x_1 + 1;
		lol_1 = f(a_1, b_1, c_1) {
			d_1 = a_1 + b_1 + c_1;
		};
	};
	y_2 = y_2 + 1;
};
`
	prog := parser.ParseProgram(src)
	RenameIdents(prog)

	destProg := parser.ParseProgram(destSrc)
	if destProg.Funcs["main"].String() != prog.Funcs["main"].String() {
		t.Fatal("Source program not equal to dest program")
	}
}
