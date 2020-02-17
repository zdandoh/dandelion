package transform

import "ahead/ast"

type NameReplacer struct {
	nameChecker func(string) bool
	nodeGen     func(origNode ast.Node) ast.Node
}

func ReplaceName(rootNode ast.Node, nameChecker func(string) bool, nodeGen func(ast.Node) ast.Node) ast.Node {
	r := &NameReplacer{}
	r.nameChecker = nameChecker
	r.nodeGen = nodeGen

	return ast.WalkAst(rootNode, r)
}

func (r *NameReplacer) WalkNode(astNode ast.Node) ast.Node {
	switch node := astNode.(type) {
	case *ast.Ident:
		if r.nameChecker(node.Value) {
			return r.nodeGen(node)
		}
	}

	return nil
}

func (r *NameReplacer) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}
