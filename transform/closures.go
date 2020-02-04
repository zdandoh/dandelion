package transform

import (
	"ahead/ast"
	"fmt"
)

type UnboundVars map[string]string

type ClosureExtractor struct {
	FuncUnbounds map[string]UnboundVars
	Prog         *ast.Program
}

type UnboundFinder struct {
	Defs    map[string]bool
	Unbound UnboundVars
}

func NewUnboundFinder(funDef *ast.FunDef, prog *ast.Program) *UnboundFinder {
	f := &UnboundFinder{}
	f.Defs = make(map[string]bool)
	f.Unbound = make(map[string]string)

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
		prog.Funcs[fName] = ast.WalkAst(fun, f).(*ast.FunDef)
		c.FuncUnbounds[fName] = f.Unbound
		fmt.Println(f.Unbound)
	}

	for i, fun := range prog.Funcs {
		prog.Funcs[i] = ast.WalkAst(fun, c).(*ast.FunDef)
	}
}

func (f *UnboundFinder) WalkNode(astNode ast.Node) ast.Node {
	var retVal ast.Node

	switch node := astNode.(type) {
	case *ast.Assign:
		targetIdent, ok := node.Target.(*ast.Ident)
		if !ok {
			// Not assigning to an identifier
			break
		}
		f.Defs[targetIdent.Value] = true
	case *ast.Ident:
		_, ok := f.Defs[node.Value]
		if !ok {
			f.Unbound[node.Value] = node.Value + ".unbound"
		}

		newName, unbound := f.Unbound[node.Value]
		if unbound {
			retVal = &ast.Ident{newName}
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

		// Create closure
		enclosedFunc := c.Prog.Funcs[ident.Value]

		retLines := &ast.LineBundle{}
		cloName := fmt.Sprintf("clo.%s", ident)
		tupleName := fmt.Sprintf("%s.tup", cloName)
		argName := fmt.Sprintf("%s.arg", cloName)

		enclosedFunc.Args = append([]ast.Node{&ast.Ident{argName}}, enclosedFunc.Args...)

		unboundNames := make([]ast.Node, 0)
		i := 0
		for unboundName, newName := range unboundVals {
			unboundNames = append(unboundNames, &ast.Ident{unboundName})
			unboundAssign := &ast.Assign{&ast.Ident{newName}, &ast.SliceNode{&ast.Num{int64(i)}, &ast.Ident{argName}}}
			enclosedFunc.Body.Lines = append([]ast.Node{unboundAssign}, enclosedFunc.Body.Lines...)
			i++
		}
		cloTuple := &ast.TupleLiteral{unboundNames}
		tupAssign := &ast.Assign{&ast.Ident{tupleName}, cloTuple}

		enclosedFunc.Body.Lines = append(enclosedFunc.Body.Lines)

		retLines.Lines = append(retLines.Lines, tupAssign)

		closure := &ast.Closure{}
		closure.Target = ident
		closure.ArgTup = &ast.Ident{tupleName}
		closure.NewFunc = node.Target

		retLines.Lines = append(retLines.Lines, &ast.Assign{node.Target, closure})
		retVal = retLines
	}

	return retVal
}

func (c *ClosureExtractor) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}
