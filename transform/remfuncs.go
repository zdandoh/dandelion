package transform

import (
	"ahead/ast"
	"fmt"
	"strings"
)

type FuncRemover struct {
	funcs       map[string]*ast.FunDef
	nameCounter int
}

const FunSuffix = "-imp"

// Remove all inline function definitions from the program and add them to the Funcs.
// Anonymous functions are named fun_<number>
func RemFuncs(prog *ast.Program) {
	remover := &FuncRemover{}
	remover.funcs = make(map[string]*ast.FunDef)

	prog.Funcs["main"].Body = ast.WalkBlock(prog.Funcs["main"].Body, remover)

	for name, removedFun := range remover.funcs {
		prog.Funcs[name] = removedFun
	}
}

func (r *FuncRemover) newFunName() string {
	name := fmt.Sprintf("fun_%d"+FunSuffix, r.nameCounter)
	r.nameCounter++
	return name
}

func (r *FuncRemover) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.Assign:
		// Whenever a function definition is directly assigned to an identifier, give it that name globally.
		targetIdent, isTargetIdent := node.Target.(*ast.Ident)
		exprFunc, isExprFunc := node.Expr.(*ast.FunDef)
		if isExprFunc && isTargetIdent {
			r.funcs[targetIdent.Value+FunSuffix] = exprFunc
			r.funcs[targetIdent.Value+FunSuffix].Body = ast.WalkBlock(exprFunc.Body, r)
			retVal = &ast.Assign{targetIdent, &ast.Ident{targetIdent.Value + FunSuffix}}
		}
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

func TrimFunSuffix(fName string) string {
	return strings.TrimSuffix(fName, FunSuffix)
}
