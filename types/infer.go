package types

import "ahead/ast"

type TypeGraph map[ast.Node]*TypeInfNode

type TypeInfNode struct {
	deps     []ast.Node
	node     ast.Node
	nodeType Type
}

type TypeInferer struct {
	Graph TypeGraph
}

func NewTypeInferer() *TypeInferer {
	newInf := &TypeInferer{}
	newInf.Graph = make(TypeGraph)

	return newInf
}

func Infer(prog *ast.Program) TypeGraph {
	infer := NewTypeInferer()
	ast.WalkAst(prog.MainFunc, infer)

	return infer.Graph
}

func (i *TypeInferer) WalkNode(astNode ast.Node) ast.Node {
	switch node := astNode.(type) {
	case *ast.FunApp:
		infNode := &TypeInfNode{}
		infNode.deps = node.Args
		i.Graph[node] = infNode
	}

	return nil
}

func (i *TypeInferer) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

//func ResolveGraph(g TypeGraph) {
//	noDeps := make(TypeGraph)
//	for key, infNode := range noDeps {
//		if len(infNode.deps) == 0 {
//			noDeps[key] =
//		}
//	}
//}
