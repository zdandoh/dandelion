package transform

import (
	"ahead/ast"
	"fmt"
)

type FuncRemover struct {
	funcs       map[string]*ast.FunDef
	nameCounter int
}

// Remove all inline function definitions from the program and add them to the Funcs.
// Generated functions are named fun_<number>
func RemFuncs(prog *ast.Program) {
	remover := &FuncRemover{}
	remover.funcs = make(map[string]*ast.FunDef)

	prog.MainFunc.Body = ast.WalkBlock(prog.MainFunc.Body, remover)
	prog.Funcs = remover.funcs
}

func (r *FuncRemover) newFunName() string {
	name := fmt.Sprintf("fun_%d", r.nameCounter)
	r.nameCounter++
	return name
}

func (r *FuncRemover) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.FunDef:
		newName := r.newFunName()
		r.funcs[newName] = node
		retVal = &ast.Ident{newName}
		r.funcs[newName].Body = ast.WalkBlock(node.Body, r)
	}

	return retVal
}

func (r *FuncRemover) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

