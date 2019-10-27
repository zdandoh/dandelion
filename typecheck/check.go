package typecheck

import (
	"ahead/ast"
	"ahead/types"
	"errors"
	"reflect"
)

type TypeChecker struct {
	TEnv     map[string]types.Type
	Cons     Constraints
	CurrFunc *ast.FunDef
}

type Constraints map[string][]*TypeConstraint

type TypeConstraint struct {
	FunDef *ast.FunDef
	Args   []ast.Node
}

func NewTypeChecker() *TypeChecker {
	checker := &TypeChecker{}
	checker.TEnv = NewTEnv()
	checker.Cons = make(Constraints)

	return checker
}

func NewTEnv() map[string]types.Type {
	tenv := make(map[string]types.Type)
	tenv["p"] = &types.FuncType{[]types.Type{types.ArrayType{types.AnyType{}}}, types.AnyType{}}
	tenv["abs"] = &types.FuncType{[]types.Type{types.IntType{}}, types.IntType{}}

	return tenv
}

func TypeCheck(prog *ast.Program) (map[string]types.Type, error) {
	checker := NewTypeChecker()

	for name, fun := range prog.Funcs {
		checker.TEnv[name] = fun.Type
	}
	checker.TEnv["main"] = &types.FuncType{[]types.Type{}, types.IntType{}}

	for _, funDef := range prog.Funcs {
		checker.CurrFunc = funDef
		_, err := checker.TypeCheck(funDef)
		if err != nil {
			return checker.TEnv, err
		}
	}

	return checker.TEnv, nil
}

func (c *TypeChecker) TypeCheck(astNode ast.Node) (types.Type, error) {
	var retErr error
	var retType types.Type

	switch node := astNode.(type) {
	case *ast.Assign:
		switch target := node.Target.(type) {
		case *ast.Ident:
			var exprType types.Type
			exprType, retErr = c.TypeCheck(node.Expr)
			existType, exists := c.TEnv[target.Value]
			if exists && existType != exprType {
				retErr = errors.New("Cannot reassign variable type")
				break
			}
			c.TEnv[target.Value], retErr = c.TypeCheck(node.Expr)
		}

		retType = types.NullType{}
	case *ast.Num:
		retType = types.IntType{}
	case *ast.Ident:
		retType = c.TEnv[node.Value]
	case *ast.AddSub:
		left, lerr := c.TypeCheck(node.Left)
		right, rerr := c.TypeCheck(node.Right)
		if lerr != nil || rerr != nil || left != right {
			retErr = errors.New("Types don't match")
		} else {
			retType = left
		}
	case *ast.MulDiv:
		left, lerr := c.TypeCheck(node.Left)
		right, rerr := c.TypeCheck(node.Right)
		if lerr != nil || rerr != nil || left != right {
			retErr = errors.New("Types don't match")
		} else {
			retType = left
		}
	case *ast.ReturnExp:
		expType, err := c.TypeCheck(node.Target)
		if err != nil {
			retErr = err
			break
		}

		if expType != c.CurrFunc.Type.RetType {
			retErr = errors.New("Return type doesn't match function type")
		}
		retType = types.NullType{}
	case *ast.FunDef:
		for i := 0; i < len(node.Args); i++ {
			c.TEnv[node.Args[i].(*ast.Ident).Value] = node.Type.ArgTypes[i]
		}
		_, err := c.TypeCheckBlock(node.Body.Lines)
		if err != nil {
			retErr = err
		}

		retType = node.Type
	case *ast.FunApp:
		targetType, err := c.TypeCheck(node.Fun)
		if err != nil {
			retErr = err
			break
		}
		funType, ok := targetType.(*types.FuncType)
		if !ok {
			retErr = errors.New("Tried to call non-function: " + reflect.TypeOf(targetType).String())
			break
		}

		for i := 0; i < len(funType.ArgTypes); i++ {
			argType, err := c.TypeCheck(node.Args[i])
			if err != nil {
				retErr = err
				break
			}

			if argType != funType.ArgTypes[i] {
				retErr = errors.New("Incorrect type for argument to function")
				break
			}
		}

		retType = funType.RetType
	case *ast.While:
		retType = types.NullType{}
	case *ast.If:
		retType = types.NullType{}
	case *ast.CompNode:
		left, lerr := c.TypeCheck(node.Left)
		right, rerr := c.TypeCheck(node.Right)
		if lerr != nil || rerr != nil || left != right {
			retErr = errors.New("Types don't match")
		} else {
			retType = types.BoolType{}
		}
	case *ast.ArrayLiteral:
		exprTypes, err := c.TypeCheckBlock(node.Exprs)
		if err != nil {
			retErr = err
			break
		}
		if !c.SameType(exprTypes) {
			retErr = errors.New("Array literal must have all same types")
			break
		}
		if len(exprTypes) == 0 {
			retType = types.ArrayType{types.NullType{}}
		}
		retType = types.ArrayType{exprTypes[0]}
	case *ast.SliceNode:
		indexType, err := c.TypeCheck(node.Index)
		if err != nil {
			retErr = err
			break
		}
		_, isInt := indexType.(types.IntType)
		if !isInt {
			retErr = errors.New("Index must be int")
			break
		}

		arrType, err := c.TypeCheck(node.Arr)
		if err != nil {
			retErr = err
			break
		}

		arr, isArr := arrType.(types.ArrayType)
		if !isArr {
			retErr = errors.New("Must slice into array type")
			break
		}

		retType = arr.Subtype
	case *ast.StrExp:
		retType = types.StringType{}
	default:
		panic("Typecheck not defined for node: " + reflect.TypeOf(node).String())
	}

	return retType, retErr
}

func (c *TypeChecker) TypeCheckBlock(lines []ast.Node) ([]types.Type, error) {
	newLines := make([]types.Type, 0)
	for _, line := range lines {
		lineType, err := c.TypeCheck(line)
		if err != nil {
			return nil, err
		}
		newLines = append(newLines, lineType)
	}

	return newLines, nil
}

func (c *TypeChecker) SameType(types []types.Type) bool {
	if len(types) == 0 {
		return true
	}

	matchType := types[0]
	for _, t := range types {
		if t != matchType {
			return false
		}
	}

	return true
}
