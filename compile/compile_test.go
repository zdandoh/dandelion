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

d = my_func(21, 32);
`

	CompileOutput(src, "")
}
