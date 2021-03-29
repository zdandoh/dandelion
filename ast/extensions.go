package ast

type retFinder struct {
	rets []Node
}

func (f *FunDef) TermExprs() []Node {
	fi := &retFinder{}
	WalkAst(f, fi)

	if len(fi.rets) == 0 && len(f.Body.Lines) > 0 && !Statement(f.Body.Lines[len(f.Body.Lines) - 1]) {
		fi.rets = append(fi.rets, f.Body.Lines[len(f.Body.Lines) - 1])
	}

	return fi.rets
}

func (f *retFinder) WalkNode(astNode Node) Node {
	switch node := astNode.(type) {
	case *ReturnExp:
		f.rets = append(f.rets, node)
	case *YieldExp:
		f.rets = append(f.rets, node)
	}

	return nil
}

func (f *retFinder) WalkBlock(block *Block) *Block {
	return nil
}