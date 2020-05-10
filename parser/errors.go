package parser

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"os"
)

type ErrorStrategy struct {
	parseErrors int
	antlr.DefaultErrorStrategy
}

func (e *ErrorStrategy) RecoverInline(p antlr.Parser) antlr.Token {
	e.parseErrors++
	currToken := p.GetCurrentToken()
	fmt.Fprintf(os.Stderr, "Parse Error: line %d:%d - %s unexpected\n", currToken.GetLine(), currToken.GetColumn(), e.GetTokenErrorDisplay(currToken))
	return currToken
}

type ErrorListener struct {
	*antlr.DefaultErrorListener
}

func (d *ErrorListener) SyntaxError(r antlr.Recognizer, sym interface{}, line int, column int, msg string, e antlr.RecognitionException) {
	fmt.Fprintf(os.Stderr, "Fatal Parse Error: line %d:%d - %s\n", line, column, msg)
	os.Exit(1)
}
