package main

type Program struct {
	funcs map[string]Function
}

type Function struct {
	funcType FuncType
	funcs    map[string]*Function
	vars     map[string]*Var
	args     []*Var
	code     []Statement
}

type Var struct {
	dataType DataType
	data     []byte
}

type Statement struct {
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

const (
	VarToken TokenType = iota
	AssignToken
	StartFuncCallToken
	EndFuncCallToken
	OpToken
	PipeToken
	EndLineToken
	StartArrayLiteralToken
	EndArrayLiteralToken
	ByteLiteralToken
	IntLiteralToken
	FloatLiteralToken
	HexLiteralToken
)

func (t TokenType) String() string {
	switch t {
	case VarToken:
		return "VarToken"
	case AssignToken:
		return "AssignToken"
	case StartFuncCallToken:
		return "StartFuncCallToken"
	case EndFuncCallToken:
		return "EndFuncCallToken"
	case OpToken:
		return "OpToken"
	case PipeToken:
		return "PipeToken"
	case EndLineToken:
		return "EndLineToken"
	case StartArrayLiteralToken:
		return "StartArrayLiteralToken"
	case EndArrayLiteralToken:
		return "EndArrayLiteralToken"
	case ByteLiteralToken:
		return "ByteLiteralToken"
	case IntLiteralToken:
		return "IntLiteralToken"
	case FloatLiteralToken:
		return "FloatLiteralToken"
	case HexLiteralToken:
		return "HexLiteralToken"
	default:
		panic("Unknown token!")
	}
}
