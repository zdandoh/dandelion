package types

import (
	"fmt"
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
	Subtype Type
}

func (a ArrayType) TypeString() string {
	return fmt.Sprintf("%s[]", a.Subtype.TypeString())
}

type NullType struct {
}

func (a NullType) TypeString() string {
	return "null"
}

type FuncType struct {
	ArgTypes []Type
	RetType  Type
}

func (f FuncType) TypeString() string {
	argStrings := make([]string, 0)
	for _, arg := range f.ArgTypes {
		argStrings = append(argStrings, arg.TypeString())
	}
	argString := strings.Join(argStrings, ",")
	return fmt.Sprintf("f(%s) %s", argString, f.RetType.TypeString())
}

type AnyType struct {
}

func (f AnyType) TypeString() string {
	return "any"
}
