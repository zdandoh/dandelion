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

type StructType struct {
	MemberTypes []Type
	MemberNames []string
}

func (f StructType) MemberType(memberName string) Type {
	for i, member := range f.MemberTypes {
		if f.MemberNames[i] == memberName {
			return member
		}
	}

	panic("Unknown member name:" + memberName)
}

func (f StructType) Offset(offsetName string) int {
	structOffset := -1
	for i, memberName := range f.MemberNames {
		if memberName == offsetName {
			structOffset = i
			break
		}
	}

	return structOffset
}

func (f StructType) TypeString() string {
	memberTypes := make([]string, 0)
	for _, member := range f.MemberTypes {
		memberTypes = append(memberTypes, member.TypeString())
	}
	return "struct{" + strings.Join(memberTypes, ", ") + "}"
}

type AnyType struct {
}

func (f AnyType) TypeString() string {
	return "any"
}
