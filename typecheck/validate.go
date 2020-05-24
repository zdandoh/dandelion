package typecheck

import (
	"dandelion/ast"
	"dandelion/errs"
	"dandelion/transform"
	"dandelion/types"
	"fmt"
	"reflect"
	"runtime/debug"
)

type TypeValidator struct {
	progTypes map[ast.NodeHash]types.Type
	prog      *ast.Program
}

type TypeList []types.Type
type NodeList []ast.Node

// Ast node categories
var Assignable = NodeList{&ast.Ident{}, &ast.StructAccess{}, &ast.SliceNode{}}

// Type categories
var Addable = TypeList{types.StringType{}, types.ByteType{}, types.IntType{}, types.FloatType{}}
var Number = TypeList{types.ByteType{}, types.IntType{}, types.FloatType{}}
var Natural = TypeList{types.IntType{}, types.ByteType{}}
var Sliceable = TypeList{types.TupleType{}, types.ArrayType{}, types.StringType{}}
var Index = TypeList{types.IntType{}}
var Conditional = TypeList{types.BoolType{}}
var Iterable = TypeList{types.ArrayType{}, types.CoroutineType{}}
var DotAccess = TypeList{types.StructType{}, types.ArrayType{}}
var Invocable = TypeList{types.FuncType{}}
var Nullable = TypeList{types.CoroutineType{}, types.FuncType{}, types.StructType{}, types.TupleType{}, types.VoidType{}, types.ArrayType{}, types.AnyType{}}
var Ordered = TypeList{types.IntType{}, types.BoolType{}, types.FloatType{}, types.ByteType{}}
var Lenable = TypeList{types.StringType{}, types.ArrayType{}, types.TupleType{}}

func ValidateProg(prog *ast.Program, tys map[ast.NodeHash]types.Type) {
	v := &TypeValidator{}
	v.progTypes = tys
	v.prog = prog

	for _, fun := range prog.Funcs {
		ast.WalkAst(fun, v)
	}
}

func (v *TypeValidator) Type(node ast.Node) types.Type {
	return v.progTypes[ast.HashNode(node)]
}

func (v *TypeValidator) likeType(node ast.Node, list TypeList) bool {
	ty := v.Type(node)
	for _, item := range list {
		if reflect.TypeOf(item) == reflect.TypeOf(ty) {
			return true
		}
	}

	return false
}

func (v *TypeValidator) isType(node ast.Node, list TypeList) bool {
	ty := v.Type(node)
	for _, item := range list {
		if item == ty {
			return true
		}
	}

	return false
}

func isNode(node ast.Node, list NodeList) bool {
	for _, item := range list {
		if reflect.TypeOf(item) == reflect.TypeOf(node) {
			return true
		}
	}

	return false
}

func (v *TypeValidator) checkVoid(nodes ...ast.Node) {
	for _, node := range nodes {
		nodeType := v.Type(node)
		_, isVoidType := nodeType.(types.VoidType)
		if isVoidType {
			debug.PrintStack()
			errs.Error(errs.ErrorValue, node, "void type used as value")
		}
	}
}

