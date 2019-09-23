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

	prog.MainFunc.Body.Lines = ast.ApplyBlock(prog.MainFunc.Body.Lines, remover.remExprFuncs)
	prog.Funcs = remover.funcs
}

func (r *FuncRemover) newFunName() string {
	name := fmt.Sprintf("fun_%d", r.nameCounter)
	r.nameCounter++
	return name
}

func (r *FuncRemover) remExprFuncs(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.FunDef:
		newName := r.newFunName()
		r.funcs[newName] = node
		retVal = &ast.Ident{newName}
		r.funcs[newName].Body.Lines = ast.ApplyBlock(node.Body.Lines, r.remExprFuncs)
	}

	return retVal
}
