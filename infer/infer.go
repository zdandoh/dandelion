package infer

import (
	"dandelion/ast"
	"dandelion/types"
	"fmt"
	"reflect"
)

type Inferer struct {
	prog *ast.Program
	currVar TypeVar
	currFuncID int
	currFunc string
	varList []StorableType
	varLibrary map[StoreKey][]TypeRef
	refs map[ast.NodeHash]TypeRef
	cons []*TCons
	currMeta int
	funLookup map[string]TypeRef
}

func NewInferer() *Inferer {
	i := &Inferer{}
	i.varLibrary = make(map[StoreKey][]TypeRef, 0)
	i.refs = make(map[ast.NodeHash]TypeRef)
	i.funLookup = make(map[string]TypeRef)
	return i
}

func (i *Inferer) AddCons(left TypeRef, right TypeRef) {
	i.cons = append(i.cons, &TCons{left, right})
}

func (i *Inferer) printCons() {
	fmt.Println("----- CONS -----")
	for _, con := range i.cons {
		fmt.Println(i.String(con.Left) + " = " + i.String(con.Right))
	}
}

func InferTypes(prog *ast.Program) map[ast.NodeHash]types.Type {
	i := NewInferer()
	i.prog = prog

	fmt.Println(prog)
	i.inferProg(prog)

	Unify(i)
	i.printCons()

	progTypes := Resolve(prog, i)

	return progTypes
}

func (i *Inferer) inferProg(prog *ast.Program) {
	// Give all functions a basic type ref that they can reference
	for name, fun := range prog.Funcs {
		i.funLookup[name] = i.funDefCons(name, fun)

		// Add the function name to the global scope
		i.AddCons(i.TypeRef(&ast.Ident{name, prog.NewNodeID()}), i.TypeRef(fun))
		if name == "main" {
			mainType := i.FuncRef(KindFunc, i.BaseRef(TypeBase{types.IntType{}}))
			i.AddCons(i.TypeRef(fun), mainType)
		}
	}

	for name, fun := range prog.Funcs {
		i.currFunc = name
		ast.WalkAst(fun, i)
	}
}

func (i *Inferer) funDefCons(fName string, node *ast.FunDef) TypeRef {
	currRef := i.TypeRef(node)
	retVar := i.NewVar()
	funRef := i.FuncRef(KindFunc, retVar, i.TypeRefs(node.Args)...)
	i.AddCons(currRef, funRef)

	if node.TypeHint != nil {
		i.AddCons(currRef, i.typeToRef(*node.TypeHint))
	}

	termExprs := node.TermExprs()
	if len(termExprs) == 0 && fName != "main" {
		i.AddCons(retVar, i.BaseRef(TypeBase{types.VoidType{}}))
	} else {
		for _, expr := range termExprs {
			retExp, isRet := expr.(*ast.ReturnExp)
			if isRet {
				i.AddCons(retVar, i.TypeRef(retExp.Target))
			} else {
				i.AddCons(retVar, i.TypeRef(expr))
			}
		}
	}

	return funRef
}

