package transform

import (
	"dandelion/ast"
	"fmt"
)

type PipeRemover struct {
}

func RemovePipes(prog *ast.Program) {
	r := &PipeRemover{}
	d := &PipeDesugar{}

	for i, fun := range prog.Funcs {
		prog.Funcs[i] = ast.WalkAst(fun, r).(*ast.FunDef)
		prog.Funcs[i] = ast.WalkAst(prog.Funcs[i], d).(*ast.FunDef)
	}
}

func (r *PipeRemover) WalkNode(astNode ast.Node) ast.Node {
	var retNode ast.Node

	switch node := astNode.(type) {
	case *ast.PipeExp:
		// Walk down the pipeline an collect all the operations
		newPipeline := &ast.Pipeline{}
		currPipe := node
		for {
			newPipeline.Ops = append([]ast.Node{currPipe.Right}, newPipeline.Ops...)

			leftOp, isPipe := currPipe.Left.(*ast.PipeExp)
			if isPipe {
				currPipe = leftOp
			} else {
				newPipeline.Ops = append([]ast.Node{currPipe.Left}, newPipeline.Ops...)
				break
			}
		}

		// Check for pipelines that might be nested
		for i, op := range newPipeline.Ops {
			newPipeline.Ops[i] = ast.WalkAst(op, r)
		}

		retNode = newPipeline
	}

	return retNode
}

func (r *PipeRemover) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

type PipeDesugar struct {
	pipeCount int
}

func (d *PipeDesugar) WalkNode(astNode ast.Node) ast.Node {
	pipe, isPipe := astNode.(*ast.Pipeline)
	if isPipe {
		return d.DesugarPipeline(pipe)
	}

	return nil
}

func (d *PipeDesugar) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

func (d *PipeDesugar) DesugarPipeline(pipe *ast.Pipeline) ast.Node {
	d.pipeCount++
	dataNode := pipe.Ops[0]

	pipeCounter := &ast.Ident{fmt.Sprintf("pipecount-%d", d.pipeCount), ast.NoID}
	counterAssign := &ast.Assign{pipeCounter, &ast.Num{0, ast.NoID}, ast.NoID}
	origStep := &ast.Ident{fmt.Sprintf("pipestep-%d", d.pipeCount), ast.NoID}
	stepIdent := origStep

	lines := make([]ast.Node, 0)
	for i := 1; i < len(pipe.Ops); i++ {
		args := []ast.Node{stepIdent, pipeCounter, dataNode}
		funApp := &ast.FunApp{pipe.Ops[i], args, false, ast.NoID}
		stepIdent = &ast.Ident{fmt.Sprintf("piperes-%d-%d", d.pipeCount, i), ast.NoID}
		stepAssign := &ast.Assign{stepIdent, funApp, ast.NoID}
		lines = append(lines, stepAssign)
	}
	lines = append(lines, &ast.Assign{pipeCounter, &ast.AddSub{pipeCounter, &ast.Num{1, ast.NoID}, "+", ast.NoID}, ast.NoID})

	pipeBody := &ast.Block{lines}
	forIter := &ast.ForIter{origStep, dataNode, pipeBody, ast.NoID}
	contBlock := &ast.BlockExp{&ast.Block{[]ast.Node{counterAssign, forIter}}, ast.NoID}

	return contBlock
}
