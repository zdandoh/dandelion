package typecheck

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
	"strings"
)

type TypeVar int

func (t TypeVar) String() string {
	return fmt.Sprintf("t%d", t)
}

func (t TypeVar) ConsString() string {
	return t.String()
}

type Constraint struct {
	Left  TypeVar
	Right Constrainable
}

type Constrainable interface {
	ConsString() string
}

type Same struct {
	Types []TypeVar
}

func (t Same) ConsString() string {
	varStrings := make([]string, 0)
	for _, tVar := range t.Types {
		varStrings = append(varStrings, tVar.ConsString())
	}

	return fmt.Sprintf("same(%s)", strings.Join(varStrings, ", "))
}

type BaseType struct {
	types.Type
}

func (t BaseType) ConsString() string {
	return t.Type.TypeString()
}

type Fun struct {
	Args []TypeVar
	Ret  TypeVar
}

func (t Fun) ConsString() string {
	varStrings := make([]string, 0)
	for _, tVar := range t.Args {
		varStrings = append(varStrings, tVar.ConsString())
	}

	return fmt.Sprintf("(%s -> %s)", strings.Join(varStrings, ", "), t.Ret.ConsString())
}

type TypeInferer struct {
	TypeNo      TypeVar
	Subexps     []ast.Node
	HashToType  map[string]TypeVar
	Constraints []Constraint
}

func NewTypeInferer() *TypeInferer {
	newInf := &TypeInferer{}
	newInf.Subexps = make([]ast.Node, 0)
	newInf.HashToType = make(map[string]TypeVar)
	newInf.Constraints = make([]Constraint, 0)

	return newInf
}

func Infer(prog *ast.Program) {
	infer := NewTypeInferer()

	fmt.Println(prog)
	// Label all subexpressions
	for _, funDef := range prog.Funcs {
		ast.WalkAst(funDef, infer)
	}

	//for funVarName, funDef := range prog.Funcs {
	//	fmt.Println(funVarName)
	//	infer.AddCons(Constraint{infer.GetTypeVar(&ast.Ident{funVarName}), infer.GetTypeVar(funDef)})
	//}

	infer.CreateConstraints()
	UnifyConstraints(infer.Constraints)
}

func (i *TypeInferer) AddCons(con Constraint) {
	i.Constraints = append(i.Constraints, con)
}

func (i *TypeInferer) NodeToTypeVar(astNode ast.Node) (TypeVar, bool) {
	nodeHash := ast.HashNode(astNode)
	typeNo, ok := i.HashToType[nodeHash]
	if !ok {
		i.TypeNo++
		i.HashToType[nodeHash] = i.TypeNo
		typeNo = i.TypeNo
	}

	return typeNo, ok
}

func (i *TypeInferer) GetTypeVar(astNode ast.Node) TypeVar {
	tVar, _ := i.NodeToTypeVar(astNode)
	return tVar
}

func (i *TypeInferer) WalkNode(astNode ast.Node) ast.Node {
	return nil
}

func (i *TypeInferer) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

func (i *TypeInferer) PostWalk(astNode ast.Node) {
	_, existed := i.NodeToTypeVar(astNode)
	if !existed {
		i.Subexps = append(i.Subexps, astNode)
	}
}

func (i *TypeInferer) CreateConstraints() {
	fmt.Println("--- ORDERED SUBEXPS ---")
	for _, astNode := range i.Subexps {
		typeVar, _ := i.NodeToTypeVar(astNode)
		fmt.Println(typeVar, "-", astNode)
	}

	for _, astNode := range i.Subexps {
		typeVar, _ := i.NodeToTypeVar(astNode)

		switch node := astNode.(type) {
		case *ast.Num:
			i.AddCons(Constraint{typeVar, BaseType{types.IntType{}}})
		case *ast.StrExp:
			i.AddCons(Constraint{typeVar, BaseType{types.StringType{}}})
		case *ast.AddSub:
			i.AddCons(Constraint{i.GetTypeVar(node.Left), i.GetTypeVar(node.Right)})
			i.AddCons(Constraint{typeVar, i.GetTypeVar(node.Right)})
		case *ast.MulDiv:
			i.AddCons(Constraint{i.GetTypeVar(node.Left), i.GetTypeVar(node.Right)})
		case *ast.Assign:
			i.AddCons(Constraint{i.GetTypeVar(node.Target), i.GetTypeVar(node.Expr)})
		case *ast.FunApp:
			argVars := make([]TypeVar, 0)
			for _, arg := range node.Args {
				argVars = append(argVars, i.GetTypeVar(arg))
			}

			i.AddCons(Constraint{i.GetTypeVar(node.Fun), Fun{argVars, typeVar}})
		case *ast.Ident:
			// Identifiers don't add any additional constraints
		}
	}

	fmt.Println("------ CONSTRAINTS ------")
	for _, c := range i.Constraints {
		fmt.Printf("%s = %s\n", c.Left, c.Right.ConsString())
	}
}
