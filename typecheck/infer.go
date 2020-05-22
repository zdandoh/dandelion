package typecheck

import (
	"dandelion/ast"
	"dandelion/errs"
	"dandelion/types"
	"fmt"
	"reflect"
)

var DebugTypeInf = true

func DebugInfer(more ...interface{}) {
	if DebugTypeInf {
		fmt.Println(more...)
	}
}

func DebugInferf(format string, more ...interface{}) {
	if DebugTypeInf {
		fmt.Printf(format, more...)
	}
}

type TypeInferer struct {
	TypeNo      TypeVar
	ContainerNo int
	Subexps     []ast.Node
	HashToType  map[ast.NodeHash]TypeVar
	TypeToNode  map[TypeVar]ast.Node
	Constraints []Constraint
	FunLookup   map[string]Fun
	currFun     string
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

	DebugInfer("--- Program ast before inference ---")
	DebugInferf("%+v\n", prog)
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
	for fName, funDef := range prog.Funcs {
		infer.currFun = fName
		ast.WalkAst(funDef, infer)
	}

	infer.CreateConstraints(prog)

	unifier := NewUnifier(infer.Constraints, prog, infer.FunLookup)
	subs, _ := unifier.UnifyAll()
	errs.CheckExit()

	DebugInfer("------ FINAL CONSTRAINTS ------")
	for _, c := range infer.Constraints {
		DebugInfer(c.Left.ConsString(), "=", c.Right.ConsString())
	}

	return infer.ConstructTypes(subs)
}

func (i *TypeInferer) ConstructTypes(subs Subs) map[ast.NodeHash]types.Type {
	DebugInfer("--- SUBS ---")
	for _, pair := range subs {
		DebugInfer(pair.Old.ConsString(), "->", pair.New.ConsString())
	}

	finalTypes := make(map[ast.NodeHash]types.Type)
	DebugInfer("--- FINAL TYPES ---")

	for _, subExp := range i.Subexps {
		if ast.Statement(subExp) {
			finalTypes[ast.HashNode(subExp)] = types.VoidType{}
			continue
		}

		initialVar := i.GetTypeVar(subExp)
		DebugInfer("Resolving", subExp)
		resolvedType := i.ResolveType(initialVar, subs)

		finalTypes[ast.HashNode(subExp)] = resolvedType
		DebugInfer(subExp, "-", resolvedType.TypeString())
	}

	for fName, _ := range i.FunLookup {
		fIdent := &ast.Ident{fName, ast.NoID}
		initialVar := i.GetTypeVar(fIdent)
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
		nextCons, ok := subs.Get(cons)
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
		case types.VoidType:
			nextCont, ok := subs.Get(cons)
			if !ok {
				// TODO fix this bug in a way that actually makes sense. If a subtype gets replaced, it isn't updated in
				// Subs, so that array can no longer math a path through subs.
				subRes := i.ResolveType(cons.Subtype, subs)
				cons.Subtype = BaseType{subRes}
				nextCont, ok := subs.Get(cons)
				if ok {
					return i.ResolveType(nextCont, subs)
				}

				newCons, inSubs := subs.Get(cons)
				if inSubs {
					return i.ResolveType(newCons, subs)
				}

				return types.VoidType{}
			}

			return i.ResolveType(nextCont, subs)
		}
	case Tup:
		tupType := types.TupleType{}
		for _, sub := range cons.Subtypes {
			tupType.Types = append(tupType.Types, i.ResolveType(sub, subs))
		}

		return tupType
	case Coroutine:
		coType := types.CoroutineType{}
		coType.Yields = i.ResolveType(cons.Yields, subs)
		coType.Reads = i.ResolveType(cons.Reads, subs)

		return coType
	default:
		panic(fmt.Sprintf("Unknown constraint type %v", reflect.TypeOf(cons)))
	}

	return types.VoidType{}
}

func (i *TypeInferer) AddCons(left Constrainable, right Constrainable, source ast.Node) {
	cons := Constraint{left, right, source}
	i.Constraints = append(i.Constraints, cons)
}

