package parser

import (
	"dandelion/ast"
	"dandelion/types"
	"fmt"
)

var iterNo = 0

type TypeMap map[ast.Node]types.Type

func DesugarForIter(body *ast.Block, iter ast.Node, itemName ast.Node, iterType types.Type) (ast.Node, TypeMap) {
	var retNode ast.Node
	var typeMap TypeMap

	switch iterType.(type) {
	case types.CoroutineType:
		retNode, typeMap = DesugarForIterCoro(body, iter, itemName, iterType)
	case types.ArrayType:
		retNode, typeMap = DesugarForIterArr(body, iter, itemName, iterType)
	default:
		panic("Invalid iter type for iterator-style for loop")
	}

	return retNode, typeMap
}

func DesugarForIterCoro(body *ast.Block, iter ast.Node, itemName ast.Node, iterType types.Type) (ast.Node, TypeMap) {
	iterNo++
	iterName := fmt.Sprintf("iter-%d", iterNo)

	iterIdent := &ast.Ident{iterName, ast.NoID}
	iterAssign := &ast.Assign{iterIdent, iter, ast.NoID}
	itemAssign := &ast.Assign{itemName, &ast.BuiltinExp{[]ast.Node{iterIdent}, ast.BuiltinNext, ast.NoID}, ast.NoID}
	doneCheck := &ast.BuiltinExp{[]ast.Node{iterIdent}, ast.BuiltinDone, ast.NoID}

	forItem := &ast.For{
		&ast.Num{0, ast.NoID},
		&ast.CompNode{"==", doneCheck, &ast.BoolExp{false, ast.NoID}, ast.NoID},
		itemAssign,
		body,
		ast.NoID}
	contBlock := &ast.BlockExp{&ast.Block{[]ast.Node{iterAssign, itemAssign, forItem}}, ast.NoID}

	typeMap := make(TypeMap)
	typeMap[iterIdent] = iterType

	return contBlock, typeMap
}

func DesugarForIterArr(body *ast.Block, arr ast.Node, itemName ast.Node, iterType types.Type) (ast.Node, TypeMap) {
	iterNo++
	iterName := fmt.Sprintf("iter-%d", iterNo)
	arrIdent := &ast.Ident{fmt.Sprintf("arr-%d", iterNo), ast.NoID}
	arrAss := &ast.Assign{arrIdent, arr, ast.NoID}
	iterIdent := &ast.Ident{iterName, ast.NoID}
	init := &ast.Assign{iterIdent, &ast.Num{0, ast.NoID}, ast.NoID}
	cond := &ast.CompNode{"<", iterIdent, &ast.BuiltinExp{[]ast.Node{arrIdent}, ast.BuiltinLen, ast.NoID}, ast.NoID}
	incr := &ast.Assign{iterIdent, &ast.AddSub{iterIdent, &ast.Num{1, ast.NoID}, "+", ast.NoID}, ast.NoID}
	forLoop := &ast.For{init, cond, incr, body, ast.NoID}
	itemAssign := &ast.Assign{itemName, &ast.SliceNode{iterIdent, arrIdent, ast.NoID}, ast.NoID}
	body.Lines = append([]ast.Node{itemAssign}, body.Lines...)
	contBlock := &ast.BlockExp{&ast.Block{[]ast.Node{arrAss, forLoop}}, ast.NoID}

	typeMap := make(TypeMap)
	typeMap[arrIdent] = iterType
	typeMap[iterIdent] = types.IntType{}

	return contBlock, typeMap
}
