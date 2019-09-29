package parser

import (
	parser "ahead/aparser"
	"ahead/ast"
	"fmt"
	"math"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type calcListener struct {
	*parser.BaseCalcListener
	currNode   ast.Node
	rootNode   ast.Node
	nodeStack  NodeStack
	blockStack BlockStack
	mainFunc   *ast.FunDef
}

func (l *calcListener) EnterAddSub(c *parser.AddSubContext) {
	fmt.Println("Enter addsub " + c.GetText())
}

func (l *calcListener) ExitAddSub(c *parser.AddSubContext) {
	fmt.Println("Exit addsub " + c.GetText())

	addNode := &ast.AddSub{}
	addNode.Op = c.GetOp().GetText()
	addNode.Right = l.nodeStack.Pop()
	addNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(addNode)
}

func (l *calcListener) EnterMulDiv(c *parser.MulDivContext) {
	// l.nodeStack.Push(&MulDiv{})
	fmt.Println("Enter multdiv " + c.GetText())
}

func (l *calcListener) ExitMulDiv(c *parser.MulDivContext) {
	fmt.Println("Exit muldiv " + c.GetText())
	mulNode := &ast.MulDiv{}
	mulNode.Op = c.GetOp().GetText()
	mulNode.Right = l.nodeStack.Pop()
	mulNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(mulNode)
}

func (l *calcListener) EnterNumber(c *parser.NumberContext) {
	l.nodeStack.Push(&ast.Num{c.GetText()})
	fmt.Println("enter numb " + c.GetText())
}

func (l *calcListener) ExitNumber(c *parser.NumberContext) {
	fmt.Println("exit numb " + c.GetText())
}

func (l *calcListener) EnterIdent(c *parser.IdentContext) {
	l.nodeStack.Push(&ast.Ident{c.GetText()})
}

func (l *calcListener) ExitIdent(c *parser.IdentContext) {
	fmt.Println("Exiting ident")
}

func (l *calcListener) EnterLine(c *parser.LineContext) {
	fmt.Println("Entered line: " + c.GetText())
}

func (l *calcListener) ExitLine(c *parser.LineContext) {
	fmt.Println("Exit line: " + c.GetText())
	l.blockStack.Top.Lines = append(l.blockStack.Top.Lines, l.nodeStack.Pop())
}

func (l *calcListener) ExitExpr(c *parser.ExprContext) {
	fmt.Println("YOGO " + c.GetText())
}

func (l *calcListener) EnterStart(c *parser.StartContext) {
	// Setup, basically
	mainFunc := ast.NewFunDef()
	l.mainFunc = mainFunc

	mainBlock := &ast.Block{}
	l.blockStack.Push(mainBlock)
}

func (l *calcListener) ExitStart(c *parser.StartContext) {
	mainFunc := ast.NewFunDef()
	mainFunc.Body = l.blockStack.Pop()

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
	funApp := &ast.FunApp{}
	for i := 0; i < argCount; i++ {
		funApp.Args = append([]ast.Node{l.nodeStack.Pop()}, funApp.Args...)
	}
	funApp.Fun = l.nodeStack.Pop()

	l.nodeStack.Push(funApp)
	fmt.Println("Exiting funapp ", argCount)
}

func (l *calcListener) EnterFunDef(c *parser.FunDefContext) {
	fmt.Println("Entering fun def")
	l.blockStack.Push(&ast.Block{})
}

func (l *calcListener) ExitFunDef(c *parser.FunDefContext) {
	fmt.Println("Exiting fun def")

	funDef := ast.NewFunDef()
	args := c.GetArgs()
	argIdents := make([]ast.Node, 0)
	if args != nil {
		notCommas := filterCommas(args.GetChildren())
		for _, arg := range notCommas {
			argStr := fmt.Sprintf("%s", arg)
			argIdents = append(argIdents, &ast.Ident{argStr})
		}
	} else {
		argIdents = []ast.Node{}
	}
	funDef.Args = argIdents
	funDef.Body = l.blockStack.Pop()
	l.nodeStack.Push(funDef)
}

func (l *calcListener) EnterWhile(c *parser.WhileContext) {
	fmt.Println("Entering while")

	l.blockStack.Push(&ast.Block{})
}

func (l *calcListener) ExitWhile(c *parser.WhileContext) {
	fmt.Println("Exiting while")

	whileNode := &ast.While{}
	whileNode.Cond = l.nodeStack.Pop()
	whileNode.Body = l.blockStack.Pop()

	l.nodeStack.Push(whileNode)
}

func (l *calcListener) EnterIf(c *parser.IfContext) {
	fmt.Println("Entering if")

	l.blockStack.Push(&ast.Block{})
}

func (l *calcListener) ExitIf(c *parser.IfContext) {
	fmt.Println("Exiting if")

	ifNode := &ast.If{}
	ifNode.Cond = l.nodeStack.Pop()
	ifNode.Body = l.blockStack.Pop()

	l.nodeStack.Push(ifNode)
}

func (l *calcListener) EnterAssign(c *parser.AssignContext) {
	fmt.Println("Enter assign")
}

func (l *calcListener) ExitAssign(c *parser.AssignContext) {
	fmt.Println("Exit assign")
	assignNode := &ast.Assign{}
	assignNode.Ident = c.GetIdent().GetText()
	assignNode.Expr = l.nodeStack.Pop()
	l.nodeStack.Push(assignNode)
}

func (l *calcListener) EnterCompExp(c *parser.CompExpContext) {
	fmt.Println("Enter comp exp")
}

func (l *calcListener) ExitCompExp(c *parser.CompExpContext) {
	fmt.Println("exit comp exp")

	compNode := &ast.CompNode{}
	compNode.Op = c.GetOp().GetText()
	compNode.Right = l.nodeStack.Pop()
	compNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(compNode)
}

func (l *calcListener) EnterArray(c *parser.ArrayContext) {
	fmt.Println("Entering array literal")
}

func (l *calcListener) ExitArray(c *parser.ArrayContext) {
	fmt.Println("Exiting array literal")

	newArr := &ast.ArrayLiteral{}
	newArr.Length = len(filterCommas(c.GetElems().GetChildren()))

	for i := 0; i < newArr.Length; i++ {
		newArr.Exprs = append([]ast.Node{l.nodeStack.Pop()}, newArr.Exprs...)
	}

	l.nodeStack.Push(newArr)
}

func (l *calcListener) EnterSliceExp(c *parser.SliceExpContext) {
	fmt.Println("Entering slice exp")
}

func (l *calcListener) ExitSliceExp(c *parser.SliceExpContext) {
	fmt.Println("Exiting slice exp")

	sliceNode := &ast.SliceNode{}
	sliceNode.Index = l.nodeStack.Pop()
	sliceNode.Arr = l.nodeStack.Pop()

	l.nodeStack.Push(sliceNode)
}

func (l *calcListener) EnterStrExp(c *parser.StrExpContext) {
	fmt.Println("Entering string")
}

func (l *calcListener) ExitStrExp(c *parser.StrExpContext) {
	fmt.Println("Exiting string")
	text := c.GetText()[1 : len(c.GetText())-1]
	l.nodeStack.Push(&ast.StrExp{text})
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

func NewProgram(mainFunc *ast.FunDef) *ast.Program {
	newProg := &ast.Program{}
	newProg.MainFunc = mainFunc

	return newProg
}

func ParseProgram(text string) *ast.Program {
	is := antlr.NewInputStream(text)
	lexer := parser.NewCalcLex(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewCalc(stream)

	l := &calcListener{}
	antlr.ParseTreeWalkerDefault.Walk(l, p.Start())

	prog := NewProgram(l.mainFunc)
	return prog
}
