package compile

import (
	"testing"
)

func TestCompileBasic(t *testing.T) {
	src := `
	a = 6;
	b = 7;
	d = a + b;
	c = "str";
	c + c;
`
	CompileOutput(src, "")
}
