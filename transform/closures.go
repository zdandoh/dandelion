package transform

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
)

type UnboundVars map[string]bool

type ClosureExtractor struct {
	FuncUnbounds map[string]UnboundVars
	Prog *ast.Program
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
		targetIdent, ok := node.Target.(*ast.Ident)
		if !ok {
			// Not assigning to an identifier
			break
		}
		f.Defs[targetIdent.Value] = true
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

		// Create closure
		enclosedFunc := c.Prog.Funcs[ident.Value]
		retLines := &ast.LineBundle{}
		cloName := fmt.Sprintf("closure.%s", ident)
		tupleName := fmt.Sprintf("%s.tup", cloName)

		unboundNames := make([]ast.Node, 0)
		for unboundName, _ := range unboundVals {
			unboundNames = append(unboundNames, &ast.Ident{unboundName})
		}
		cloTuple := &ast.TupleLiteral{unboundNames, types.TupleType{}}
		tupAssign := &ast.Assign{&ast.Ident{tupleName}, cloTuple}

		retLines.Lines = append(retLines.Lines, tupAssign)

		closingArgs := make([]ast.Node, 0)
		for _, arg := range enclosedFunc.Args {
			closingArgs = append(closingArgs, &ast.Ident{arg.(*ast.Ident).Value})
		}
		enclosedFunc.Args = append(enclosedFunc.Args, &ast.Ident{tupleName})

		closingFunc := &ast.FunDef{
			Body: &ast.Block{[]ast.Node{}},
			Args: closingArgs,
			Type: enclosedFunc.Type,
		}
		c.Prog.Funcs[cloName] = closingFunc
		closure := &ast.Closure{}
		closure.Target = ident

		retLines.Lines = append(retLines.Lines, &ast.Assign{node.Target, closure})
		retVal = retLines
	}

	return retVal
}

func (c *ClosureExtractor) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}
