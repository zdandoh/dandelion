package typecheck

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
	"strings"
)

const DebugTypeInf = true

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

type Options struct {
	Types []types.Type
}

func (t Options) ConsString() string {
	typeStrings := make([]string, 0)
	for _, t := range t.Types {
		typeStrings = append(typeStrings, t.TypeString())
	}

	return fmt.Sprintf("option[%s]", strings.Join(typeStrings, ", "))
}

var Addable = Options{[]types.Type{types.FloatType{}, types.IntType{}, types.StringType{}}}

type Container struct {
	Type    types.Type
	Subtype Constrainable
	Index   int
}

func (c Container) ConsString() string {
	return fmt.Sprintf("container<%v>[%v]", c.Type.TypeString(), c.Subtype.ConsString())
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

func DebugInfer(more ...interface{}) {
	if DebugTypeInf {
		fmt.Println(more...)
	}
}

type TypeInferer struct {
	TypeNo       TypeVar
	Subexps      []ast.Node
	Subtypes     map[TypeVar]TypeVar
	HashToType   map[ast.NodeHash]TypeVar
	TypeToNode   map[TypeVar]ast.Node
	Constraints  []Constraint
	FunLookup    map[string]Fun
	FunDefLookup map[*ast.FunDef]Fun
}

func NewTypeInferer() *TypeInferer {
	newInf := &TypeInferer{}
	newInf.Subexps = make([]ast.Node, 0)
	newInf.HashToType = make(map[ast.NodeHash]TypeVar)
	newInf.Constraints = make([]Constraint, 0)
	newInf.FunLookup = make(map[string]Fun)
	newInf.FunDefLookup = make(map[*ast.FunDef]Fun)
	newInf.TypeToNode = make(map[TypeVar]ast.Node)
	newInf.Subtypes = make(map[TypeVar]TypeVar)

	return newInf
}

func Infer(prog *ast.Program) map[ast.NodeHash]types.Type {
	infer := NewTypeInferer()

	DebugInfer("--- Program ast before inference ---")
	DebugInfer(prog)
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
		infer.FunDefLookup[funDef] = funCons
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
	DebugInfer("--- FINAL TYPES ---")

	for _, subExp := range i.Subexps {
		initialVar := i.GetTypeVar(subExp)
		resolvedType := i.ResolveType(initialVar, subs)

		finalTypes[ast.HashNode(subExp)] = resolvedType
		DebugInfer(subExp, "-", resolvedType.TypeString())
	}

	for fName, _ := range i.FunLookup {
		fIdent := &ast.Ident{fName}
		initialVar := i.GetTypeVar(&ast.Ident{fName})
		resolvedType := i.ResolveType(initialVar, subs)

		finalTypes[ast.HashNode(fIdent)] = resolvedType
		DebugInfer(fName, "-", resolvedType.TypeString())
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
		funType.RetType = i.ResolveType(i.FunDefLookup[node].Ret, subs)
		return funType
	case *ast.ArrayLiteral:
		listType := types.ArrayType{}
		subTypevar := i.GetSubtype(node)
		subType := i.ResolveType(subTypevar, subs)
		listType.Subtype = subType
		return listType
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

func (i *TypeInferer) GetSubtype(astNode ast.Node) TypeVar {
	typeVar := i.GetTypeVar(astNode)
	subtype, ok := i.Subtypes[typeVar]
	if !ok {
		subtype = i.NewTypeVar()
		i.Subtypes[typeVar] = subtype
		fmt.Println("Created subtype", subtype)
	}

	return subtype
}

func (i *TypeInferer) CreateConstraints(prog *ast.Program) {
	DebugInfer("--- ORDERED SUBEXPS ---")
	for _, astNode := range i.Subexps {
		typeVar, _ := i.NodeToTypeVar(astNode)
		DebugInfer(typeVar, "-", astNode)
	}

	// Create constraints for all globally defined functions
	for fName, funDef := range prog.Funcs {
		fIdent := &ast.Ident{fName}
		baseFun := i.FunLookup[fName]
		i.AddCons(Constraint{i.GetTypeVar(fIdent), i.GetTypeVar(funDef)}) // Constraint to assign variable name to actual function

		if funDef.TypeHint != nil {
			// The user provided the type of the function
			i.AddCons(Constraint{i.GetTypeVar(fIdent), BaseType{*funDef.TypeHint}})
			for argNo, arg := range funDef.Args {
				i.AddCons(Constraint{i.GetTypeVar(arg), BaseType{funDef.TypeHint.ArgTypes[argNo]}})
			}
			i.AddCons(Constraint{i.GetTypeVar(baseFun.Ret), BaseType{funDef.TypeHint.RetType}})
		}

		// Check if function has a non-return last line. If so, setup inference for that line.
		var lastLine ast.Node
		if len(funDef.Body.Lines) > 0 {
			lastLine = funDef.Body.Lines[len(funDef.Body.Lines)-1]
		}

		_, isReturn := lastLine.(*ast.ReturnExp)
		if lastLine != nil && !isReturn {
			i.AddCons(Constraint{baseFun.Ret, i.GetTypeVar(lastLine)})
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
			i.AddCons(Constraint{typeVar, i.GetTypeVar(node.Right)})
		case *ast.ParenExp:
			i.AddCons(Constraint{i.GetTypeVar(node), i.GetTypeVar(node.Exp)})
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
		case *ast.ArrayLiteral:
			subtypeVar := i.GetSubtype(node)
			i.AddCons(Constraint{typeVar, Container{types.ArrayType{types.NullType{}}, subtypeVar, 0}})
			if len(node.Exprs) > 0 {
				i.AddCons(Constraint{subtypeVar, i.GetTypeVar(node.Exprs[0])})
			}
		case *ast.SliceNode:
			subTypevar := i.GetSubtype(i.GetTypeVar(node))
			i.AddCons(Constraint{typeVar, subTypevar})
		}
	}

	DebugInfer("------ CONSTRAINTS ------")
	for _, c := range i.Constraints {
		DebugInfer(c.Left, "=", c.Right.ConsString())
	}
}
