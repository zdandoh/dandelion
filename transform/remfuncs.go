package transform

import (
	"dandelion/ast"
	"dandelion/types"
	"fmt"
	"strings"
)

type FunSources map[string][]string

var ThisCount = 0

func (s FunSources) Children(name string) []string {
	return s[name]
}

type FuncRemover struct {
	funcs       map[string]*ast.FunDef
	prog        *ast.Program
	nameStack   *StringStack
	funDefLocs  FunSources
	nameCounter int
}

const FunSuffix = "-imp"

// Remove all inline function definitions from the program and add them to the Funcs.
// Anonymous functions are named fun_<number>
func RemFuncs(prog *ast.Program) FunSources {
	remover := &FuncRemover{}
	remover.prog = prog
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

func (r *FuncRemover) newAnonName() string {
	name := fmt.Sprintf("fun_%d", r.nameCounter)
	r.nameCounter++
	return name
}

func (r *FuncRemover) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.Assign:
		// Whenever a function definition is directly assigned to an identifier, give it that name globally.
		targetIdent, isTargetIdent := node.Target.(*ast.Ident)
		structAccess, isStructAccess := node.Target.(*ast.StructAccess)
		exprFunc, isExprFunc := node.Expr.(*ast.FunDef)

		var newName string
		if isExprFunc && isTargetIdent {
			newName = targetIdent.Value + FunSuffix
			retVal = &ast.Assign{targetIdent, &ast.Ident{newName, targetIdent.NodeID}, node.NodeID}
		} else if isExprFunc && isStructAccess {
			accessTargetIdent, isAccessTargetIdent := structAccess.Target.(*ast.Ident)
			if !isAccessTargetIdent {
				break
			}

			structName := BaseName(accessTargetIdent.Value)

			// TODO future optimization, remove linear runtime
			var foundStruct *ast.StructDef
			for i := 0; i < r.prog.StructCount(); i++ {
				structDef := r.prog.StructNo(i)
				if structDef.Type.Name == structName {
					foundStruct = structDef
					break
				}
			}

			if foundStruct == nil {
				break
			}
			// This is a struct method definition, pull it out to a normal function definition
			// and update the source struct
			methodName := structAccess.Field.(*ast.Ident).Value
			newName = structName + ".method." + methodName
			newMethod := &ast.StructMethod{methodName, newName}
			foundStruct.Methods = append(foundStruct.Methods, newMethod)

			// Rewrite function definition to add 'this' arg & member references to this
			rewriteMethod(exprFunc, foundStruct)

			retVal = &ast.LineBundle{}
		} else {
			break
		}

		r.funDefLocs[r.nameStack.Peek()] = append(r.funDefLocs[r.nameStack.Peek()], newName)
		r.nameStack.Push(newName)

		r.funcs[newName] = exprFunc
		r.funcs[newName].Body = ast.WalkBlock(exprFunc.Body, r)

		r.nameStack.Pop()
	case *ast.FunDef:
		newName := r.newFunName()
		newTarget := "anon-" + r.newAnonName()
		targetIdent := &ast.Ident{newTarget, ast.NoID}
		r.funcs[newName] = node

		beginExp := &ast.BeginExp{[]ast.Node{
			&ast.Assign{targetIdent, &ast.Ident{newName, ast.NoID}, ast.NoID},
			targetIdent,
		}, ast.NoID}
		retVal = beginExp

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

func rewriteMethod(origFun *ast.FunDef, destStruct *ast.StructDef) {
	ThisCount++
	thisName := fmt.Sprintf("__this_%d", ThisCount)
	origFun.Args = append([]ast.Node{&ast.Ident{thisName, ast.NoID}}, origFun.Args...)
	if origFun.TypeHint == nil {
		origFun.TypeHint = &types.FuncType{}
	}
	origFun.TypeHint.ArgTypes = make([]types.Type, len(origFun.Args))
	origFun.TypeHint.ArgTypes[0] = destStruct.Type

	nameChecker := func(name string) bool {
		if destStruct.Has(BaseName(name)) {
			return true
		}
		return false
	}

	nodeGen := func(origNode ast.Node) ast.Node {
		return &ast.StructAccess{&ast.Ident{BaseName(origNode.(*ast.Ident).Value), ast.NoID}, &ast.Ident{thisName, ast.NoID}, ast.NoID}
	}

	for lno, line := range origFun.Body.Lines {
		origFun.Body.Lines[lno] = ReplaceName(line, nameChecker, nodeGen)
	}
}

func TrimFunSuffix(fName string) string {
	return strings.TrimSuffix(fName, FunSuffix)
}