func (i *TypeInferer) NewTypeVar() TypeVar {
	i.TypeNo++
	return i.TypeNo
}

func (i *TypeInferer) NewContainerID() int {
	i.ContainerNo++
	return i.ContainerNo
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

	ret, isRet := astNode.(*ast.ReturnExp)
	if isRet {
		ret.SourceFunc = i.currFun
	}
	yield, isYield := astNode.(*ast.YieldExp)
	if isYield {
		yield.SourceFunc = i.currFun
	}

	if !existed {
		i.Subexps = append(i.Subexps, astNode)
	}
}

// TODO complete implementation of this function, won't handle many cases
func (i *TypeInferer) hintToCons(hintType types.Type, box *consBox) TypeVar {
	newVar := i.NewTypeVar()
	var varVal Constrainable

	switch ty := hintType.(type) {
	case types.FuncType:
		fun := Fun{}
		fun.Ret = i.hintToCons(ty.RetType, box)

		for _, argT := range ty.ArgTypes {
			fun.Args = append(fun.Args, i.hintToCons(argT, box))
		}

		varVal = fun
	case types.CoroutineType:
		panic("Coroutine has undefined type hint syntax")
	case types.TupleType:
		tup := Tup{}
		for _, subtype := range ty.Types {
			tup.Subtypes = append(tup.Subtypes, i.hintToCons(subtype, box))
		}
		varVal = tup
	case types.ArrayType:
		subTypeVar := i.hintToCons(ty.Subtype, box)
		cont := Container{types.ArrayType{types.VoidType{}}, subTypeVar, 0, i.NewContainerID()}
		varVal = cont
	case types.StructType:
		varVal = StructOptions{[]types.Type{ty}, make(map[TypeVar]string)}
	case types.IntType, types.StringType, types.FloatType, types.ByteType, types.BoolType, types.VoidType, types.AnyType:
		varVal = BaseType{ty}
	default:
		panic("Unknown hint type: " + hintType.TypeString())
	}

	cons := Constraint{newVar, varVal, NoSource}
	box.cons = append(box.cons, cons)
	return newVar
}