func (i *Inferer) WalkNode(astNode ast.Node) ast.Node {
	currRef := i.TypeRef(astNode)
	fmt.Println(fmt.Sprintf("%s | %s", i.Resolve(currRef), astNode))

	meta := i.prog.Meta(astNode)
	if meta != nil && meta.Hint != nil {
		i.AddCons(currRef, i.typeToRef(meta.Hint))
	}

	switch node := astNode.(type) {
	case *ast.FunDef:
		// Function definition constrains are pre-generated
	case *ast.FunApp:
		retRef := i.NewVar()
		funRef := i.FuncRef(KindFunc, retRef, i.TypeRefs(node.Args)...)
		i.AddCons(currRef, retRef)

		i.AddCons(funRef, i.TypeRef(node.Fun))
	case *ast.Closure:
		baseRef := i.funLookup[node.Target.(*ast.Ident).Value]
		baseFun := i.Resolve(baseRef).(TypeFunc)
		cloFun := i.FuncRef(KindFunc, baseFun.Ret, baseFun.Args[1:]...)

		i.AddCons(currRef, cloFun)
		i.AddCons(baseFun.Args[0], i.BaseRef(TypeBase{types.VoidType{}}))
	case *ast.Assign:
		i.AddCons(i.TypeRef(node.Target), i.TypeRef(node.Expr))
	case *ast.Num:
		i.AddCons(currRef, i.BaseRef(TypeBase{types.IntType{}}))
	case *ast.BoolExp:
		i.AddCons(currRef, i.BaseRef(TypeBase{types.BoolType{}}))
	case *ast.FloatExp:
		i.AddCons(currRef, i.BaseRef(TypeBase{types.FloatType{}}))
	case *ast.StrExp:
		i.AddCons(currRef, i.BaseRef(TypeBase{types.StringType{}}))
	case *ast.ByteExp:
		i.AddCons(currRef, i.BaseRef(TypeBase{types.ByteType{}}))
	case *ast.ArrayLiteral:
		elemType := i.NewVar()
		for _, elem := range node.Exprs {
			i.AddCons(elemType, i.TypeRef(elem))
		}

		arrRef := i.ArrRef(elemType)
		i.AddCons(currRef, arrRef)
	case *ast.SliceNode:
		elemType := i.TypeRef(node)
		arrRef := i.FuncRef(KindArray, elemType, elemType)

		i.AddCons(arrRef, i.TypeRef(node.Arr))
	case *ast.AddSub:
		i.AddCons(i.TypeRef(node.Right), i.TypeRef(node.Left))
		i.AddCons(currRef, i.TypeRef(node.Right))
	case *ast.MulDiv:
		i.AddCons(i.TypeRef(node.Right), i.TypeRef(node.Left))
		i.AddCons(currRef, i.TypeRef(node.Right))
	case *ast.Mod:
		i.AddCons(i.TypeRef(node.Left), i.TypeRef(node.Right))
		i.AddCons(currRef, i.TypeRef(node.Left))
	case *ast.CompNode:
		i.AddCons(i.TypeRef(node.Left), i.TypeRef(node.Right))
		i.AddCons(currRef, i.BaseRef(TypeBase{types.BoolType{}}))
	case *ast.Ident:
		// Identifiers don't add any additional constraints
	case *ast.ReturnExp:
		// Returns are handled when walking the function definition
	case *ast.YieldExp:
		// If a function contains a yield, it automatically returns a coroutine object
		currFun := i.Resolve(i.funLookup[i.currFunc]).(TypeFunc)

		newCo := i.CoroRef(i.TypeRef(node.Target), i.NewVar())
		i.AddCons(currFun.Ret, newCo)
	case *ast.BeginExp:
		lastItem := node.Nodes[len(node.Nodes)-1]
		i.AddCons(currRef, i.TypeRef(lastItem))
	case *ast.TupleLiteral:
		i.AddCons(currRef, i.TupleRef(i.TypeRefs(node.Exprs)...))
	case *ast.TupleAccess:
		sourceTup := i.TypeRef(node.Tup)
		tupAccess := i.FuncRef(KindTupleAccess, i.FuncMeta(node.Index), sourceTup)

		i.AddCons(currRef, tupAccess)
	case *ast.StructInstance:
		i.AddCons(currRef, i.StructRef(node.DefRef))
	case *ast.StructAccess:
		target := i.TypeRef(node.Target)
		propName := node.Field.(*ast.Ident).Value
		propAccess := i.FuncRef(KindPropAccess, i.FuncMeta(propName), target)

		i.AddCons(currRef, propAccess)
	case *ast.TypeAssert:
		targetType := i.typeToRef(node.TargetType)
		i.AddCons(currRef, targetType)
	case *ast.IsExp:
		i.AddCons(currRef, i.BaseRef(TypeBase{types.BoolType{}}))
	case *ast.ParenExp:
		i.AddCons(currRef, i.TypeRef(node.Exp))
	case *ast.BuiltinExp:
		i.genBuiltinConstraints(node, currRef)
	case *ast.While:
		i.AddCons(i.TypeRef(node.Cond), i.BaseRef(TypeBase{types.BoolType{}}))
	case *ast.For:
		i.AddCons(i.TypeRef(node.Cond), i.BaseRef(TypeBase{types.BoolType{}}))
	case *ast.ForIter:
		sourceCoro := i.CoroRef(i.TypeRef(node.Item), i.NewVar())
		i.AddCons(i.TypeRef(node.Iter), sourceCoro)
	case *ast.Pipeline:
		dataNode := node.Ops[0]
		subtype := i.NewVar()

		newContainer := i.ContainerRef(subtype)
		i.AddCons(i.TypeRef(dataNode), newContainer)

		lastRet := subtype
		for k := 1; k < len(node.Ops); k++ {
			currRet := i.NewVar()
			e := lastRet
			iArg := i.BaseRef(TypeBase{types.IntType{}})
			a := newContainer

			stageFun := i.FuncRef(KindFunc, currRet, e, iArg, a)
			i.AddCons(i.TypeRef(node.Ops[k]), stageFun)
			lastRet = currRet
		}

		i.AddCons(currRef, i.ArrRef(lastRet))
	case *ast.NullExp:
	case *ast.If:
	case *ast.BlockExp:
	case *ast.FlowControl:
	default:
		panic("type constraint selection not implemented for " + reflect.TypeOf(node).String())
	}

	return nil
}

