package main

import (
	"ahead/parser"
	"fmt"
	"math"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type calcListener struct {
	*parser.BaseCalcListener
	currNode   AstNode
	rootNode   AstNode
	nodeStack  NodeStack
	blockStack BlockStack
	mainFunc   *FunDef
}

func (l *calcListener) EnterAddSub(c *parser.AddSubContext) {
	fmt.Println("Enter addsub " + c.GetText())
	// l.nodeStack.Push(addNode)
}

func (l *calcListener) ExitAddSub(c *parser.AddSubContext) {
	fmt.Println("Exit addsub " + c.GetText())

	addNode := &AddSub{}
	addNode.op = c.GetOp().GetText()
	addNode.right = l.nodeStack.Pop()
	addNode.left = l.nodeStack.Pop()

	l.nodeStack.Push(addNode)
}

func (l *calcListener) EnterMulDiv(c *parser.MulDivContext) {
	// l.nodeStack.Push(&MulDiv{})
	fmt.Println("Enter multdiv " + c.GetText())
}

func (l *calcListener) ExitMulDiv(c *parser.MulDivContext) {
	fmt.Println("Exit muldiv " + c.GetText())
	mulNode := &MulDiv{}
	mulNode.op = c.GetOp().GetText()
	mulNode.right = l.nodeStack.Pop()
	mulNode.left = l.nodeStack.Pop()

	l.nodeStack.Push(mulNode)
}

func (l *calcListener) EnterNumber(c *parser.NumberContext) {
	l.nodeStack.Push(&Num{c.GetText()})
	fmt.Println("enter numb " + c.GetText())
}

func (l *calcListener) ExitNumber(c *parser.NumberContext) {
	fmt.Println("exit numb " + c.GetText())
}

func (l *calcListener) EnterIdent(c *parser.IdentContext) {
	l.nodeStack.Push(&Ident{c.GetText()})
}

func (l *calcListener) ExitIdent(c *parser.IdentContext) {
	fmt.Println("Exiting ident")
}

func (l *calcListener) EnterLine(c *parser.LineContext) {
	fmt.Println("Entered line: " + c.GetText())
}

func (l *calcListener) ExitLine(c *parser.LineContext) {
	fmt.Println("Exit line: " + c.GetText())
	l.blockStack.Top.lines = append(l.blockStack.Top.lines, l.nodeStack.Pop())
}

func (l *calcListener) ExitExpr(c *parser.ExprContext) {
	fmt.Println("YOGO " + c.GetText())
}

func (l *calcListener) EnterStart(c *parser.StartContext) {
	// Setup, basically
	mainFunc := NewFunDef()
	l.mainFunc = mainFunc

	mainBlock := &Block{}
	l.blockStack.Push(mainBlock)
}

func (l *calcListener) ExitStart(c *parser.StartContext) {
	mainFunc := NewFunDef()
	mainFunc.body = l.blockStack.Pop()

	l.mainFunc = mainFunc
}

func (l *calcListener) EnterFunApp(c *parser.FunAppContext) {
	fmt.Println("Entering funapp")
}

func (l *calcListener) ExitFunApp(c *parser.FunAppContext) {
	args := c.GetArgs()
	var argCount int
	if args != nil {
		argCount = int(math.Ceil(float64(c.GetArgs().GetChildCount()) / 2.0)) // Basically the dumbest way possible to count args
	} else {
		argCount = 0
	}
	funApp := &FunApp{}
	for i := 0; i < argCount; i++ {
		funApp.args = append([]AstNode{l.nodeStack.Pop()}, funApp.args...)
	}
	funApp.fun = l.nodeStack.Pop()

	l.nodeStack.Push(funApp)
	fmt.Println("Exiting funapp ", argCount)
}

func (l *calcListener) EnterFunDef(c *parser.FunDefContext) {
	fmt.Println("Entering fun def")
	l.blockStack.Push(&Block{})
}

func (l *calcListener) ExitFunDef(c *parser.FunDefContext) {
	fmt.Println("Exiting fun def")

	funDef := NewFunDef()
	args := c.GetArgs()
	argNames := make([]string, 0)
	if args != nil {
		notCommas := filterCommas(args.GetChildren())
		for _, arg := range notCommas {
			argStr := fmt.Sprintf("%s", arg)
			argNames = append(argNames, argStr)
		}
	} else {
		argNames = []string{}
	}
	funDef.args = argNames
	funDef.body = l.blockStack.Pop()
	l.nodeStack.Push(funDef)
}

func (l *calcListener) EnterWhile(c *parser.WhileContext) {
	fmt.Println("Entering while")

	l.blockStack.Push(&Block{})
}

func (l *calcListener) ExitWhile(c *parser.WhileContext) {
	fmt.Println("Exiting while")

	whileNode := &While{}
	whileNode.cond = l.nodeStack.Pop()
	whileNode.body = l.blockStack.Pop()

	l.nodeStack.Push(whileNode)
}

func (l *calcListener) EnterIf(c *parser.IfContext) {
	fmt.Println("Entering if")

	l.blockStack.Push(&Block{})
}

func (l *calcListener) ExitIf(c *parser.IfContext) {
	fmt.Println("Exiting if")

	ifNode := &If{}
	ifNode.cond = l.nodeStack.Pop()
	ifNode.body = l.blockStack.Pop()

	l.nodeStack.Push(ifNode)
}

func (l *calcListener) EnterAssign(c *parser.AssignContext) {
	fmt.Println("Enter assign")
}

func (l *calcListener) ExitAssign(c *parser.AssignContext) {
	fmt.Println("Exit assign")
	assignNode := &Assign{}
	assignNode.ident = c.GetIdent().GetText()
	assignNode.expr = l.nodeStack.Pop()
	l.nodeStack.Push(assignNode)
}

func (l *calcListener) EnterCompExp(c *parser.CompExpContext) {
	fmt.Println("Enter comp exp")
}

func (l *calcListener) ExitCompExp(c *parser.CompExpContext) {
	fmt.Println("exit comp exp")

	compNode := &CompNode{}
	compNode.op = c.GetOp().GetText()
	compNode.right = l.nodeStack.Pop()
	compNode.left = l.nodeStack.Pop()

	l.nodeStack.Push(compNode)
}

func (l *calcListener) EnterArray(c *parser.ArrayContext) {
	fmt.Println("Entering array literal")
}

func (l *calcListener) ExitArray(c *parser.ArrayContext) {
	fmt.Println("Exiting array literal")

	newArr := &ArrayLiteral{}
	newArr.length = len(filterCommas(c.GetElems().GetChildren()))

	for i := 0; i < newArr.length; i++ {
		newArr.exprs = append([]AstNode{l.nodeStack.Pop()}, newArr.exprs...)
	}

	l.nodeStack.Push(newArr)
}

func (l *calcListener) EnterSliceExp(c *parser.SliceExpContext) {
	fmt.Println("Entering slice exp")
}

func (l *calcListener) ExitSliceExp(c *parser.SliceExpContext) {
	fmt.Println("Exiting slice exp")

	sliceNode := &SliceNode{}
	sliceNode.index = l.nodeStack.Pop()
	sliceNode.arr = l.nodeStack.Pop()

	l.nodeStack.Push(sliceNode)
}

func (l *calcListener) EnterStrExp(c *parser.StrExpContext) {
	fmt.Println("Entering string")
}

func (l *calcListener) ExitStrExp(c *parser.StrExpContext) {
	fmt.Println("Exiting string")
	text := c.GetText()[1 : len(c.GetText())-1]
	l.nodeStack.Push(&StrExp{text})
}

func filterCommas(elems []antlr.Tree) []antlr.Tree {
	notCommas := make([]antlr.Tree, 0)

	for _, elem := range elems {
		if fmt.Sprintf("%s", elem) != "," {
			notCommas = append(notCommas, elem)
		}
	}

	return notCommas
}

func ParseProgram(text string) *Program {
	is := antlr.NewInputStream(text)

	lexer := parser.NewCalcLex(is)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewCalc(stream)

	l := &calcListener{}
	antlr.ParseTreeWalkerDefault.Walk(l, p.Start())

	prog := NewProgram(l.mainFunc)
	return prog
}
