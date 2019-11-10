package typecheck

import (
	"ahead/ast"
	"ahead/types"
	"errors"
	"fmt"
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
	case *ast.ParenExp:
		retType, retErr = c.TypeCheck(node.Exp)
	case *ast.Assign:
		// TODO expand this to actually check other assignment types
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
		case *ast.StructAccess:
			_, retErr = c.TypeCheck(node.Target)
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
		node.Type = types.ArrayType{exprTypes[0]}
		retType = node.Type
	case *ast.TupleLiteral:
		tupType := types.TupleType{}
		for _, elem := range node.Exprs {
			elemType, err := c.TypeCheck(elem)
			if err != nil {
				retErr = err
				break
			}
			tupType.Types = append(tupType.Types, elemType)
		}

		node.Type = tupType
		retType = tupType
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

		slicedType, err := c.TypeCheck(node.Arr)
		if err != nil {
			retErr = err
			break
		}

		switch targetType := slicedType.(type) {
		case types.ArrayType:
			retType = targetType.Subtype
		case types.TupleType:
			// Tuples must be indexed with a constant int
			literalIndex, ok := node.Index.(*ast.Num)
			if !ok {
				retErr = errors.New("Tuple must be indexed by constant integer")
				break
			}
			retType = targetType.Types[literalIndex.Value]
		default:
			retErr = errors.New("Target type doesn't support slicing")
		}
	case *ast.StrExp:
		retType = types.StringType{}
	case *ast.Pipeline:
		firstType, err := c.TypeCheck(node.Ops[0])
		if err != nil {
			retErr = err
			break
		}

		arrType, ok := firstType.(types.ArrayType)
		if !ok {
			retErr = errors.New("First element of pipeline should evaluate to array")
			break
		}

		var lastOutputType types.Type = arrType.Subtype
		for i := 1; i < len(node.Ops); i++ {
			currOp := node.Ops[i]
			opType, err := c.TypeCheck(currOp)
			if err != nil {
				retErr = err
				break
			}

			funType, ok := opType.(*types.FuncType)
			if !ok {
				retErr = errors.New("Pipe operations must be functions")
				break
			}

			fmt.Printf("%+v\n", funType)
			if len(funType.ArgTypes) != 3 ||
				!(funType.ArgTypes[0] == types.IntType{} &&
					funType.ArgTypes[1] == lastOutputType &&
					funType.ArgTypes[2] == types.ArrayType{lastOutputType}) {
				retErr = errors.New("Pipeline function doesn't have correct signature")
				break
			}

			lastOutputType = funType.RetType
		}

		retType = types.ArrayType{lastOutputType}
	case *ast.StructInstance:
		memberTypes := make([]types.Type, len(node.Values))
		memberNames := make([]string, len(node.Values))
		for i, member := range node.DefRef.Members {
			memberTypes[i] = member.Type
			memberNames[i] = member.Name.Value
		}

		structType := types.StructType{node.DefRef.Type.Name, memberTypes, memberNames}
		node.DefRef.Type = structType
		retType = structType
	case *ast.StructAccess:
		targetType, err := c.TypeCheck(node.Target)
		if err != nil {
			retErr = err
			break
		}

		structType := targetType.(types.StructType)
		node.TargetType = structType
		retType = structType.MemberType(node.Field.(*ast.Ident).Value)
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
