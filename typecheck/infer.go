package typecheck

import (
	"ahead/ast"
)

type DepStack struct {
	arr []NodeDeps
	top NodeDeps
}
type NodeDeps []ast.Node

type TypeInferer struct {
	Deps   *DepStack
	FunEnv map[string]*ast.FunDef
}

func (m *DepStack) PushDeps(deps NodeDeps) {
	m.arr = append(m.arr, deps)
	m.top = deps
}

func (m *DepStack) PopEnv() {
	m.arr = m.arr[:len(m.arr)-1]
	m.top = m.arr[len(m.arr)-1]
}

func NewTypeInferer() *TypeInferer {
	newInf := &TypeInferer{}
	newInf.Deps = &DepStack{}

	return newInf
}

func Infer(prog *ast.Program) {
	infer := NewTypeInferer()
	ast.WalkAst(prog.Funcs["main"], infer)
}

func (i *TypeInferer) WalkNode(astNode ast.Node) ast.Node {
	switch node := astNode.(type) {
	case *ast.Ident:
		i.Deps.top = append(i.Deps.top, node)
	case *ast.FunApp:

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
