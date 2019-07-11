// Code generated from Math.g4 by ANTLR 4.7.2. DO NOT EDIT.

package parser // Math
import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseMathListener is a complete listener for a parse tree produced by MathParser.
type BaseMathListener struct{}

var _ MathListener = &BaseMathListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseMathListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseMathListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseMathListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseMathListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterProg is called when production prog is entered.
func (s *BaseMathListener) EnterProg(ctx *ProgContext) {}

// ExitProg is called when production prog is exited.
func (s *BaseMathListener) ExitProg(ctx *ProgContext) {}

// EnterLine is called when production line is entered.
func (s *BaseMathListener) EnterLine(ctx *LineContext) {}

// ExitLine is called when production line is exited.
func (s *BaseMathListener) ExitLine(ctx *LineContext) {}

// EnterParens is called when production parens is entered.
func (s *BaseMathListener) EnterParens(ctx *ParensContext) {}

// ExitParens is called when production parens is exited.
func (s *BaseMathListener) ExitParens(ctx *ParensContext) {}

// EnterBinOp is called when production BinOp is entered.
func (s *BaseMathListener) EnterBinOp(ctx *BinOpContext) {}

// ExitBinOp is called when production BinOp is exited.
func (s *BaseMathListener) ExitBinOp(ctx *BinOpContext) {}

// EnterInt is called when production int is entered.
func (s *BaseMathListener) EnterInt(ctx *IntContext) {}

// ExitInt is called when production int is exited.
func (s *BaseMathListener) ExitInt(ctx *IntContext) {}
