package transform

import (
	"ahead/ast"
	"fmt"
	"strings"
)

type FunSources map[string][]string

func (s FunSources) Children(name string) []string {
	return s[name]
}

type FuncRemover struct {
	funcs       map[string]*ast.FunDef
	nameStack   *StringStack
	funDefLocs  FunSources
	nameCounter int
}

const FunSuffix = "-imp"

// Remove all inline function definitions from the program and add them to the Funcs.
// Anonymous functions are named fun_<number>
func RemFuncs(prog *ast.Program) FunSources {
	remover := &FuncRemover{}
	remover.funcs = make(map[string]*ast.FunDef)
	remover.funDefLocs = make(FunSources)
	remover.nameStack = &StringStack{}

	remover.nameStack.Push("main")
	prog.Funcs["main"].Body = ast.WalkBlock(prog.Funcs["main"].Body, remover)
	remover.nameStack.Pop()

	for name, removedFun := range remover.funcs {
		prog.Funcs[name] = removedFun
	}

	return remover.funDefLocs
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
			newName := targetIdent.Value + FunSuffix
			r.funDefLocs[r.nameStack.Peek()] = append(r.funDefLocs[r.nameStack.Peek()], newName)
			r.nameStack.Push(newName)

			r.funcs[targetIdent.Value+FunSuffix] = exprFunc
			r.funcs[targetIdent.Value+FunSuffix].Body = ast.WalkBlock(exprFunc.Body, r)
			retVal = &ast.Assign{targetIdent, &ast.Ident{newName}}

			r.nameStack.Pop()
		}
	case *ast.FunDef:
		newName := r.newFunName()
		r.funcs[newName] = node
		retVal = &ast.Ident{newName}

		r.funDefLocs[r.nameStack.Peek()] = append(r.funDefLocs[r.nameStack.Peek()], newName)
		r.nameStack.Push(newName)
		r.funcs[newName].Body = ast.WalkBlock(node.Body, r)
		r.nameStack.Pop()
	}

	return retVal
}

func (r *FuncRemover) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

func TrimFunSuffix(fName string) string {
	return strings.TrimSuffix(fName, FunSuffix)
}
