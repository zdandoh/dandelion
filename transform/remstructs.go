package transform

import (
	"ahead/ast"
	"fmt"
)

type StructRemover struct {
	prog     *ast.Program
	structNo int
}

func RemoveStructs(prog *ast.Program) {
	remover := &StructRemover{}
	remover.prog = prog

	prog.MainFunc.Body = ast.WalkBlock(prog.MainFunc.Body, remover)
}

func (r *StructRemover) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.StructDef:
		r.structNo++
		newName := fmt.Sprintf("s_%d", r.structNo)
		r.prog.Structs[newName] = node

		args := make([]ast.Node, 0)
		defaultValues := make(map[*ast.Ident]ast.Node)
		for _, member := range node.Members {
			args = append(args, &ast.Ident{member.Name.Value})
			defaultValues[member.Name] = &ast.Ident{member.Name.Value}
		}

		constructor := &ast.FunDef{
			Body: &ast.Block{
				[]ast.Node{&ast.StructInstance{
					newName,
					defaultValues,
				}},
			},
			Args: args,
		}
		retVal = constructor
	}

	return retVal
}

func (r *StructRemover) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}
