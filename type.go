package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Type interface {
	TypeString() string
}

type StringType struct {
}

func (s StringType) TypeString() string {
	return "string"
}

type IntType struct {
}

func (i IntType) TypeString() string {
	return "int"
}

type BoolType struct {
}

func (i BoolType) TypeString() string {
	return "bool"
}

type ArrayType struct {
	subtype Type
}

func (a ArrayType) TypeString() string {
	return fmt.Sprintf("array[%s]", a.subtype.TypeString())
}

type NullType struct {
}

func (a NullType) TypeString() string {
	return "null"
}

type FuncType struct {
	argTypes []Type
	retType  Type
}

func (f FuncType) TypeString() string {
	argStrings := make([]string, 0)
	for _, arg := range f.argTypes {
		argStrings = append(argStrings, arg.TypeString())
	}
	argString := strings.Join(argStrings, ",")
	return fmt.Sprintf("f(%s) %s", argString, f.retType.TypeString())
}

type TypeChecker struct {
	TEnv map[string]Type
}

func TypeCheck(prog *Program) (Type, error) {
	checker := &TypeChecker{}
	checker.TEnv = make(map[string]Type)

	t, err := checker.TypeCheck(prog.mainFunc)
	return t, err
}

func (c *TypeChecker) TypeCheck(ast AstNode) (Type, error) {
	var retErr error
	var retType Type

	switch node := ast.(type) {
	case *Assign:
		retType = NullType{}
	case *Num:
		retType = IntType{}
	case *Ident:
		retType = c.TEnv[node.value]
	case *AddSub:
		left, lerr := c.TypeCheck(node.left)
		right, rerr := c.TypeCheck(node.right)
		if lerr != nil || rerr != nil || left != right {
			retErr = errors.New("Types don't match")
		} else {
			retType = left
		}
	case *MulDiv:
		left, lerr := c.TypeCheck(node.left)
		right, rerr := c.TypeCheck(node.right)
		fmt.Println(left, right)
		if lerr != nil || rerr != nil || left != right {
			retErr = errors.New("Types don't match")
		} else {
			retType = left
		}
	case *FunDef:
		_, err := c.TypeCheckBlock(node.body.lines)
		if err != nil {
			retErr = err
		}
		// This is where smart type inference needs to happen
		retType = FuncType{make([]Type, 0), IntType{}}
	case *FunApp:
		funType, err := c.TypeCheck(node.fun)
		if err != nil {
			retErr = err
			break
		}
		_, err = c.TypeCheckBlock(node.args)
		if err != nil {
			retErr = err
			break
		}

		fmt.Println(funType)
		retType = (funType.(FuncType)).retType
	case *While:
		retType = NullType{}
	case *If:
		retType = NullType{}
	case *CompNode:
		left, lerr := c.TypeCheck(node.left)
		right, rerr := c.TypeCheck(node.right)
		if lerr != nil || rerr != nil || left != right {
			retErr = errors.New("Types don't match")
		} else {
			retType = left
		}
	case *ArrayLiteral:
		exprTypes, err := c.TypeCheckBlock(node.exprs)
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
	case *SliceNode:
		indexType, err := c.TypeCheck(node.index)
		if err != nil {
			retErr = err
			break
		}
		_, isInt := indexType.(IntType)
		if !isInt {
			retErr = errors.New("Index must be int")
			break
		}

		arrType, err := c.TypeCheck(node.arr)
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
	case *StrExp:
		retType = StringType{}
	default:
		panic("ApplyFunc not defined for type: " + reflect.TypeOf(node).String())
	}

	return retType, retErr
}

func (c *TypeChecker) TypeCheckBlock(lines []AstNode) ([]Type, error) {
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
