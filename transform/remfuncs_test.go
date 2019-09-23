package transform

import (
	"ahead/ast"
	"ahead/parser"
	"testing"
)

func TestRem(t *testing.T) {
	src := `
f{
	f{
		f{
			p(4 + 5);
		}();
	}();
}();
`

	prog := parser.ParseProgram(src)
	RemFuncs(prog)

	if len(prog.Funcs) != 3 {
		// Didn't gen 3 func defs
		t.Fatal("Not enough generated functions")
	}

	ast.ApplyBlock(prog.MainFunc.Body.Lines, func(astNode ast.Node) ast.Node {
		switch astNode.(type) {
		case *ast.FunDef:
			t.Fatal("Found function definition in main function")
		}

		return nil
	})
}
