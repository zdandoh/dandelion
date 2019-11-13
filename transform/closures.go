package transform

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
)

type UnboundVars map[string]bool

type ClosureExtractor struct {
	Prog         *ast.Program
	FuncUnbounds map[string]UnboundVars
	CloCount     int
}

type UnboundFinder struct {
	Defs    map[string]bool
	Unbound UnboundVars
}

func NewUnboundFinder(funDef *ast.FunDef, prog *ast.Program) *UnboundFinder {
	f := &UnboundFinder{}
	f.Defs = make(map[string]bool)
	f.Unbound = make(map[string]bool)

	for name, _ := range prog.Funcs {
		f.Defs[name] = true
	}

	for _, arg := range funDef.Args {
		f.Defs[arg.(*ast.Ident).Value] = true
	}

	return f
}

func ExtractClosures(prog *ast.Program) {
	c := &ClosureExtractor{}
	c.FuncUnbounds = make(map[string]UnboundVars)
	c.Prog = prog

	// Collect unbound names for each function
	for fName, fun := range prog.Funcs {
		f := NewUnboundFinder(fun, prog)
		ast.WalkAst(fun, f)
		c.FuncUnbounds[fName] = f.Unbound
	}

	for i, fun := range prog.Funcs {
		prog.Funcs[i] = ast.WalkAst(fun, c).(*ast.FunDef)
	}
}

func (f *UnboundFinder) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.Assign:
		targetIdent := node.Target.(*ast.Ident).Value
		f.Defs[targetIdent] = true
	case *ast.Ident:
		_, ok := f.Defs[node.Value]
		if !ok {
			f.Unbound[node.Value] = true
		}
	}

	return retVal
}

func (f *UnboundFinder) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

func (c *ClosureExtractor) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.Assign:
		// Literal function definitions are transformed into assignments, so this is where we make our closures
		ident, isExprIdent := node.Expr.(*ast.Ident)
		if !isExprIdent {
			break
		}

		// Check if this identifier is a function name
		unboundVals, isFunc := c.FuncUnbounds[ident.Value]
		if !isFunc || len(unboundVals) == 0 {
			break
		}

		// Create the closure
		closure := &ast.Closure{}
		closure.Target = ident

		unboundNames := make([]ast.Node, 0)
		for unboundName, _ := range unboundVals {
			unboundNames = append(unboundNames, &ast.Ident{unboundName})
		}

		cloContainer := &ast.LineBundle{}

		c.CloCount++
		cloName := fmt.Sprintf("closure.%d", c.CloCount)
		closure.Name = cloName
		cloTuple := &ast.TupleLiteral{unboundNames, types.TupleType{}}
		cloAssign := &ast.Assign{&ast.Ident{cloName}, cloTuple}
		cloContainer.Lines = append(cloContainer.Lines, cloAssign)
		cloContainer.Lines = append(cloContainer.Lines, &ast.Assign{node.Target, closure})

		// Update the function def to take an extra arg and unpack the closure
		funDef := c.Prog.Funcs[ident.Value]
		funDef.Unbound = unboundNames
		funDef.Args = append([]ast.Node{&ast.Ident{cloName}}, funDef.Args...)
		funDef.Type.ArgTypes = append([]types.Type{types.TupleType{}}, funDef.Type.ArgTypes...)

		for i, name := range unboundNames {
			assign := &ast.Assign{name, &ast.SliceNode{&ast.Num{int64(i)}, &ast.Ident{cloName}}}
			funDef.Body.Lines = append([]ast.Node{assign}, funDef.Body.Lines...)
		}

		retVal = cloContainer
	}

	return retVal
}

func (c *ClosureExtractor) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}
