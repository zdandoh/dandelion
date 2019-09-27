package transform

import (
	"ahead/ast"
	"fmt"
)

type Renamer struct {
	NameVersions map[string]int
	LocalNames   map[string]string
}

func (r *Renamer) Clone() *Renamer {
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
	localName = fmt.Sprintf("%s_%d", name, nameNo)
	r.LocalNames[name] = localName

	return localName
}

func RenameIdents(prog *ast.Program) {
	renamer := &Renamer{}
	renamer.NameVersions = make(map[string]int)
	renamer.LocalNames = make(map[string]string)

	prog.MainFunc.Body = ast.WalkBlock(prog.MainFunc.Body, renamer)
}

func (r *Renamer) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.Ident:
		newName := r.getName(node.Value)
		retVal = &ast.Ident{newName}
	case *ast.Assign:
		newName := r.getName(node.Ident)
		retVal = &ast.Assign{newName, ast.WalkAst(node.Expr, r)}
	}

	return retVal
}

func (r *Renamer) WalkBlock(block *ast.Block) *ast.Block {
	renameCopy := r.Clone()
	fmt.Printf("%p\n", renameCopy)
	fmt.Println("Cloned block:", renameCopy, block)

	newLines := make([]ast.Node, 0)
	for _, line := range block.Lines {
		newLines = append(newLines, ast.WalkAst(line, renameCopy))
	}
	return &ast.Block{newLines}
}
