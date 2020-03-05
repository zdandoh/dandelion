package transform

import (
	"dandelion/ast"
)

type PipeRemover struct {
}

func RemovePipes(prog *ast.Program) {
	r := &PipeRemover{}

	for i, fun := range prog.Funcs {
		prog.Funcs[i] = ast.WalkAst(fun, r).(*ast.FunDef)
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
