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

func DesugarPipeline(pipe *ast.Pipeline, lookupType func(node ast.Node)types.Type) (ast.Node, TypeMap) {
	iterNo++
	typeMap := make(TypeMap)
	dataNode := pipe.Ops[0]

	dataName := &ast.Ident{fmt.Sprintf("pipedata-%d", iterNo), ast.NoID}
	dataSetup := &ast.Assign{dataName, dataNode, ast.NoID}

	retName := &ast.Ident{fmt.Sprintf("piperet-%d", iterNo), ast.NoID}
	emptyArr := &ast.ArrayLiteral{0, []ast.Node{}, -1, ast.NoID}
	retSetup := &ast.Assign{retName, emptyArr, ast.NoID}

	pipeCounter := &ast.Ident{fmt.Sprintf("pipecount-%d", iterNo), ast.NoID}
	counterAssign := &ast.Assign{pipeCounter, &ast.Num{0, ast.NoID}, ast.NoID}
	origStep := &ast.Ident{fmt.Sprintf("pipestep-%d", iterNo), ast.NoID}

	stepIdent := origStep

	lines := make([]ast.Node, 0)
	var lastType types.Type
	var lastStep ast.Node
	for i := 1; i < len(pipe.Ops); i++ {
		args := []ast.Node{stepIdent, pipeCounter, dataName}
		funApp := &ast.FunApp{pipe.Ops[i], args, false, ast.NoID}
		stepIdent = &ast.Ident{fmt.Sprintf("piperes-%d-%d", iterNo, i), ast.NoID}
		lastStep = stepIdent
		stepAssign := &ast.Assign{stepIdent, funApp, ast.NoID}

		stepRetType := lookupType(pipe.Ops[i]).(types.FuncType).RetType
		_, isStepVoid := stepRetType.(types.VoidType)
		lastType = stepRetType
		if isStepVoid {
			lines = append(lines, funApp)
		} else {
			lines = append(lines, stepAssign)
			typeMap[stepIdent] = stepRetType
		}
	}
	counterIncr := &ast.Assign{pipeCounter, &ast.AddSub{pipeCounter, &ast.Num{1, ast.NoID}, "+", ast.NoID}, ast.NoID}
	lines = append(lines, counterIncr)
	_, isLastVoid := lastType.(types.VoidType)
	if !isLastVoid {
		push := &ast.FunApp{
			Fun:&ast.StructAccess{&ast.Ident{"push", ast.NoID}, retName, ast.NoID},
			Args: []ast.Node{lastStep},
			Extern: false,
			NodeID: ast.NoID,
		}
		lines = append(lines, push)
	}

	pipeBody := &ast.Block{lines}
	forIter := &ast.ForIter{origStep, dataName, pipeBody, ast.NoID}

	blockLines := []ast.Node{counterAssign, dataSetup}
	if !isLastVoid {
		blockLines = append(blockLines, retSetup)
	}

	contBlock := &ast.BlockExp{&ast.Block{append(blockLines, forIter)}, ast.NoID}

	typeMap[pipeCounter] = types.IntType{}

	dataType := lookupType(dataNode)
	switch dTy := dataType.(type) {
	case types.ArrayType:
		typeMap[origStep] = dTy.Subtype
	case types.CoroutineType:
		typeMap[origStep] = dTy.Yields
	default:
		panic("Invalid pipe iterator type")
	}

	typeMap[dataName] = dataType
	typeMap[retName] = types.ArrayType{lastType}
	typeMap[emptyArr] = types.ArrayType{lastType}

	var retExp ast.Node
	if isLastVoid {
		retExp = contBlock
	} else {
		retExp = &ast.BeginExp{[]ast.Node{contBlock, retName}, ast.NoID}
	}

	return retExp, typeMap
}
