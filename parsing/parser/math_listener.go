// Code generated from Math.g4 by ANTLR 4.7.2. DO NOT EDIT.

package parser // Math
import "github.com/antlr/antlr4/runtime/Go/antlr"

// MathListener is a complete listener for a parse tree produced by MathParser.
type MathListener interface {
	antlr.ParseTreeListener

	// EnterProg is called when entering the prog production.
	EnterProg(c *ProgContext)

	// EnterLine is called when entering the line production.
	EnterLine(c *LineContext)

	// EnterParens is called when entering the parens production.
	EnterParens(c *ParensContext)

	// EnterBinOp is called when entering the BinOp production.
	EnterBinOp(c *BinOpContext)

	// EnterInt is called when entering the int production.
	EnterInt(c *IntContext)

	// ExitProg is called when exiting the prog production.
	ExitProg(c *ProgContext)

	// ExitLine is called when exiting the line production.
	ExitLine(c *LineContext)

	// ExitParens is called when exiting the parens production.
	ExitParens(c *ParensContext)

	// ExitBinOp is called when exiting the BinOp production.
	ExitBinOp(c *BinOpContext)

	// ExitInt is called when exiting the int production.
	ExitInt(c *IntContext)
}
