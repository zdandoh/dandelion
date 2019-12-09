package typecheck

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
	"reflect"
	"strings"
)

type TypeVar int

func (t TypeVar) String() string {
	return fmt.Sprintf("t%d", t)
}

func (t TypeVar) ConsString() string {
	return t.String()
}

type Constraint interface {
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
	Subexps     map[TypeVar]ast.Node
	HashToType  map[string]TypeVar
	Constraints map[TypeVar]Constraint
}

func NewTypeInferer() *TypeInferer {
	newInf := &TypeInferer{}
	newInf.Subexps = make(map[TypeVar]ast.Node)
	newInf.HashToType = make(map[string]TypeVar)
	newInf.Constraints = make(map[TypeVar]Constraint)

	return newInf
}

func Infer(prog *ast.Program) {
	infer := NewTypeInferer()

	fmt.Println(prog)
	// Label all subexpressions
	for _, funDef := range prog.Funcs {
		ast.WalkAst(funDef, infer)
	}

	for k, v := range infer.Subexps {
		fmt.Printf("%s\t%s\t%+v\n", k, reflect.TypeOf(v), v)
	}

	infer.CreateConstraints()
}

func (i *TypeInferer) NodeToTypeVar(astNode ast.Node) TypeVar {
	nodeHash := ast.HashNode(astNode)
	typeNo, ok := i.HashToType[nodeHash]
	if !ok {
		i.TypeNo++
		i.HashToType[nodeHash] = i.TypeNo
		typeNo = i.TypeNo
	}

	return typeNo
}

func (i *TypeInferer) WalkNode(astNode ast.Node) ast.Node {
	typeNo := i.NodeToTypeVar(astNode)
	i.Subexps[typeNo] = astNode

	return nil
}

func (i *TypeInferer) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

func (i *TypeInferer) CreateConstraints() {
	for typeVar, astNode := range i.Subexps {
		switch node := astNode.(type) {
		case *ast.Num:
			i.Constraints[typeVar] = BaseType{types.IntType{}}
		case *ast.StrExp:
			i.Constraints[typeVar] = BaseType{types.StringType{}}
		case *ast.AddSub:
			i.Constraints[typeVar] = Same{[]TypeVar{i.NodeToTypeVar(node.Left), i.NodeToTypeVar(node.Right)}}
		case *ast.MulDiv:
			i.Constraints[typeVar] = Same{[]TypeVar{i.NodeToTypeVar(node.Left), i.NodeToTypeVar(node.Right)}}
		case *ast.Assign:
			i.Constraints[i.NodeToTypeVar(node.Target)] = i.NodeToTypeVar(node.Expr)
			i.Constraints[typeVar] = BaseType{types.NullType{}}
		case *ast.FunApp:
			argVars := make([]TypeVar, 0)
			for _, arg := range node.Args {
				argVars = append(argVars, i.NodeToTypeVar(arg))
			}

			i.Constraints[typeVar] = Fun{argVars, i.NodeToTypeVar(node.)}
		case *ast.Ident:
			// Identifiers don't add any additional constraints
		}
	}

	fmt.Println("------ CONSTRAINTS ------")
	for cName, c := range i.Constraints {
		fmt.Printf("%s = %s\n", cName, c.ConsString())
	}
}
