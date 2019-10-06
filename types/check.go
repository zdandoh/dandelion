package types

import (
	"ahead/ast"
	"errors"
	"fmt"
	"reflect"
)

type TypeChecker struct {
	TEnv map[string]Type
}

func NewTypeChecker() *TypeChecker {
	checker := &TypeChecker{}
	checker.TEnv = NewTEnv()

	return checker
}

func NewTEnv() map[string]Type {
	tenv := make(map[string]Type)
	tenv["p"] = FuncType{[]Type{ArrayType{AnyType{}}}, AnyType{}}

	return tenv
}

func TypeCheck(prog *ast.Program) (Type, error) {
	checker := NewTypeChecker()
	t, err := checker.TypeCheck(prog.MainFunc)
	return t, err
}

func (c *TypeChecker) TypeCheck(astNode ast.Node) (Type, error) {
	var retErr error
	var retType Type

	switch node := astNode.(type) {
	case *ast.Assign:
		c.TEnv[node.Ident], retErr = c.TypeCheck(node.Expr)
		retType = NullType{}
	case *ast.Num:
		retType = IntType{}
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
		fmt.Println(left, right)
		if lerr != nil || rerr != nil || left != right {
			retErr = errors.New("Types don't match")
		} else {
			retType = left
		}
	case *ast.FunDef:
		_, err := c.TypeCheckBlock(node.Body.Lines)
		if err != nil {
			retErr = err
		}
		// This is where smart type inference needs to happen
		retType = FuncType{make([]Type, 0), IntType{}}
	case *ast.FunApp:
		fmt.Println(node.Fun)
		funType, err := c.TypeCheck(node.Fun)
		if err != nil {
			retErr = err
			break
		}
		_, err = c.TypeCheckBlock(node.Args)
		if err != nil {
			retErr = err
			break
		}

		retType = (funType.(FuncType)).retType
	case *ast.While:
		retType = NullType{}
	case *ast.If:
		retType = NullType{}
	case *ast.CompNode:
		left, lerr := c.TypeCheck(node.Left)
		right, rerr := c.TypeCheck(node.Right)
		if lerr != nil || rerr != nil || left != right {
			retErr = errors.New("Types don't match")
		} else {
			retType = left
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
			retType = ArrayType{NullType{}}
		}
		retType = ArrayType{exprTypes[0]}
	case *ast.SliceNode:
		indexType, err := c.TypeCheck(node.Index)
		if err != nil {
			retErr = err
			break
		}
		_, isInt := indexType.(IntType)
		if !isInt {
			retErr = errors.New("Index must be int")
			break
		}

		arrType, err := c.TypeCheck(node.Arr)
		if err != nil {
			retErr = err
			break
		}

		arr, isArr := arrType.(ArrayType)
		if !isArr {
			retErr = errors.New("Must slice into array type")
			break
		}

		retType = arr.subtype
	case *ast.StrExp:
		retType = StringType{}
	default:
		panic("Typecheck not defined for type: " + reflect.TypeOf(node).String())
	}

	return retType, retErr
}

func (c *TypeChecker) TypeCheckBlock(lines []ast.Node) ([]Type, error) {
	newLines := make([]Type, 0)
	for _, line := range lines {
		lineType, err := c.TypeCheck(line)
		if err != nil {
			return nil, err
		}
		newLines = append(newLines, lineType)
	}

	return newLines, nil
}

// func (c *TypeCheck) InferTypes(ast AstNode) AstNode {
// 	var retNode AstNode

// 	switch node := ast.(type) {
// 	case *FunApp:
// 		depNode := TypeInfNode{}
// 	}

// 	return nil
// }

func (c *TypeChecker) SameType(types []Type) bool {
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
