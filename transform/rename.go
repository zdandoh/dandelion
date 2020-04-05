package transform

import (
	"dandelion/ast"
	"fmt"
	"strings"
)

const NameSep = "_"

type Renamer struct {
	NameVersions map[string]int
	LocalNames   map[string]string
}

func (r *Renamer) LocalCopy() *Renamer {
	newRenamer := &Renamer{}
	newRenamer.LocalNames = make(map[string]string)
	newRenamer.NameVersions = r.NameVersions

	for key, value := range r.LocalNames {
		newRenamer.LocalNames[key] = value
	}

	return newRenamer
}

func (r *Renamer) getName(name string) string {
	localName, exists := r.LocalNames[name]
	if exists {
		return localName
	}

	r.NameVersions[name]++
	nameNo := r.NameVersions[name]
	localName = fmt.Sprintf("%s%s%d", name, NameSep, nameNo)
	r.LocalNames[name] = localName

	return localName
}

func BaseName(name string) string {
	return strings.Split(name, NameSep)[0]
}

func RenameIdents(prog *ast.Program) {
	renamer := &Renamer{}
	renamer.NameVersions = make(map[string]int)
	renamer.LocalNames = make(map[string]string)

	// Setup builtins
	renamer.LocalNames["p"] = "p"
	renamer.LocalNames["abs"] = "abs"

	prog.Funcs["main"].Body = ast.WalkBlock(prog.Funcs["main"].Body, renamer)
}

func (r *Renamer) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.FunDef:
		// We have to do this manually to make args go in the local scope
		renameCopy := r.LocalCopy()
		newArgs := make([]ast.Node, 0)
		for _, arg := range node.Args {
			argIdent := arg.(*ast.Ident)
			argName := argIdent.Value
			renamedArg := renameCopy.getName(argName)
			newArgs = append(newArgs, &ast.Ident{renamedArg, argIdent.NodeID})
		}
		newBlock := renameCopy.WalkBlock(node.Body)
		retVal = &ast.FunDef{newBlock, newArgs, node.TypeHint, node.IsCoro, node.NodeID}
	case *ast.Ident:
		newName := r.getName(node.Value)
		retVal = &ast.Ident{newName, node.NodeID}
	}

	return retVal
}

func (r *Renamer) WalkBlock(block *ast.Block) *ast.Block {
	renameCopy := r.LocalCopy()

	newLines := make([]ast.Node, 0)
	for _, line := range block.Lines {
		newLines = append(newLines, ast.WalkAst(line, renameCopy))
	}
	return &ast.Block{newLines}
}
