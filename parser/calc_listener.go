// Code generated from Calc.g4 by ANTLR 4.7.2. DO NOT EDIT.

package parser // Calc

import "github.com/antlr/antlr4/runtime/Go/antlr"

// CalcListener is a complete listener for a parse tree produced by CalcParser.
type CalcListener interface {
	antlr.ParseTreeListener

	// EnterStart is called when entering the start production.
	EnterStart(c *StartContext)

	// EnterLine is called when entering the line production.
	EnterLine(c *LineContext)

	// EnterArglist is called when entering the arglist production.
	EnterArglist(c *ArglistContext)

	// EnterExplist is called when entering the explist production.
	EnterExplist(c *ExplistContext)

	// EnterBody is called when entering the body production.
	EnterBody(c *BodyContext)

	// EnterArray is called when entering the Array production.
	EnterArray(c *ArrayContext)

	// EnterFunDef is called when entering the FunDef production.
	EnterFunDef(c *FunDefContext)

	// EnterCompExp is called when entering the CompExp production.
	EnterCompExp(c *CompExpContext)

	// EnterIdent is called when entering the Ident production.
	EnterIdent(c *IdentContext)

	// EnterNumber is called when entering the Number production.
	EnterNumber(c *NumberContext)

	// EnterFunApp is called when entering the FunApp production.
	EnterFunApp(c *FunAppContext)

	// EnterMulDiv is called when entering the MulDiv production.
	EnterMulDiv(c *MulDivContext)

	// EnterAddSub is called when entering the AddSub production.
	EnterAddSub(c *AddSubContext)

	// EnterParenExp is called when entering the ParenExp production.
	EnterParenExp(c *ParenExpContext)

	// EnterSliceExp is called when entering the SliceExp production.
	EnterSliceExp(c *SliceExpContext)

	// EnterAssign is called when entering the Assign production.
	EnterAssign(c *AssignContext)

	// EnterIf is called when entering the If production.
	EnterIf(c *IfContext)

	// EnterFor is called when entering the For production.
	EnterFor(c *ForContext)

	// EnterWhile is called when entering the While production.
	EnterWhile(c *WhileContext)

	// ExitStart is called when exiting the start production.
	ExitStart(c *StartContext)

	// ExitLine is called when exiting the line production.
	ExitLine(c *LineContext)

	// ExitArglist is called when exiting the arglist production.
	ExitArglist(c *ArglistContext)

	// ExitExplist is called when exiting the explist production.
	ExitExplist(c *ExplistContext)

	// ExitBody is called when exiting the body production.
	ExitBody(c *BodyContext)

	// ExitArray is called when exiting the Array production.
	ExitArray(c *ArrayContext)

	// ExitFunDef is called when exiting the FunDef production.
	ExitFunDef(c *FunDefContext)

	// ExitCompExp is called when exiting the CompExp production.
	ExitCompExp(c *CompExpContext)

	// ExitIdent is called when exiting the Ident production.
	ExitIdent(c *IdentContext)

	// ExitNumber is called when exiting the Number production.
	ExitNumber(c *NumberContext)

	// ExitFunApp is called when exiting the FunApp production.
	ExitFunApp(c *FunAppContext)

	// ExitMulDiv is called when exiting the MulDiv production.
	ExitMulDiv(c *MulDivContext)

	// ExitAddSub is called when exiting the AddSub production.
	ExitAddSub(c *AddSubContext)

	// ExitParenExp is called when exiting the ParenExp production.
	ExitParenExp(c *ParenExpContext)

	// ExitSliceExp is called when exiting the SliceExp production.
	ExitSliceExp(c *SliceExpContext)

	// ExitAssign is called when exiting the Assign production.
	ExitAssign(c *AssignContext)

	// ExitIf is called when exiting the If production.
	ExitIf(c *IfContext)

	// ExitFor is called when exiting the For production.
	ExitFor(c *ForContext)

	// ExitWhile is called when exiting the While production.
	ExitWhile(c *WhileContext)
}
