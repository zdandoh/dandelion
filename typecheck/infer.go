package typecheck

import (
	"ahead/ast"
	"ahead/types"
	"fmt"
	"os"
	"reflect"
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

func (c Constraint) String() string {
	return fmt.Sprintf("%s = %s", c.Left.ConsString(), c.Right.ConsString())
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

type StructOptions struct {
	Types      []types.Type
	Dependants map[TypeVar]string
}

func (o StructOptions) ConsString() string {
	typeStrings := make([]string, 0)
	depStrings := make([]string, 0)

	for _, t := range o.Types {
		typeStrings = append(typeStrings, t.TypeString())
	}
	for k, v := range o.Dependants {
		depStrings = append(depStrings, fmt.Sprintf("%v: %v", k, v))
	}

	return fmt.Sprintf("struct-options[%s]<%s>", strings.Join(typeStrings, ", "), strings.Join(depStrings, ", "))
}

var Addable = Options{[]types.Type{types.FloatType{}, types.IntType{}, types.StringType{}}}

type Container struct {
	Type    types.Type
	Subtype Constrainable
	Index   int
}

func (c Container) ConsString() string {
	return fmt.Sprintf("container<%v>[%v]#%d", c.Type.TypeString(), c.Subtype.ConsString(), c.Index)
}

type Tup struct {
	Subtypes []Constrainable
}

func (t Tup) ConsString() string {
	subStrings := make([]string, 0)
	for _, sub := range t.Subtypes {
		subStrings = append(subStrings, sub.ConsString())
	}
	return fmt.Sprintf("(%s)", strings.Join(subStrings, ", "))
}

type Fun struct {
	Args []Constrainable
	Ret  Constrainable
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

	unifier := NewUnifier(infer.Constraints)
	subs, err := unifier.UnifyAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	DebugInfer("------ FINAL CONSTRAINTS ------")
	for _, c := range infer.Constraints {
		DebugInfer(c.Left.ConsString(), "=", c.Right.ConsString())
	}

	return infer.ConstructTypes(subs)
}

func (i *TypeInferer) ConstructTypes(subs Subs) map[ast.NodeHash]types.Type {
	DebugInfer("--- SUBS ---")
	for k, v := range subs {
		fmt.Println(k.ConsString(), "->", v.ConsString())
	}

	finalTypes := make(map[ast.NodeHash]types.Type)
	DebugInfer("--- FINAL TYPES ---")

	for _, subExp := range i.Subexps {
		if ast.Statement(subExp) {
			finalTypes[ast.HashNode(subExp)] = types.NullType{}
			continue
		}

		initialVar := i.GetTypeVar(subExp)
		fmt.Println("Resolving", subExp)
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

func (i *TypeInferer) ResolveType(consItem Constrainable, subs Subs) types.Type {
	switch cons := consItem.(type) {
	case BaseType:
		return cons.Type
	case TypeVar:
		nextCons, ok := subs[cons]
		if !ok {
			// Terminates at a type variable
			return types.AnyType{}
		}
		return i.ResolveType(nextCons, subs)
	case Fun:
		funType := types.FuncType{}
		for _, arg := range cons.Args {
			argType := i.ResolveType(arg, subs)
			funType.ArgTypes = append(funType.ArgTypes, argType)
		}
		funType.RetType = i.ResolveType(cons.Ret, subs)
		return funType
	case Container:
		switch cons.Type.(type) {
		case types.ArrayType:
			listType := types.ArrayType{}
			listType.Subtype = i.ResolveType(cons.Subtype, subs)
			return listType
		case types.NullType:
			nextCont, ok := subs[cons]
			if !ok {
				// TODO fix this bug in a way that actually makes sense. If a subtype gets replaced, it isn't updated in
				// Subs, so that array can no longer math a path through subs.
				subRes := i.ResolveType(cons.Subtype, subs)
				cons.Subtype = BaseType{subRes}
				nextCont, ok := subs[cons]
				if ok {
					return i.ResolveType(nextCont, subs)
				}
				return types.NullType{}
			}

			return i.ResolveType(nextCont, subs)
		}
	case Tup:
		tupType := types.TupleType{}
		for _, sub := range cons.Subtypes {
			tupType.Types = append(tupType.Types, i.ResolveType(sub, subs))
		}

		return tupType
	default:
		panic(fmt.Sprintf("Unknown constraint type %v", reflect.TypeOf(cons)))
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

// TODO complete implementation of this function, won't handle many cases
func hintToCons(hintType types.Type) Constrainable {
	switch ty := hintType.(type) {
	case types.FuncType:
		fun := Fun{}
		fun.Ret = hintToCons(ty.RetType)
		for _, argT := range ty.ArgTypes {
			fun.Args = append(fun.Args, hintToCons(argT))
		}

		return fun
	default:
		return BaseType{hintType}
	}
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

		// Constraint to assign variable name to actual function def
		i.AddCons(Constraint{i.GetTypeVar(fIdent), baseFun})
		i.AddCons(Constraint{i.GetTypeVar(fIdent), i.GetTypeVar(funDef)})

		if funDef.TypeHint != nil {
			// The user provided the type of the function
			i.AddCons(Constraint{i.GetTypeVar(fIdent), hintToCons(*funDef.TypeHint)})
		}

		// Check if function has a non-return last line. If so, setup inference for that line.
		var lastLine ast.Node
		if len(funDef.Body.Lines) > 0 {
			lastLine = funDef.Body.Lines[len(funDef.Body.Lines)-1]
		}

		_, isReturn := lastLine.(*ast.ReturnExp)
		// Don't try to implicitly return from main
		if lastLine != nil && !isReturn && fName != "main" {
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
			i.AddCons(Constraint{i.GetTypeVar(node.Target), i.GetTypeVar(node.Expr)})
			i.AddCons(Constraint{typeVar, BaseType{types.NullType{}}})
		case *ast.FunApp:
			newFun := Fun{}
			for _, arg := range node.Args {
				newFun.Args = append(newFun.Args, i.GetTypeVar(arg))
			}
			newFun.Ret = i.NewTypeVar()

			i.AddCons(Constraint{i.GetTypeVar(node.Fun), newFun})
			i.AddCons(Constraint{typeVar, newFun.Ret})

			funIdent, ok := node.Fun.(*ast.Ident)
			if !ok {
				// Function node isn't a simple identifier, can't use additional inference
				break
			}

			baseFun, ok := i.FunLookup[funIdent.Value]
			if !ok {
				// Not a 'global' level function definition. We can't use additional inference rules
				break
			}

			i.AddCons(Constraint{typeVar, baseFun.Ret})
			i.AddCons(Constraint{baseFun, newFun})
		case *ast.Ident:
		// Identifiers don't add any additional constraints
		case *ast.CompNode:
			i.AddCons(Constraint{i.GetTypeVar(node.Left), i.GetTypeVar(node.Right)})
			i.AddCons(Constraint{typeVar, BaseType{types.BoolType{}}})
		case *ast.ArrayLiteral:
			subtypeVar := i.NewTypeVar()
			i.AddCons(Constraint{typeVar, Container{types.ArrayType{types.NullType{}}, subtypeVar, 0}})
			if len(node.Exprs) > 0 {
				i.AddCons(Constraint{subtypeVar, i.GetTypeVar(node.Exprs[0])})
			}
		case *ast.TupleLiteral:
			tup := Tup{}
			for _, exp := range node.Exprs {
				expSubtype := i.GetTypeVar(exp)
				tup.Subtypes = append(tup.Subtypes, expSubtype)
			}
			i.AddCons(Constraint{typeVar, tup})
		case *ast.SliceNode:
			index := -1
			indexNode, isIndexNum := node.Index.(*ast.Num)
			if isIndexNum {
				index = int(indexNode.Value)
			}

			subtypeVar := i.NewTypeVar()
			i.AddCons(Constraint{i.GetTypeVar(node.Arr), Container{types.NullType{}, subtypeVar, index}})
			i.AddCons(Constraint{typeVar, subtypeVar})
		case *ast.StructAccess:
			options := StructOptions{[]types.Type{}, make(map[TypeVar]string)}
			fieldName := node.Field.(*ast.Ident).Value
			for _, structDef := range prog.Structs {
				for _, member := range structDef.Members {
					if member.Name.Value == fieldName {
						options.Types = append(options.Types, structDef.Type)
					}
				}
			}

			options.Dependants[typeVar] = fieldName
			i.AddCons(Constraint{i.GetTypeVar(node.Target), options})
		}
	}

	DebugInfer("------ CONSTRAINTS ------")
	for _, c := range i.Constraints {
		DebugInfer(c.Left.ConsString(), "=", c.Right.ConsString())
	}
}