func (v *TypeValidator) WalkNode(astNode ast.Node) ast.Node {
	switch node := astNode.(type) {
	case *ast.Assign:
		v.checkVoid(node.Expr)
		if !isNode(node.Target, Assignable) {
			errs.Error(errs.ErrorType, node, "target is not assignable")
		}
	case *ast.AddSub:
		v.checkVoid(node.Left, node.Right)
		if node.Op == "+" && (!v.isType(node.Right, Addable) || !v.isType(node.Left, Addable)) {
			errs.Error(errs.ErrorType, node, "operand is not addable")
		}
		if node.Op == "-" && (!v.isType(node.Right, Number) || !v.isType(node.Left, Number)) {
			errs.Error(errs.ErrorType, node, "operand is not a number")
		}
	case *ast.SliceNode:
		v.checkVoid(node.Index)
		if transform.IsCloArg(node.Arr) {
			break
		}

		v.checkVoid(node.Arr)

		if !v.likeType(node.Arr, Sliceable) {
			ty := v.Type(node.Arr)
			errs.Error(errs.ErrorType, node.Arr, "type '%s' is not sliceable", ty.TypeString())
		}
		if !v.isType(node.Index, Index) {
			ty := v.Type(node.Index)
			errs.Error(errs.ErrorType, node.Index, "type '%s' is not a valid index", ty.TypeString())
		}
	case *ast.ForIter:
		v.checkVoid(node.Iter, node.Item)
		if !v.likeType(node.Iter, Iterable) {
			ty := v.Type(node.Iter)
			errs.Error(errs.ErrorType, node.Iter, "type '%s' is not iterable", ty.TypeString())
		}
	case *ast.For:
		v.checkVoid(node.Cond)
		if !v.isType(node.Cond, Conditional) {
			ty := v.Type(node.Cond)
			errs.Error(errs.ErrorType, node.Cond, "type '%s' is not a valid conditional", ty.TypeString())
		}
	case *ast.While:
		v.checkVoid(node.Cond)
		if !v.isType(node.Cond, Conditional) {
			ty := v.Type(node.Cond)
			errs.Error(errs.ErrorType, node.Cond, "type '%s' is not a valid conditional", ty.TypeString())
		}
	case *ast.If:
		v.checkVoid(node.Cond)
		if !v.isType(node.Cond, Conditional) {
			ty := v.Type(node.Cond)
			errs.Error(errs.ErrorType, node.Cond, "type '%s' is not a valid conditional", ty.TypeString())
		}
	case *ast.StructAccess:
		v.checkVoid(node.Target)
		if !v.likeType(node.Target, DotAccess) {
			ty := v.Type(node.Target)
			errs.Error(errs.ErrorType, node.Target, "can't access attribute of type '%s'", ty.TypeString())
		}
	case *ast.MulDiv:
		v.checkVoid(node.Left, node.Right)
		if !v.isType(node.Left, Number) || !v.isType(node.Right, Number) {
			errs.Error(errs.ErrorType, node, "operand is not number")
		}
	case *ast.FunApp:
		v.checkVoid(node.Fun)

		for _, arg := range node.Args {
			if !transform.IsCloArg(arg) {
				v.checkVoid(arg)
			}
		}

		if !v.likeType(node.Fun, Invocable) {
			errs.Error(errs.ErrorType, node, "target is not invocable")
		}
	case *ast.NullExp:
		if !v.likeType(node, Nullable) {
			ty := v.Type(node)
			errs.Error(errs.ErrorType, node, "target of type '%s' isn't nullable", ty.TypeString())
		}
	case *ast.CompNode:
		v.checkVoid(node.Left, node.Right)
		if node.Op == "==" || node.Op == "!=" {
			// These don't need to be ordered
			break
		}

		if !v.isType(node.Right, Ordered) || !v.isType(node.Left, Ordered) {
			errs.Error(errs.ErrorType, node, "operand isn't ordered")
		}
	case *ast.ArrayLiteral:
		v.checkVoid(node.Exprs...)
		if len(node.Exprs) < 1 {
			break
		}
		nodeType := v.Type(node.Exprs[0])
		for _, node := range node.Exprs {
			if !types.Equals(v.Type(node), nodeType) {
				errs.Error(errs.ErrorType, node, "list elements must be of same type")
				break
			}
		}
	case *ast.Mod:
		v.checkVoid(node.Left, node.Right)
		if !v.isType(node.Left, Natural) || !v.isType(node.Right, Natural) {
			errs.Error(errs.ErrorType, node, "operand isn't a natural number")
		}
	case *ast.BuiltinExp:
		v.checkVoid(node.Args...)
		switch node.Type {
		case ast.BuiltinDone:
			ty := v.Type(node.Args[0])
			_, isCoro := ty.(types.CoroutineType)
			if !isCoro {
				errs.Error(errs.ErrorType, node, "argument to done must be coroutine")
			}
		case ast.BuiltinNext:
			ty := v.Type(node.Args[0])
			_, isCoro := ty.(types.CoroutineType)
			if !isCoro {
				errs.Error(errs.ErrorType, node, "argument to next must be coroutine")
			}
		case ast.BuiltinSend:
			ty := v.Type(node.Args[0])
			_, isCoro := ty.(types.CoroutineType)
			if !isCoro {
				errs.Error(errs.ErrorType, node, "argument to send must be coroutine")
			}
		case ast.BuiltinLen:
			if !v.likeType(node.Args[0], Lenable) {
				ty := v.Type(node.Args[0])
				errs.Error(errs.ErrorType, node, "cannot take length of type '%s'", ty.TypeString())
			}
		case ast.BuiltinAny:
		case ast.BuiltinType:
		default:
			panic("Validation step undefined for builtin: " + node.Type)
		}
	case *ast.StructDef:
		for _, member := range node.Members {
			_, isVoid := member.Type.(types.VoidType)
			if isVoid {
				errs.Error(errs.ErrorValue, node, "void type not allowed in struct")
			}
		}
	case *ast.IsExp:
		v.checkVoid(node.CheckNode)
	case *ast.TypeAssert:
		v.checkVoid(node.Target)
	case *ast.YieldExp:
		v.checkVoid(node.Target)
	case *ast.TupleLiteral:
		v.checkVoid(node.Exprs...)
	case *ast.BeginExp:
	case *ast.FlowControl:
	case *ast.ParenExp:
	case *ast.Closure:
	case *ast.ReturnExp:
	case *ast.BlockExp:
	case *ast.PipeExp:
	case *ast.Pipeline:
	case *ast.StructInstance:
	case *ast.FunDef:
	case *ast.Ident:
	case *ast.Num:
	case *ast.StrExp:
	case *ast.BoolExp:
	case *ast.FloatExp:
	case *ast.ByteExp:
	default:
		panic(fmt.Sprintf("Validation step undefined for node: %v", reflect.TypeOf(node)))
	}

	return nil
}

func (v *TypeValidator) WalkBlock(block *ast.Block) *ast.Block {
	return nil
}
