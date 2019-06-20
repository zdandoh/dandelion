package main

import (
	"fmt"
	"testing"
)

func TestTokenizeVarAndByteLiteral(t *testing.T) {
	answer := stringTokens([]Token{
		Token{kind: ByteLiteralToken, data: "bobie!"},
		Token{kind: AssignToken},
		Token{kind: VarToken, data: "var"},
	})

	result := stringTokens(tokenizeCode(`"bobie!" = var`))
	fmt.Println(result)
	if answer != result {
		t.Fail()
	}
}

func TestTokenizeByteLiteral(t *testing.T) {
	answer := stringTokens([]Token{
		Token{kind: ByteLiteralToken, data: "bobie!"},
	})

	result := stringTokens(tokenizeCode(`"bobie!"`))
	fmt.Println(result)
	if answer != result {
		t.Fail()
	}
}

func TestTokenizeVar(t *testing.T) {
	answer := stringTokens([]Token{
		Token{kind: VarToken, data: "variable"},
	})

	result := stringTokens(tokenizeCode(`variable`))
	fmt.Println(result)
	if answer != result {
		t.Fail()
	}
}

func TestTokenizeNumbers(t *testing.T) {
	answer := stringTokens([]Token{
		Token{kind: IntLiteralToken, data: "456"},
		Token{kind: AssignToken},
		Token{kind: FloatLiteralToken, data: "45.6704"},
		Token{kind: AssignToken},
		Token{kind: HexLiteralToken, data: "0x456"},
	})

	result := stringTokens(tokenizeCode(`456 = 45.6704 = 0x456`))
	fmt.Println(result)
	if answer != result {
		t.Fail()
	}
}

func TestTokenizeArrayLiteral(t *testing.T) {
	answer := stringTokens([]Token{
		Token{kind: StartArrayLiteralToken},
		Token{kind: IntLiteralToken, data: "1"},
		Token{kind: FloatLiteralToken, data: "2.45"},
		Token{kind: ByteLiteralToken, data: "bub"},
		Token{kind: EndArrayLiteralToken},
	})

	result := stringTokens(tokenizeCode(`[1, 2.45, "bub"]`))
	fmt.Println(result)
	if answer != result {
		t.Fail()
	}
}

func TestTokenizeMath(t *testing.T) {
	answer := stringTokens([]Token{
		Token{kind: VarToken, data: "var"},
		Token{kind: ModOpToken},
		Token{kind: IntLiteralToken, data: "5"},
		Token{kind: SubOpToken},
		Token{kind: IntLiteralToken, data: "56"},
		Token{kind: AddOpToken},
		Token{kind: IntLiteralToken, data: "34"},
		Token{kind: DivideOpToken},
		Token{kind: IntLiteralToken, data: "64"},
		Token{kind: MultOpToken},
		Token{kind: IntLiteralToken, data: "23"},
	})

	result := stringTokens(tokenizeCode(`var % 5 - 56 + 34 / 64 * 23`))
	fmt.Println(result)
	if answer != result {
		t.Fail()
	}
}

func TestTokenizePipe(t *testing.T) {
	answer := stringTokens([]Token{
		Token{kind: VarToken, data: "func"},
		Token{kind: PipeToken},
		Token{kind: VarToken, data: "func2"},
	})

	result := stringTokens(tokenizeCode(`func -> func2`))
	fmt.Println(result)
	if answer != result {
		t.Fail()
	}
}

func TestTokenizeLine(t *testing.T) {
	tokens := tokenizeCode(`files = ["hi", "bro", "what's up?"]
filter = f{ e == "bro" }
files -> p`)

	fmt.Println(stringTokens(tokens))
	fmt.Println(compileTokens(tokens))
}

func TestTokenizeVarGroup(t *testing.T) {
	fmt.Println(compileTokens(tokenizeCode(`()`)))
}
