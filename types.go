package main

import "fmt"

type Program struct {
	funcs map[string]Function
}

type Function struct {
	funcType   FuncType
	funcs      map[string]*Function
	vars       map[string]*Var
	args       []*Var
	statements []Statement
}

func (f Function) String() {
	return fmt.Sprintf("Statements")
}

type Var interface{}

type VarFloat float64
type VarInt int64
type VarBytes string
type VarArray []Var

type Var struct {
	dataType DataType
	data     []byte
}

type Statement interface {
	Run()
}

type DataType uint8
type FuncType uint8
type TokenType uint8

const (
	BytesType    DataType = 0
	IntType      DataType = 1
	FloatType    DataType = 2
	FunctionType DataType = 3
)

const (
	NormalFunc FuncType = 0
	FilterFunc FuncType = 1
	MapFunc    FuncType = 2
)

//go:generate stringer -type TokenType
const (
	VarToken TokenType = iota
	AssignToken
	StartVarGroup
	EndVarGroup
	LineEndToken

	// Operations
	MultOpToken
	DivideOpToken
	SubOpToken
	AddOpToken
	ModOpToken

	PipeToken
	EndLineToken

	// Literals
	StartArrayLiteralToken
	EndArrayLiteralToken
	ByteLiteralToken
	IntLiteralToken
	FloatLiteralToken
	HexLiteralToken

	// Function definitions
	StartFunctionDefinition
	EndFunctionDefinition

	// Comparisons
	EqualityCompareToken
)
