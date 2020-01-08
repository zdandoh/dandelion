package typecheck

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
	"strings"
)

type Constrainable interface {
	ConsString() string
}

type TypeVar int

func (t TypeVar) String() string {
	return fmt.Sprintf("t%d", t)
}

func (t TypeVar) ConsString() string {
	return t.String()
}

type Constraint struct {
	Left  Constrainable
	Right Constrainable
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
	TypeToNode  map[TypeVar]ast.Node
	Constraints []Constraint
	FunLookup   map[string]Fun
}

func NewTypeInferer() *TypeInferer {
	newInf := &TypeInferer{}
	newInf.Subexps = make([]ast.Node, 0)
	newInf.HashToType = make(map[string]TypeVar)
	newInf.Constraints = make([]Constraint, 0)
	newInf.FunLookup = make(map[string]Fun)
	newInf.TypeToNode = make(map[TypeVar]ast.Node)

	return newInf
}

func Infer(prog *ast.Program) {
	infer := NewTypeInferer()

	fmt.Println(prog)
	// Setup all function defs
	for fName, funDef := range prog.Funcs {
		funCons := Fun{}
		for _, arg := range funDef.Args {
			funCons.Args = append(funCons.Args, infer.GetTypeVar(arg))
		}

		retVar := infer.NewTypeVar()
		funCons.Ret = retVar

		infer.FunLookup[fName] = funCons
	}

	// Collect all unique subexpressions
	for _, funDef := range prog.Funcs {
		ast.WalkAst(funDef, infer)
	}

	fmt.Println(infer.FunLookup)
	infer.CreateConstraints(prog)
	subs := Unify(infer.Constraints, make(map[Constrainable]Constrainable), 0)
	infer.ConstructTypes(subs)
}

func (i *TypeInferer) ConstructTypes(subs Subs) {
	fmt.Println("--- FINAL TYPES ---")
	for _, subExp := range i.Subexps {
		initialVar := i.GetTypeVar(subExp)
		resolvedType := i.ResolveType(initialVar, subs)

		fmt.Println(subExp, "-", resolvedType.TypeString())
	}
}

func (i *TypeInferer) ResolveType(typeVar TypeVar, subs Subs) types.Type {
	finalVar := LookupVar(typeVar, subs)
	baseType, isBaseType := finalVar.(BaseType)
	if isBaseType {
		return baseType.Type
	}

	finalTVar := finalVar.(TypeVar)
	sourceNode := i.TypeToNode[finalTVar]

	switch node := sourceNode.(type) {
	case *ast.FunDef:
		funType := types.FuncType{}
		for _, arg := range node.Args {
			argType := i.ResolveType(i.GetTypeVar(arg), subs)
			funType.ArgTypes = append(funType.ArgTypes, argType)
		}
		funType.RetType = i.ResolveType(i.GetTypeVar(node.Body.Lines[len(node.Body.Lines)-1]), subs)
		return funType
	}

	return types.NullType{}
}

func (i *TypeInferer) AddCons(con Constraint) {
	i.Constraints = append(i.Constraints, con)
}

func (i *TypeInferer) NewTypeVar() TypeVar {
	i.TypeNo++
	return i.TypeNo
}

func (i *TypeInferer) NodeToTypeVar(astNode ast.Node) (TypeVar, bool) {
	nodeHash := ast.HashNode(astNode)
	typeNo, ok := i.HashToType[nodeHash]
	if !ok {
		i.TypeNo++
		i.HashToType[nodeHash] = i.TypeNo
		i.TypeToNode[i.TypeNo] = astNode
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

func (i *TypeInferer) CreateConstraints(prog *ast.Program) {
	fmt.Println("--- ORDERED SUBEXPS ---")
	for _, astNode := range i.Subexps {
		typeVar, _ := i.NodeToTypeVar(astNode)
		fmt.Println(typeVar, "-", astNode)
	}

	for fName, funDef := range prog.Funcs {
		baseFun := i.FunLookup[fName]
		if fName == "main" {
			i.AddCons(Constraint{baseFun.Ret, BaseType{types.IntType{}}})
		} else {
			fName := &ast.Ident{fName}
			i.AddCons(Constraint{i.GetTypeVar(fName), i.GetTypeVar(funDef)})                              // Constraint to assign variable name to actual function
			i.AddCons(Constraint{baseFun.Ret, i.GetTypeVar(funDef.Body.Lines[len(funDef.Body.Lines)-1])}) // Setup return constraint for function
		}
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
			baseFun := i.FunLookup[node.Fun.(*ast.Ident).Value]
			fmt.Println("BASE", baseFun)
			newFun := Fun{}
			for k, arg := range node.Args {
				i.AddCons(Constraint{i.GetTypeVar(arg), baseFun.Args[k]})
				newFun.Args = append(newFun.Args, i.GetTypeVar(arg))
			}
			i.AddCons(Constraint{i.GetTypeVar(node), baseFun.Ret})
			newFun.Ret = i.GetTypeVar(node)

			fmt.Println(baseFun.ConsString(), "=", newFun.ConsString())
		case *ast.Ident:
			// Identifiers don't add any additional constraints
		}
	}

	fmt.Println("------ CONSTRAINTS ------")
	for _, c := range i.Constraints {
		fmt.Printf("%s = %s\n", c.Left, c.Right.ConsString())
	}
}
