package transform

import (
	"dandelion/ast"
)

func TransformAst(prog *ast.Program) {
	RemoveStructs(prog)
	RenameIdents(prog)
	sources := RemFuncs(prog)
	MarkCoroutines(prog)
	RemovePipes(prog)
	ExtractClosures(prog, sources)
}
