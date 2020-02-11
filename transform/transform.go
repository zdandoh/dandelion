package transform

import (
	"ahead/ast"
)

func TransformAst(prog *ast.Program) {
	RemoveStructs(prog)
	RenameIdents(prog)
	sources := RemFuncs(prog)
	RemovePipes(prog)
	ExtractClosures(prog, sources)
}