func (i *TypeInferer) CreateConstraints(prog *ast.Program) {
	DebugInfer("--- ORDERED SUBEXPS ---")
	for _, astNode := range i.Subexps {
		typeVar, _ := i.NodeToTypeVar(astNode)
		DebugInfer(typeVar, "-", astNode)
	}

	// Create constraints for all globally defined functions
	for fName, funDef := range prog.Funcs {
		fIdent := &ast.Ident{fName, ast.NoID}
		baseFun := i.FunLookup[fName]

		// Constraint to assign variable name to actual function def
		i.AddCons(i.GetTypeVar(fIdent), baseFun, funDef)
		i.AddCons(i.GetTypeVar(fIdent), i.GetTypeVar(funDef), funDef)

		if funDef.TypeHint != nil {
			// The user provided the type of the function
			cons := make([]Constraint, 0)
			box := &consBox{cons}
			hintVar := i.hintToCons(*funDef.TypeHint, box)
			i.AddCons(i.GetTypeVar(fIdent), hintVar, funDef)
			i.Constraints = append(i.Constraints, box.cons...)
		}

		// Check if function has a non-return last line. If so, setup inference for that line.
		var lastLine ast.Node
		if len(funDef.Body.Lines) > 0 {
			lastLine = funDef.Body.Lines[len(funDef.Body.Lines)-1]
		} else if len(funDef.Body.Lines) == 0 && funDef.TypeHint == nil {
			// Function has no body, must be void unless hinted
			i.AddCons(baseFun.Ret, BaseType{types.VoidType{}}, funDef)
		}

		_, isReturn := lastLine.(*ast.ReturnExp)
		// Don't try to implicitly return from main or coroutines
		if lastLine != nil && !isReturn && fName != "main" && !*funDef.IsCoro {
			i.AddCons(baseFun.Ret, i.GetTypeVar(lastLine), lastLine)
		}
	}

	for _, astNode := range i.Subexps {
		typeVar, _ := i.NodeToTypeVar(astNode)

		// Create constraints for type hint
		meta := prog.Meta(astNode)
		if meta != nil && meta.Hint != nil {
			box := &consBox{make([]Constraint, 0)}
			hintVar := i.hintToCons(meta.Hint, box)
			i.AddCons(i.GetTypeVar(astNode), hintVar, astNode)
			i.Constraints = append(i.Constraints, box.cons...)
		}

		switch node := astNode.(type) {
		case *ast.Num:
			i.AddCons(typeVar, BaseType{types.IntType{}}, node)
		case *ast.FloatExp:
			i.AddCons(typeVar, BaseType{types.FloatType{}}, node)
		case *ast.StrExp:
			i.AddCons(typeVar, BaseType{types.StringType{}}, node)
		case *ast.ByteExp:
			i.AddCons(typeVar, BaseType{types.ByteType{}}, node)
		case *ast.BoolExp:
			i.AddCons(typeVar, BaseType{types.BoolType{}}, node)
		case *ast.AddSub:
			i.AddCons(i.GetTypeVar(node.Left), i.GetTypeVar(node.Right), node)
			i.AddCons(typeVar, i.GetTypeVar(node.Right), node)
		case *ast.MulDiv:
			i.AddCons(i.GetTypeVar(node.Left), i.GetTypeVar(node.Right), node)
			i.AddCons(typeVar, i.GetTypeVar(node.Right), node)
		case *ast.ParenExp:
			i.AddCons(i.GetTypeVar(node), i.GetTypeVar(node.Exp), node)
		case *ast.Assign:
			i.AddCons(i.GetTypeVar(node.Target), i.GetTypeVar(node.Expr), node)
			i.AddCons(typeVar, BaseType{types.VoidType{}}, node)
		case *ast.Closure:
			baseFun := i.FunLookup[node.Target.(*ast.Ident).Value]
			cloFun := remFirstArg(baseFun)

			i.AddCons(typeVar, cloFun, node)
			i.AddCons(baseFun.Args[0], BaseType{types.VoidType{}}, node)
		case *ast.FunApp:
			newFun := Fun{}
			for _, arg := range node.Args {
				newFun.Args = append(newFun.Args, i.GetTypeVar(arg))
			}
			newFun.Ret = i.NewTypeVar()

			i.AddCons(i.GetTypeVar(node.Fun), newFun, node)
			i.AddCons(typeVar, newFun.Ret, node)

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

			i.AddCons(typeVar, baseFun.Ret, node)
			i.AddCons(baseFun, newFun, node)
		case *ast.Ident:
		// Identifiers don't add any additional constraints
		case *ast.ReturnExp:
			sourceFun := i.FunLookup[node.SourceFunc]
			i.AddCons(i.GetTypeVar(node.Target), sourceFun.Ret, node)
		case *ast.YieldExp:
			// If a function contains a yield, it automatically returns a coroutine object
			currFun := i.FunLookup[node.SourceFunc]

			newCo := Coroutine{}
			newCo.Yields = i.GetTypeVar(node.Target)
			newCo.Reads = i.NewTypeVar()

			i.AddCons(currFun.Ret, newCo, node)
			i.AddCons(typeVar, BaseType{types.VoidType{}}, node)
		case *ast.BuiltinExp:
			i.getBuiltinConstraints(node, typeVar)
		case *ast.CompNode:
			i.AddCons(i.GetTypeVar(node.Left), i.GetTypeVar(node.Right), node)
			i.AddCons(typeVar, BaseType{types.BoolType{}}, node)
		case *ast.ArrayLiteral:
			subtypeVar := i.NewTypeVar()
			i.AddCons(typeVar, Container{types.ArrayType{types.VoidType{}}, subtypeVar, 0, i.NewContainerID()}, node)
			if len(node.Exprs) > 0 {
				i.AddCons(subtypeVar, i.GetTypeVar(node.Exprs[0]), node)
			}
		case *ast.TupleLiteral:
			tup := Tup{}
			for _, exp := range node.Exprs {
				expSubtype := i.GetTypeVar(exp)
				tup.Subtypes = append(tup.Subtypes, expSubtype)
			}
			i.AddCons(typeVar, tup, node)
		case *ast.SliceNode:
			index := -1
			indexNode, isIndexNum := node.Index.(*ast.Num)
			if isIndexNum {
				index = int(indexNode.Value)
			}

			subtypeVar := i.NewTypeVar()
			i.AddCons(i.GetTypeVar(node.Arr), Container{types.VoidType{}, subtypeVar, index, i.NewContainerID()}, node)
			i.AddCons(typeVar, subtypeVar, node)
		case *ast.TypeAssert:
			cons := make([]Constraint, 0)
			box := &consBox{cons}
			hintVar := i.hintToCons(node.TargetType, box)
			i.AddCons(typeVar, hintVar, node)
			i.Constraints = append(i.Constraints, box.cons...)
		case *ast.IsExp:
			i.AddCons(typeVar, BaseType{types.BoolType{}}, node)
		case *ast.ForIter:
			subtype := i.NewTypeVar()
			i.AddCons(i.GetTypeVar(node.Iter), Container{types.VoidType{}, subtype, 0, i.NewContainerID()}, node)
			i.AddCons(subtype, i.GetTypeVar(node.Item), node)
		case *ast.StructAccess:
			options := StructOptions{[]types.Type{}, make(map[TypeVar]string)}
			fieldName := node.Field.(*ast.Ident).Value
			for i := 0; i < prog.StructCount(); i++ {
				structDef := prog.StructNo(i)
				for _, member := range structDef.Members {
					if member.Name.Value == fieldName {
						options.Types = append(options.Types, structDef.Type)
					}
				}
				for _, method := range structDef.Methods {
					if method.Name == fieldName {
						options.Types = append(options.Types, structDef.Type)
					}
				}
			}

			options.Dependants[typeVar] = fieldName
			i.AddCons(i.GetTypeVar(node.Target), options, node)
		}
	}

	DebugInfer("------ CONSTRAINTS ------")
	for _, c := range i.Constraints {
		DebugInfer(c.Left.ConsString(), "=", c.Right.ConsString())
	}
}

