package interp

import (
	"ahead/parser"
	"ahead/transform"
	"ahead/types"
	"fmt"
	"testing"
)

func TestTypeCheck(t *testing.T) {
	prog := `
f{
	x = 1 + 6;
	y = "bob" + " hi!";
};
`
	p := parser.ParseProgram(prog)
	checker := types.NewTypeChecker()
	checker.TypeCheck(p.MainFunc)
	if (checker.TEnv["x"] != types.IntType{}) {
		t.Fatal("x not int")
	}
	if (checker.TEnv["y"] != types.StringType{}) {
		t.Fatal("y not string")
	}
}

func TestTypeInfer(t *testing.T) {
	src := `
func = f(a,b,c){
	a + b + c;
};
dep1 = f(a) {
	a;
};
dep2 = f(b) {
	b;
};
dep3 = f(c) {
	c;
};
d = 5;
p(func(dep1(d), dep2(4), dep3(5)));
`

	prog := parser.ParseProgram(src)
	transform.TransformAst(prog)
	i := NewInterpreter()
	i.Interp(prog)

	fmt.Println(i.Output())
}
