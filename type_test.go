package main

import "testing"

func TestTypeCheck(t *testing.T) {
	prog := `
f{
	x = 1 + 6;
	y = "bob" + " hi!";
};
`
	p := ParseProgram(prog)
	checker := NewTypeChecker()
	checker.TypeCheck(p.MainFunc)
	if (checker.TEnv["x"] != IntType{}) {
		t.Fatal("x not int")
	}
	if (checker.TEnv["y"] != StringType{}) {
		t.Fatal("y not string")
	}
}
