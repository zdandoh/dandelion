package main

import (
	"ahead/parsing/parser"
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type AheadVisitor struct {
	antlr.ParseTreeVisitor
}

func NewAheadVisitor() *AheadVisitor {
	return &AheadVisitor{
		ParseTreeVisitor: &parser.BaseMathVisitor{},
	}
}

func (v *AheadVisitor) Visit(tree antlr.ParseTree) interface{} {
	fmt.Println(reflect.TypeOf(tree))

	switch val := tree.(type) {
	case *parser.LineContext:
		return val.Accept(v)
	case *parser.ProgContext:
		return v.VisitProg(val)
	case *parser.BinOpContext:
		return val.Accept(v)
	}

	panic("aya!")
}

func (v *AheadVisitor) VisitProg(ctx *parser.ProgContext) interface{} {
	ctx.Line(0)
	return 0
}

func (v *AheadVisitor) VisitBinOp(ctx *parser.BinOpContext) interface{} {
	return BinOp{left: ctx.Accept(v), right: ctx.Accept(v), op: "+"}
}

func (v *AheadVisitor) VisitInt(ctx *parser.IntContext) interface{} {
	_, num := strconv.Atoi(ctx.GetText())
	return num
}
