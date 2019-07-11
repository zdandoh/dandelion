// Code generated from Math.g4 by ANTLR 4.7.2. DO NOT EDIT.

package parser // Math
import "github.com/antlr/antlr4/runtime/Go/antlr"

// A complete Visitor for a parse tree produced by MathParser.
type MathVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by MathParser#prog.
	VisitProg(ctx *ProgContext) interface{}

	// Visit a parse tree produced by MathParser#line.
	VisitLine(ctx *LineContext) interface{}

	// Visit a parse tree produced by MathParser#parens.
	VisitParens(ctx *ParensContext) interface{}

	// Visit a parse tree produced by MathParser#BinOp.
	VisitBinOp(ctx *BinOpContext) interface{}

	// Visit a parse tree produced by MathParser#int.
	VisitInt(ctx *IntContext) interface{}
}
