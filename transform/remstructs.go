package transform

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
)

type StructRemover struct {
	prog     *ast.Program
	structNo int
}

func RemoveStructs(prog *ast.Program) {
	remover := &StructRemover{}
	remover.prog = prog

	prog.Funcs["main"].Body = ast.WalkBlock(prog.Funcs["main"].Body, remover)
}

func (r *StructRemover) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.StructDef:
		r.structNo++
		newName := fmt.Sprintf("s_%d", r.structNo)
		r.prog.Structs[newName] = node

		args := make([]ast.Node, len(node.Members))
		argTypes := make([]types.Type, len(node.Members))
		memberNames := make([]string, len(node.Members))
		instanceValues := make([]ast.Node, len(node.Members))
		for i, member := range node.Members {
			args[i] = &ast.Ident{member.Name.Value}
			argTypes[i] = member.Type
			instanceValues[i] = &ast.Ident{member.Name.Value}
			memberNames[i] = member.Name.Value
		}

		constructor := &ast.FunDef{
			Body: &ast.Block{
				[]ast.Node{&ast.StructInstance{
					instanceValues,
					node,
				}},
			},
			Args: args,
			Type: &types.FuncType{argTypes, types.StructType{node.Type.Name, argTypes, memberNames}},
		}
		retVal = constructor
	}

	return retVal
}

func (r *StructRemover) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}
