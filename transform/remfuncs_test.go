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

	ts := &struct{ testfuncinside }{}
	ts.walkNode = func(astNode ast.Node) ast.Node {
		switch astNode.(type) {
		case *ast.FunDef:

			t.Fatal("Found function definition in main function")
		}

		return nil
	}
	ts.walkBlock = func(block *ast.Block) *ast.Block {
		return nil
	}

	ast.WalkBlock(prog.MainFunc.Body, ts)
}

type testfuncinside struct {
	walkNode  func(ast.Node) ast.Node
	walkBlock func(*ast.Block) *ast.Block
}

func (t *testfuncinside) WalkBlock(block *ast.Block) *ast.Block {
	return t.walkBlock(block)
}

func (t *testfuncinside) WalkNode(node ast.Node) ast.Node {
	return t.walkNode(node)
}