func (i *Inferer) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}

// TODO complete implementation of this function, won't handle many cases
func (i *Inferer) typeToRef(hintType types.Type) TypeRef {
	switch ty := hintType.(type) {
	case types.FuncType:
		args := make([]TypeRef, len(ty.ArgTypes))
		for k, arg := range ty.ArgTypes {
			if arg != nil {
				// Args in the hint can be nil if there's no information for them
				args[k] = i.typeToRef(arg)
			} else {
				args[k] = i.NewVar()
			}
		}

		var retRef TypeRef
		if ty.RetType == nil {
			// Allow return type to be unspecified
			retRef = i.NewVar()
		} else {
			retRef = i.typeToRef(ty.RetType)
		}

		return i.FuncRef(KindFunc, retRef, args...)
	case types.CoroutineType:
		panic("coroutine has undefined type hint syntax")
	case types.TupleType:
		args := make([]TypeRef, len(ty.Types))
		for k, arg := range ty.Types {
			args[k] = i.typeToRef(arg)
		}
		return i.TupleRef(args...)
	case types.ArrayType:
		return i.ArrRef(i.typeToRef(ty.Subtype))
	case types.StructType:
		return i.StructRef(i.prog.Struct(ty.Name))
	case types.IntType, types.StringType, types.FloatType, types.ByteType, types.BoolType, types.VoidType, types.AnyType:
		return i.BaseRef(TypeBase{ty})
	default:
		panic("unknown type hint: " + hintType.TypeString())
	}
}

func (i *Inferer) genBuiltinConstraints(node *ast.BuiltinExp, ref TypeRef) {
	switch node.Type {
	case ast.BuiltinDone:
		i.AddCons(ref, i.BaseRef(TypeBase{types.BoolType{}}))
	case ast.BuiltinAny:
		i.AddCons(ref, i.BaseRef(TypeBase{types.AnyType{}}))
	case ast.BuiltinSend:
		newCo := i.CoroRef(i.NewVar(), i.TypeRef(node.Args[1]))
		i.AddCons(i.TypeRef(node.Args[0]), newCo)
		i.AddCons(ref, i.BaseRef(TypeBase{types.VoidType{}}))
	case ast.BuiltinNext:
		inputCoro := i.TypeRef(node.Args[0])

		newCo := i.CoroRef(ref, i.NewVar())
		i.AddCons(inputCoro, newCo)
	case ast.BuiltinLen:
		i.AddCons(ref, i.BaseRef(TypeBase{types.IntType{}}))
	case ast.BuiltinStr:
		i.AddCons(ref, i.BaseRef(TypeBase{types.StringType{}}))
	case ast.BuiltinType:
		i.AddCons(ref, i.BaseRef(TypeBase{types.IntType{}}))
	}
}