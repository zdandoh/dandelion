package transform

import "dandelion/ast"

func MarkCoroutines(prog *ast.Program) {
	for _, funDef := range prog.Funcs {
		yieldFinder := &FindYield{}
		ast.WalkAst(funDef, yieldFinder)
		funDef.IsCoro = &yieldFinder.hasYield
	}
}

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
