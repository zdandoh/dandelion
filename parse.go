package main

import (
	"fmt"
	"strings"
)

const (
	ASCII_0 = 48
	ASCII_9 = 57
)

type Token struct {
	data string
	kind TokenType
}

// func parseCode(code string) *Function {
// 	main := &Function{}
// 	currFunc := main

// 	line := strings.Builder{}
// 	for i := 0; i < len(code); i++ {
// 		if code[i] == '\n' || code[i] == ';' {
// 			currFunc.parseLine(line.String())
// 			line.Reset()
// 			continue
// 		}
// 		line.WriteByte(code[i])
// 	}

// 	return main
// }

func stringTokens(tokens []Token) string {
	result := strings.Builder{}

	for _, token := range tokens {
		result.WriteString(fmt.Sprintf("%s: %s\n", token.kind, token.data))
	}

	return result.String()
}

func tokenizeCode(code string) []Token {
	tokens := make([]Token, 0)
	tokenData := strings.Builder{}
	byteLiteralInProgress := false

	for i := 0; i < len(code); i++ {
		tokenSplit := true
		var newToken *Token
		char := code[i]
		if char == '"' {
			if byteLiteralInProgress {
				tokens = append(tokens, Token{kind: ByteLiteralToken, data: tokenData.String()})
				tokenData.Reset()
				byteLiteralInProgress = false
			} else {
				byteLiteralInProgress = true
			}
			continue
		}

		if byteLiteralInProgress {
			tokenData.WriteByte(char)
			tokenSplit = false
		} else if char == ' ' || char == ',' {
			// Whitespace characters
			tokenSplit = false
		} else if ((char >= ASCII_0 && char <= ASCII_9) || char == '.') && tokenData.Len() == 0 {
			numToken, length := parseNumber(code, i)
			tokens = append(tokens, numToken)
			i += length - 1 // Subtract 1 to negate the i++ at the end of the loop
		} else if char == '=' {
			newToken = &Token{kind: AssignToken}
		} else if char == '-' && i+1 < len(code) && code[i+1] == '>' {
			// Pipe operator
			newToken = &Token{kind: PipeToken}
			i++
		} else if char == '\n' || char == ';' {
			newToken = &Token{kind: LineEndToken}
		} else if char == '*' {
			newToken = &Token{kind: MultOpToken}
		} else if char == '/' {
			newToken = &Token{kind: DivideOpToken}
		} else if char == '-' {
			newToken = &Token{kind: SubOpToken}
		} else if char == '+' {
			newToken = &Token{kind: AddOpToken}
		} else if char == '%' {
			newToken = &Token{kind: ModOpToken}
		} else if char == '(' {
			newToken = &Token{kind: StartFuncCallToken}
		} else if char == ')' {
			newToken = &Token{kind: EndFuncCallToken}
		} else if char == '[' {
			newToken = &Token{kind: StartArrayLiteralToken}
		} else if char == ']' {
			newToken = &Token{kind: EndArrayLiteralToken}
		} else {
			tokenData.WriteByte(char)
			tokenSplit = false
		}

		if (tokenSplit || i == len(code)-1) && tokenData.Len() > 0 {
			tokens = append(tokens, Token{kind: VarToken, data: tokenData.String()})
			tokenData.Reset()
		}

		if newToken != nil {
			tokens = append(tokens, *newToken)
		}
	}

	return tokens
}

func parseNumber(code string, i int) (Token, int) {
	isFloat := false
	isHex := false
	numString := strings.Builder{}

	for j := 0; j+i < len(code); j++ {
		numChar := code[i+j]
		if (numChar >= ASCII_0 && numChar <= ASCII_9) ||
			numChar == '.' ||
			numChar == 'x' {

			if numChar == '.' {
				isFloat = true
			}
			if numChar == 'x' {
				isHex = true
			}
			numString.WriteByte(numChar)
		} else {
			break
		}
	}

	token := Token{}
	if isFloat {
		token.kind = FloatLiteralToken
	} else if isHex {
		token.kind = HexLiteralToken
	} else {
		token.kind = IntLiteralToken
	}
	token.data = numString.String()

	return token, numString.Len()
}

// func (f *Function) parseLine(line string) {
// 	var assignToken string
// 	var callingFunc string
// 	token := strings.Builder{}

// 	for i := 0; i < len(line); i++ {
// 		switch line[i] {
// 		case '=':
// 			// Assignment to a variable
// 			assignToken = token.String()
// 			token.Reset()
// 		case '(':
// 			// Function call
// 			callingFunc = token.String()
// 			token.Reset()
// 		default:
// 			token.WriteByte(line[i])
// 		}
// 	}

// }

func (f *Function) parseVar(name string) {

}

func getGlobals(dest map[string]Function) []Function {
	globals := make([]Function, 0)

	return globals
}
