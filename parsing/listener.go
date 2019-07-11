package main

import (
	"ahead/parsing/parser"
	"fmt"
)

type MathListener struct {
	*parser.BaseMathListener
	currExpr *Expr
}

func Accept(parseNode interface{}) interface{} {
	switch node := parseNode.(type) {
	case *parser.IntContext:

	}
}

func MakeMathListener() *MathListener {
	listener := &MathListener{}
	listener.currExpr = nil

	return listener
}

func (l *MathListener) EnterInt(ctx *parser.IntContext) {
	fmt.Println("wowie! " + ctx.GetText())
}

func (l *MathListener) EnterExpr(ctx *parser.ExprContext) {
	fmt.Println("")
}
