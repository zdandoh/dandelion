package transform

import (
	"dandelion/ast"
	"dandelion/types"
)

type RefFinder struct {
	typeRefs map[types.TypeHash]types.Type
}

func FindTypeRefs(prog *ast.Program) {
	r := &RefFinder{}
	r.typeRefs = make(map[types.TypeHash]types.Type)

	for _, fun := range prog.Funcs {
		ast.WalkAst(fun, r)
	}

	prog.RefTypes = r.typeRefs
}

func (r *RefFinder) WalkNode(astNode ast.Node) ast.Node {
	switch node := astNode.(type) {
	case *ast.IsExp:
		r.typeRefs[types.HashType(node.CheckType)] = node.CheckType
	case *ast.TypeAssert:
		r.typeRefs[types.HashType(node.TargetType)] = node.TargetType
	}

	return nil
}

func (r *RefFinder) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}
