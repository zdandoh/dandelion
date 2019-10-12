package transform

import (
	"ahead/ast"
)

func TransformAst(prog *ast.Program) {
	RemoveStructs(prog)
	RenameIdents(prog)
	RemFuncs(prog)
}
