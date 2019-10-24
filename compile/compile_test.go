package compile

import (
	"testing"
)

func TestCompileBasic(t *testing.T) {
	src := `
a = 6;
b = 7;
d = a + b + 78;
`
	CompileOutput(src, "")
}

func TestCompileFunc(t *testing.T) {
	src := `
my_func = f(int a, int b) int {
	a + b;
};

other = f(int a, int b) int {
	my_func(a, b);
}

d = other(21, 32);
`

	CompileOutput(src, "")
}
