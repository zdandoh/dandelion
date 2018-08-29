package main

func compileTokens(tokens []Token) *Function {
	main := &Function{}
	main.funcType = NormalFunc
	currFunc := main

	for i := 0; i < len(tokens); i++ {
		lastToken := false
		if i == len(tokens)-1 {
			lastToken = true
		}

		token := tokens[i]
	}

	return main
}

func (f *Function) compileOperation(tokens []Token, op_index int) {
	var opStatement Statement
	op_token := tokens[op_index]

	switch op_token.kind {
	case AddOpToken:
	}

	if op_token[op_index-1] == EndVarGroup {

	}
}
