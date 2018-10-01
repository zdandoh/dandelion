package main

import (
	"strconv"
)

func compileTokens(tokens []Token) *Function {
	main := &Function{}
	main.funcType = NormalFunc
	main.statements = make([]Statement, 0)
	main.registers = make([]Var, 16)

	for i := 0; i < len(tokens); i++ {
		// lastToken := false
		// if i == len(tokens)-1 {
		// 	lastToken = true
		// }

		token := tokens[i]
		if token.kind > OperationsStart && token.kind < OperationsEnd {
			newStatement := main.compileOperation(tokens, i)
			main.statements = append(main.statements, newStatement)
		}
	}

	return main
}

func (f *Function) compileOperation(tokens []Token, i int) Statement {
	token := tokens[i]
	var statement Statement

	a := tokens[i-1]
	b := tokens[i+1]
	if a.kind != b.kind {
		panic("Operands not the same type")
	}
	dataType := a.kind

	switch dataType {
	case IntLiteralToken, HexLiteralToken:
		base := 10
		if dataType == HexLiteralToken {
			base = 16
		}

		newStatement := &IntOpStatement{}
		parsedNumber, _ := strconv.ParseInt(a.data, base, 64)
		newStatement.a = VarInt(parsedNumber)
		parsedNumber, _ = strconv.ParseInt(b.data, base, 64)
		newStatement.b = VarInt(parsedNumber)
		newStatement.kind = token.kind

		statement = newStatement
	}
case FloatLiteralToken:
	

	return statement
}