func (i *TypeInferer) getBuiltinConstraints(node *ast.BuiltinExp, typeVar TypeVar) {
	switch node.Type {
	case ast.BuiltinDone:
		i.AddCons(typeVar, BaseType{types.BoolType{}}, node)
	case ast.BuiltinAny:
		i.AddCons(typeVar, BaseType{types.AnyType{}}, node)
	case ast.BuiltinSend:
		newCo := Coroutine{}
		newCo.Yields = i.NewTypeVar()
		newCo.Reads = i.NewTypeVar()

		i.AddCons(typeVar, BaseType{types.VoidType{}}, node)
		i.AddCons(i.GetTypeVar(node.Args[0]), newCo, node)
		i.AddCons(i.GetTypeVar(node.Args[1]), newCo.Reads, node)
	case ast.BuiltinNext:
		newCo := Coroutine{}
		newCo.Yields = typeVar
		newCo.Reads = i.NewTypeVar()

		i.AddCons(i.GetTypeVar(node.Args[0]), newCo, node)
	case ast.BuiltinLen:
		i.AddCons(typeVar, BaseType{types.IntType{}}, node)
	case ast.BuiltinType:
		i.AddCons(typeVar, BaseType{types.IntType{}}, node)
	}
}

func remFirstArg(baseFun Fun) Fun {
	newFun := Fun{}
	for i := 1; i < len(baseFun.Args); i++ {
		newFun.Args = append(newFun.Args, baseFun.Args[i])
	}
	newFun.Ret = baseFun.Ret

	return newFun
}
