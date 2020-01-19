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

//func (c Container) ConsString() string {
//	subStrs := make([]string, 0)
//	for _, sub := range c.Subtypes {
//		subStrs = append(subStrs, sub.ConsString())
//	}
//	return fmt.Sprintf("(%s)", strings.Join(subStrs, ", "))
//}

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
	HashToType  map[ast.NodeHash]TypeVar
	TypeToNode  map[TypeVar]ast.Node
	Constraints []Constraint
	FunLookup   map[string]Fun
}

func NewTypeInferer() *TypeInferer {
	newInf := &TypeInferer{}
	newInf.Subexps = make([]ast.Node, 0)
	newInf.HashToType = make(map[ast.NodeHash]TypeVar)
	newInf.Constraints = make([]Constraint, 0)
	newInf.FunLookup = make(map[string]Fun)
	newInf.TypeToNode = make(map[TypeVar]ast.Node)

	return newInf
}

func Infer(prog *ast.Program) map[ast.NodeHash]types.Type {
	infer := NewTypeInferer()

	fmt.Println(prog)
	// Setup all function defs
	for fName, funDef := range prog.Funcs {
		funCons := Fun{}
		for _, arg := range funDef.Args {
			infer.Subexps = append(infer.Subexps, arg)
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

	infer.CreateConstraints(prog)
	subs := Unify(infer.Constraints, make(map[Constrainable]Constrainable), 0)
	return infer.ConstructTypes(subs)
}

func (i *TypeInferer) ConstructTypes(subs Subs) map[ast.NodeHash]types.Type {
	finalTypes := make(map[ast.NodeHash]types.Type)
	fmt.Println("--- FINAL TYPES ---")

	for _, subExp := range i.Subexps {
		initialVar := i.GetTypeVar(subExp)
		resolvedType := i.ResolveType(initialVar, subs)

		finalTypes[ast.HashNode(subExp)] = resolvedType
		fmt.Println(subExp, "-", resolvedType.TypeString())
	}

	for fName, _ := range i.FunLookup {
		fIdent := &ast.Ident{fName}
		initialVar := i.GetTypeVar(&ast.Ident{fName})
		resolvedType := i.ResolveType(initialVar, subs)

		finalTypes[ast.HashNode(fIdent)] = resolvedType
		fmt.Println(fName, "-", resolvedType.TypeString())
	}

	return finalTypes
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

func EmptyList(node ast.Node) bool {
	list, isList := node.(*ast.ArrayLiteral)
	return isList && len(list.Exprs) == 0
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
		fIdent := &ast.Ident{fName}
		i.AddCons(Constraint{i.GetTypeVar(fIdent), i.GetTypeVar(funDef)}) // Constraint to assign variable name to actual function

		if funDef.TypeHint != nil {
			// The user provided the type of the function
			i.AddCons(Constraint{i.GetTypeVar(fIdent), BaseType{*funDef.TypeHint}})
			continue
		}

		baseFun := i.FunLookup[fName]
		if fName == "main" {
			i.AddCons(Constraint{baseFun.Ret, BaseType{types.IntType{}}})
		} else {
			// TODO make this work for empty functions
			i.AddCons(Constraint{baseFun.Ret, i.GetTypeVar(funDef.Body.Lines[len(funDef.Body.Lines)-1])}) // Setup return constraint for function
		}
	}

	for _, astNode := range i.Subexps {
		typeVar, _ := i.NodeToTypeVar(astNode)

		switch node := astNode.(type) {
		case *ast.Num:
			i.AddCons(Constraint{typeVar, BaseType{types.IntType{}}})
		case *ast.FloatExp:
			i.AddCons(Constraint{typeVar, BaseType{types.FloatType{}}})
		case *ast.StrExp:
			i.AddCons(Constraint{typeVar, BaseType{types.StringType{}}})
		case *ast.ByteExp:
			i.AddCons(Constraint{typeVar, BaseType{types.ByteType{}}})
		case *ast.BoolExp:
			i.AddCons(Constraint{typeVar, BaseType{types.BoolType{}}})
		case *ast.AddSub:
			i.AddCons(Constraint{i.GetTypeVar(node.Left), i.GetTypeVar(node.Right)})
			i.AddCons(Constraint{typeVar, i.GetTypeVar(node.Right)})
		case *ast.MulDiv:
			i.AddCons(Constraint{i.GetTypeVar(node.Left), i.GetTypeVar(node.Right)})
		case *ast.Assign:
			if EmptyList(node.Expr) {
				// This is a special case, the empty list is the only expression that does not have a distinct type,
				// so we cannot create a standard constraint in this case.
			} else {
				i.AddCons(Constraint{i.GetTypeVar(node.Target), i.GetTypeVar(node.Expr)})
			}
		case *ast.FunApp:
			baseFun, ok := i.FunLookup[node.Fun.(*ast.Ident).Value]
			if !ok {
				// Not a 'global' level function definition. We can't use an application for inference
				break
			}

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
		case *ast.CompNode:
			i.AddCons(Constraint{i.GetTypeVar(node.Left), i.GetTypeVar(node.Right)})
		//case *ast.ArrayLiteral:
		//	if len(node.Exprs) > 0 {
		//		i.AddCons(Constraint{typeVar, Container{BaseType{types.ArrayType{}}, []Constrainable{i.GetTypeVar(node.Exprs[0])}}})
		//	}
		case *ast.SliceNode:
			i.AddCons(Constraint{i.GetTypeVar(node.Arr), BaseType{types.ArrayType{}}})
		}
	}

	fmt.Println("------ CONSTRAINTS ------")
	for _, c := range i.Constraints {
		fmt.Printf("%s = %s\n", c.Left, c.Right.ConsString())
	}
}
