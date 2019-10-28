package compile

import (
	"testing"
)

func TestCompileBasic(t *testing.T) {
	src := `
a = 6;
b = 7;
d = a + b + 78;
return d;
`
	CompileCheckExit(src, 91)
}

func TestCompileFunc(t *testing.T) {
	src := `
my_func = f(int a, int b) int {
	a + b;
};

other = f(int a, int b) int {
	my_func(a, b);
};

d = other(21, 32);
`

	CompileCheckExit(src, 53)
}

func TestCompileConditional(t *testing.T) {
	src := `
x = 100;
while x > 36 {
	x = x - 1;
};
return x;
`

	CompileCheckExit(src, 36)
}

func TestArrayCompile(t *testing.T) {
	src := `
[1, 2, 3];
`

	CompileCheckExit(src, 0)
}
