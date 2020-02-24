package transform

import "ahead/ast"

type FindYield struct {
	hasYield bool
}

func (y *FindYield) WalkNode(astNode ast.Node) ast.Node {
	_, isYield := astNode.(*ast.YieldExp)
	if isYield {
		y.hasYield = true
	}

	return nil
}

func (y *FindYield) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

func HasYield(node ast.Node) bool {
	finder := &FindYield{}
	ast.WalkAst(node, finder)
	return finder.hasYield
}
