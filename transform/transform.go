package transform

import (
	"ahead/ast"
)

func TransformAst(prog *ast.Program) {
	RenameIdents(prog)
	RemFuncs(prog)
}
