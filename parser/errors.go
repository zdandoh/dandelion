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
