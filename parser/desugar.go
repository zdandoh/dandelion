package parser

import (
	"dandelion/ast"
	"fmt"
)

var iterNo = 0

func ForIterToFor(body *ast.Block, iter ast.Node, itemName ast.Node) ast.Node {
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
	return contBlock
}
