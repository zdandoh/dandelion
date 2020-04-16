package transform

import (
	"dandelion/ast"
	"fmt"
	"strings"
)

const CloArgSuffix = ".arg"
const CloTupSuffix = ".tup"

type UnboundVars map[string]bool
type DefinedVars map[string]bool

type ClosureExtractor struct {
	FuncUnbounds map[string]UnboundVars
	FuncBounds   map[string]DefinedVars
	Prog         *ast.Program
}

type UnboundFinder struct {
	Defs    DefinedVars
	Unbound UnboundVars
}

func NewUnboundFinder(funDef *ast.FunDef, prog *ast.Program) *UnboundFinder {
	f := &UnboundFinder{}
	f.Defs = make(DefinedVars)
	f.Unbound = make(map[string]bool)

	for name, _ := range prog.Funcs {
		f.Defs[name] = true
	}

	for _, arg := range funDef.Args {
		f.Defs[arg.(*ast.Ident).Value] = true
	}

	return f
}

func ExtractClosures(prog *ast.Program, funSources FunSources) {
	c := &ClosureExtractor{}
	c.FuncUnbounds = make(map[string]UnboundVars)
	c.FuncBounds = make(map[string]DefinedVars)
	c.Prog = prog

	// Collect unbound names for each function
	for fName, fun := range prog.Funcs {
		f := NewUnboundFinder(fun, prog)
		prog.Funcs[fName] = ast.WalkAst(fun, f).(*ast.FunDef)
		c.FuncUnbounds[fName] = f.Unbound
		c.FuncBounds[fName] = f.Defs
	}

	c.ResolveUnboundDeps("main", funSources)

	for i, fun := range prog.Funcs {
		prog.Funcs[i] = ast.WalkAst(fun, c).(*ast.FunDef)
	}
}

// Parent functions should have all the unbound variables that their children have,
// as long as those variables aren't bound in the parent. This function makes that happen.
func (c *ClosureExtractor) ResolveUnboundDeps(fName string, funSources FunSources) UnboundVars {
	children := funSources.Children(fName)
	if children != nil {
		for _, child := range children {
			childUnbounds := c.ResolveUnboundDeps(child, funSources)
			for unboundName, _ := range childUnbounds {
				_, isBoundInParent := c.FuncBounds[fName][unboundName]
				if !isBoundInParent {
					c.FuncUnbounds[fName][unboundName] = true
				}
			}
		}
	}

	return c.FuncUnbounds[fName]
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
		cloName := fmt.Sprintf("clo.%s", ident)
		tupName := cloName + CloTupSuffix
		argName := cloName + CloArgSuffix

		structMemberNodes := make([]ast.Node, 0)
		for unboundName, _ := range unboundVals {
			structMemberNodes = append(structMemberNodes, &ast.Ident{unboundName, ast.NoID})
		}

		enclosedFunc.Args = append([]ast.Node{&ast.Ident{argName, ast.NoID}}, enclosedFunc.Args...)

		unboundNames := make([]ast.Node, 0)
		i := 0
		// Rewrite function body to unpack all values
		for unboundName, _ := range unboundVals {
			unboundNames = append(unboundNames, &ast.Ident{unboundName, ast.NoID})
			unboundAssign := &ast.Assign{
				&ast.Ident{unboundName, ast.NoID},
				&ast.SliceNode{
					&ast.Num{int64(i), ast.NoID},
					&ast.Ident{argName, ast.NoID},
					ast.NoID},
				ast.NoID}
			enclosedFunc.Body.Lines = append([]ast.Node{unboundAssign}, enclosedFunc.Body.Lines...)
			i++
		}
		cloTuple := &ast.TupleLiteral{unboundNames, ast.NoID}
		tupIdent := &ast.Ident{tupName, ast.NoID}
		tupAssign := &ast.Assign{tupIdent, cloTuple, ast.NoID}

		enclosedFunc.Body.Lines = append(enclosedFunc.Body.Lines)

		retLines.Lines = append(retLines.Lines, &ast.Assign{node.Target, &ast.NullExp{int(c.Prog.NewNodeID()), ast.NoID}, ast.NoID})
		retLines.Lines = append(retLines.Lines, tupAssign)

		closure := &ast.Closure{}
		closure.Target = ident
		closure.ArgTup = &ast.Ident{tupName, ast.NoID}
		closure.NewFunc = node.Target

		retLines.Lines = append(retLines.Lines, &ast.Assign{node.Target, closure, ast.NoID})

		for i, tupExpr := range cloTuple.Exprs {
			if tupExpr.(*ast.Ident).Value == node.Target.(*ast.Ident).Value {
				retLines.Lines = append(retLines.Lines, &ast.Assign{&ast.SliceNode{&ast.Num{int64(i), ast.NoID}, tupIdent, ast.NoID}, node.Target, ast.NoID})
			}
		}

		retVal = retLines
	}

	return retVal
}

func (c *ClosureExtractor) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

func IsCloArg(node ast.Node) bool {
	ident, isIdent := node.(*ast.Ident)
	if isIdent && strings.HasSuffix(ident.Value, ".arg") {
		return true
	}

	return false
}

func CloArgToTupName(argName string) string {
	baseClo := strings.TrimSuffix(argName, CloArgSuffix)
	return baseClo + CloTupSuffix
}
