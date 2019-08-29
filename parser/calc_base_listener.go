// Code generated from Calc.g4 by ANTLR 4.7.2. DO NOT EDIT.

package parser // Calc

import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseCalcListener is a complete listener for a parse tree produced by CalcParser.
type BaseCalcListener struct{}

var _ CalcListener = &BaseCalcListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseCalcListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseCalcListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseCalcListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseCalcListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterStart is called when production start is entered.
func (s *BaseCalcListener) EnterStart(ctx *StartContext) {}

// ExitStart is called when production start is exited.
func (s *BaseCalcListener) ExitStart(ctx *StartContext) {}

// EnterLine is called when production line is entered.
func (s *BaseCalcListener) EnterLine(ctx *LineContext) {}

// ExitLine is called when production line is exited.
func (s *BaseCalcListener) ExitLine(ctx *LineContext) {}

// EnterArglist is called when production arglist is entered.
func (s *BaseCalcListener) EnterArglist(ctx *ArglistContext) {}

// ExitArglist is called when production arglist is exited.
func (s *BaseCalcListener) ExitArglist(ctx *ArglistContext) {}

// EnterExplist is called when production explist is entered.
func (s *BaseCalcListener) EnterExplist(ctx *ExplistContext) {}

// ExitExplist is called when production explist is exited.
func (s *BaseCalcListener) ExitExplist(ctx *ExplistContext) {}

// EnterBody is called when production body is entered.
func (s *BaseCalcListener) EnterBody(ctx *BodyContext) {}

// ExitBody is called when production body is exited.
func (s *BaseCalcListener) ExitBody(ctx *BodyContext) {}

// EnterArray is called when production Array is entered.
func (s *BaseCalcListener) EnterArray(ctx *ArrayContext) {}

// ExitArray is called when production Array is exited.
func (s *BaseCalcListener) ExitArray(ctx *ArrayContext) {}

// EnterFunDef is called when production FunDef is entered.
func (s *BaseCalcListener) EnterFunDef(ctx *FunDefContext) {}

// ExitFunDef is called when production FunDef is exited.
func (s *BaseCalcListener) ExitFunDef(ctx *FunDefContext) {}

// EnterCompExp is called when production CompExp is entered.
func (s *BaseCalcListener) EnterCompExp(ctx *CompExpContext) {}

// ExitCompExp is called when production CompExp is exited.
func (s *BaseCalcListener) ExitCompExp(ctx *CompExpContext) {}

// EnterIdent is called when production Ident is entered.
func (s *BaseCalcListener) EnterIdent(ctx *IdentContext) {}

// ExitIdent is called when production Ident is exited.
func (s *BaseCalcListener) ExitIdent(ctx *IdentContext) {}

// EnterNumber is called when production Number is entered.
func (s *BaseCalcListener) EnterNumber(ctx *NumberContext) {}

// ExitNumber is called when production Number is exited.
func (s *BaseCalcListener) ExitNumber(ctx *NumberContext) {}

// EnterFunApp is called when production FunApp is entered.
func (s *BaseCalcListener) EnterFunApp(ctx *FunAppContext) {}

// ExitFunApp is called when production FunApp is exited.
func (s *BaseCalcListener) ExitFunApp(ctx *FunAppContext) {}

// EnterMulDiv is called when production MulDiv is entered.
func (s *BaseCalcListener) EnterMulDiv(ctx *MulDivContext) {}

// ExitMulDiv is called when production MulDiv is exited.
func (s *BaseCalcListener) ExitMulDiv(ctx *MulDivContext) {}

// EnterAddSub is called when production AddSub is entered.
func (s *BaseCalcListener) EnterAddSub(ctx *AddSubContext) {}

// ExitAddSub is called when production AddSub is exited.
func (s *BaseCalcListener) ExitAddSub(ctx *AddSubContext) {}

// EnterParenExp is called when production ParenExp is entered.
func (s *BaseCalcListener) EnterParenExp(ctx *ParenExpContext) {}

// ExitParenExp is called when production ParenExp is exited.
func (s *BaseCalcListener) ExitParenExp(ctx *ParenExpContext) {}

// EnterSliceExp is called when production SliceExp is entered.
func (s *BaseCalcListener) EnterSliceExp(ctx *SliceExpContext) {}

// ExitSliceExp is called when production SliceExp is exited.
func (s *BaseCalcListener) ExitSliceExp(ctx *SliceExpContext) {}

// EnterAssign is called when production Assign is entered.
func (s *BaseCalcListener) EnterAssign(ctx *AssignContext) {}

// ExitAssign is called when production Assign is exited.
func (s *BaseCalcListener) ExitAssign(ctx *AssignContext) {}

// EnterIf is called when production If is entered.
func (s *BaseCalcListener) EnterIf(ctx *IfContext) {}

// ExitIf is called when production If is exited.
func (s *BaseCalcListener) ExitIf(ctx *IfContext) {}

// EnterFor is called when production For is entered.
func (s *BaseCalcListener) EnterFor(ctx *ForContext) {}

// ExitFor is called when production For is exited.
func (s *BaseCalcListener) ExitFor(ctx *ForContext) {}

// EnterWhile is called when production While is entered.
func (s *BaseCalcListener) EnterWhile(ctx *WhileContext) {}

// ExitWhile is called when production While is exited.
func (s *BaseCalcListener) ExitWhile(ctx *WhileContext) {}
