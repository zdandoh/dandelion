package main

import (
	"ahead/parsing/parser"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// hugs
func main() {
	is := antlr.NewInputStream("1 * 7 + 32\n")

	lexer := parser.NewMathLexer(is)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewMathParser(tokenStream)

	listener := MakeMathListener()
	antlr.ParseTreeWalkerDefault.Walk(listener, p.Prog())
}
