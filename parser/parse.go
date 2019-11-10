package parser

import (
	parser "ahead/aparser"
	"ahead/ast"
	"ahead/types"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type calcListener struct {
	*parser.BaseCalcListener
	currNode   ast.Node
	rootNode   ast.Node
	nodeStack  NodeStack
	blockStack BlockStack
	mainFunc   *ast.FunDef
	typeStack  *TypeStack
	structNo   int
}

const Debug = true

func DebugPrintln(more ...interface{}) {
	if Debug {
		fmt.Println(more...)
	}
}

func (l *calcListener) EnterParenExp(c *parser.ParenExpContext) {
	DebugPrintln("Enter paren exp")
}

func (l *calcListener) ExitParenExp(c *parser.ParenExpContext) {
	DebugPrintln("Exiting paren exp")

	l.nodeStack.Push(&ast.ParenExp{l.nodeStack.Pop()})
}

func (l *calcListener) EnterAddSub(c *parser.AddSubContext) {
	DebugPrintln("Enter addsub " + c.GetText())
}

func (l *calcListener) ExitAddSub(c *parser.AddSubContext) {
	DebugPrintln("Exit addsub " + c.GetText())

	addNode := &ast.AddSub{}
	addNode.Op = c.GetOp().GetText()
	addNode.Right = l.nodeStack.Pop()
	addNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(addNode)
}

func (l *calcListener) EnterModExp(c *parser.ModExpContext) {
	DebugPrintln("Enter mod exp")
}

func (l *calcListener) ExitModExp(c *parser.ModExpContext) {
	modNode := &ast.Mod{}
	modNode.Right = l.nodeStack.Pop()
	modNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(modNode)
}

func (l *calcListener) EnterMulDiv(c *parser.MulDivContext) {
	// l.nodeStack.Push(&MulDiv{})
	DebugPrintln("Enter multdiv " + c.GetText())
}

func (l *calcListener) ExitMulDiv(c *parser.MulDivContext) {
	DebugPrintln("Exit muldiv " + c.GetText())
	mulNode := &ast.MulDiv{}
	mulNode.Op = c.GetOp().GetText()
	mulNode.Right = l.nodeStack.Pop()
	mulNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(mulNode)
}

func (l *calcListener) EnterNumber(c *parser.NumberContext) {
	DebugPrintln("enter numb " + c.GetText())

	value, err := strconv.ParseInt(c.GetText(), 10, 64)
	if err != nil {
		panic("Invalid value for int")
	}

	l.nodeStack.Push(&ast.Num{value})
}

func (l *calcListener) ExitNumber(c *parser.NumberContext) {
	DebugPrintln("exit numb " + c.GetText())
}

func (l *calcListener) EnterIdent(c *parser.IdentContext) {
}

func (l *calcListener) ExitIdent(c *parser.IdentContext) {
	DebugPrintln("Exiting ident", c.GetText())
	l.nodeStack.Push(&ast.Ident{c.GetText()})
}

func (l *calcListener) EnterStructDef(c *parser.StructDefContext) {
	DebugPrintln("Entering struct def")

	l.blockStack.Push(&ast.Block{})
}

func (l *calcListener) ExitStructDef(c *parser.StructDefContext) {
	DebugPrintln("Exiting struct def")

	structDef := l.PopStructDef()
	l.structNo++
	structDef.Type.Name = fmt.Sprintf("anon_struct%d", l.structNo)
	l.nodeStack.Push(structDef)
}

func (l *calcListener) EnterNamedStructDef(c *parser.NamedStructDefContext) {
	DebugPrintln("Entering named struct def")

	l.blockStack.Push(&ast.Block{})
}

func (l *calcListener) ExitNamedStructDef(c *parser.NamedStructDefContext) {
	DebugPrintln("Exiting named struct def")

	ident := fmt.Sprintf("%s", c.GetIdent().GetText())
	structDef := l.PopStructDef()
	structDef.Type.Name = ident
	l.nodeStack.Push(&ast.Assign{&ast.Ident{ident}, structDef})
}

func (l *calcListener) PopStructDef() *ast.StructDef {
	block := l.blockStack.Pop()
	newStruct := &ast.StructDef{}
	for _, member := range block.Lines {
		newStruct.Members = append(newStruct.Members, member.(*ast.StructMember))
	}

	return newStruct
}

func (l *calcListener) EnterTypeline(c *parser.TypelineContext) {
	DebugPrintln("Entering type line")
}

func (l *calcListener) ExitTypeline(c *parser.TypelineContext) {
	DebugPrintln("Exiting type line")

	memberName := fmt.Sprintf("%s", c.GetIdent().GetText())
	l.blockStack.Top.Lines = append(l.blockStack.Top.Lines, &ast.StructMember{&ast.Ident{memberName}, l.typeStack.Pop()})
}

func (l *calcListener) EnterBaseType(c *parser.BaseTypeContext) {
	DebugPrintln("Entering base type")
}

func (l *calcListener) ExitBaseType(c *parser.BaseTypeContext) {
	DebugPrintln("Exiting base type")
	var t types.Type
	switch c.GetText() {
	case "string":
		t = types.StringType{}
	case "int":
		t = types.IntType{}
	case "bool":
		t = types.BoolType{}
	default:
		panic(fmt.Sprintf("Unknown type '%s'", c.GetText()))
	}

	l.typeStack.Push(t)
}

func (l *calcListener) EnterTypedFun(c *parser.TypedFunContext) {
	DebugPrintln("Entering typed fun")
}

func (l *calcListener) ExitTypedFun(c *parser.TypedFunContext) {
	DebugPrintln("Exiting typed fun")
	funType := types.FuncType{}

	// TODO figure out how to do this properly
	typeCount := int(math.Ceil(float64(c.GetFtypelist().GetChildCount()) / 2.0))
	for i := 0; i < typeCount; i++ {
		funType.ArgTypes = append(funType.ArgTypes, l.typeStack.Pop())
	}
	funType.RetType = l.typeStack.Pop()
	l.typeStack.Push(funType)
}

func (l *calcListener) EnterTypedTup(c *parser.TypedTupContext) {
	DebugPrintln("Entering typed tuple")
}

func (l *calcListener) ExitTypedTup(c *parser.TypedTupContext) {
	DebugPrintln("Exiting typed tuple")

	tupType := types.TupleType{}
	typeCount := int(math.Ceil(float64(c.GetTuptypes().GetChildCount()) / 2.0))
	for i := 0; i < typeCount; i++ {
		tupType.Types = append([]types.Type{l.typeStack.Pop()}, tupType.Types...)
	}
	l.typeStack.Push(tupType)
}

func (l *calcListener) EnterTypedArr(c *parser.TypedArrContext) {
	DebugPrintln("Entering typed arr")
}

func (l *calcListener) ExitTypedArr(c *parser.TypedArrContext) {
	DebugPrintln("Exiting typed arr")
	l.typeStack.Push(types.ArrayType{l.typeStack.Pop()})
}

func (l *calcListener) EnterStructAccess(c *parser.StructAccessContext) {
	DebugPrintln("Entering struct access")
}

func (l *calcListener) ExitStructAccess(c *parser.StructAccessContext) {
	DebugPrintln("Exiting struct access")

	access := &ast.StructAccess{}
	access.Field = &ast.Ident{c.IDENT().GetText()}
	access.Target = l.nodeStack.Pop()
	l.nodeStack.Push(access)
}

func (l *calcListener) EnterLine(c *parser.LineContext) {
	DebugPrintln("Entered line: " + c.GetText())
}

func (l *calcListener) ExitLine(c *parser.LineContext) {
	DebugPrintln("Exit line: " + c.GetText())
	l.blockStack.Top.Lines = append(l.blockStack.Top.Lines, l.nodeStack.Pop())
}

func (l *calcListener) EnterStart(c *parser.StartContext) {
	// Setup, basically
	mainBlock := &ast.Block{}
	l.blockStack.Push(mainBlock)
}

func (l *calcListener) ExitStart(c *parser.StartContext) {
	mainFunc := ast.NewFunDef()
	mainFunc.Args = []ast.Node{}
	mainFunc.Type = &types.FuncType{[]types.Type{}, types.IntType{}}
	mainFunc.Body = l.blockStack.Pop()
	l.mainFunc = mainFunc
}

func (l *calcListener) EnterFunApp(c *parser.FunAppContext) {
	DebugPrintln("Entering funapp")
}

func (l *calcListener) ExitFunApp(c *parser.FunAppContext) {
	args := c.GetArgs()
	var argCount int
	if args != nil {
		// TODO figure out how to do this properly
		argCount = int(math.Ceil(float64(c.GetArgs().GetChildCount()) / 2.0))
	} else {
		argCount = 0
	}
	funApp := &ast.FunApp{}
	for i := 0; i < argCount; i++ {
		funApp.Args = append([]ast.Node{l.nodeStack.Pop()}, funApp.Args...)
	}
	funApp.Fun = l.nodeStack.Pop()

	l.nodeStack.Push(funApp)
	DebugPrintln("Exiting funapp ", argCount)
}

func (l *calcListener) EnterFunDef(c *parser.FunDefContext) {
	DebugPrintln("Entering fun def")
	l.blockStack.Push(&ast.Block{})
}

func (l *calcListener) ExitFunDef(c *parser.FunDefContext) {
	DebugPrintln("Exiting fun def")

	funDef := ast.NewFunDef()

	// TODO figure out how to do this properly
	isPipeFunc := strings.HasPrefix(c.GetText(), "f{")

	funType := &types.FuncType{}
	isFunTyped := c.GetReturntype() != nil

	if isFunTyped {
		funType.RetType = l.typeStack.Pop()
		funDef.Type = funType
	}

	var args []ast.Node
	if isPipeFunc {
		args = []ast.Node{&ast.Ident{"i"}, &ast.Ident{"e"}, &ast.Ident{"a"}}
	} else if c.GetTypedargs() != nil {
		argTypes := filterCommas(c.GetTypedargs().GetChildren())
		for _, arg := range argTypes {
			_, ok := arg.(*antlr.TerminalNodeImpl)
			if ok {
				args = append(args, &ast.Ident{fmt.Sprintf("%s", arg)})
				funType.ArgTypes = append([]types.Type{l.typeStack.Pop()}, funType.ArgTypes...)
			}
		}
	} else if c.GetArgs() != nil {
		parsedArgs := c.GetArgs()
		argTokens := filterCommas(parsedArgs.GetChildren())
		for _, arg := range argTokens {
			argStr := fmt.Sprintf("%s", arg)
			args = append(args, &ast.Ident{argStr})
		}
	} else {
		args = []ast.Node{}
	}

	funDef.Args = args
	funDef.Body = l.blockStack.Pop()
	l.nodeStack.Push(funDef)
}

func (l *calcListener) EnterWhile(c *parser.WhileContext) {
	DebugPrintln("Entering while")

	l.blockStack.Push(&ast.Block{})
}

func (l *calcListener) ExitWhile(c *parser.WhileContext) {
	DebugPrintln("Exiting while")

	whileNode := &ast.While{}
	whileNode.Cond = l.nodeStack.Pop()
	whileNode.Body = l.blockStack.Pop()

	l.nodeStack.Push(whileNode)
}

func (l *calcListener) EnterIf(c *parser.IfContext) {
	DebugPrintln("Entering if")

	l.blockStack.Push(&ast.Block{})
}

func (l *calcListener) ExitIf(c *parser.IfContext) {
	DebugPrintln("Exiting if")

	ifNode := &ast.If{}
	ifNode.Cond = l.nodeStack.Pop()
	ifNode.Body = l.blockStack.Pop()

	l.nodeStack.Push(ifNode)
}

func (l *calcListener) EnterReturn(c *parser.ReturnContext) {
	DebugPrintln("Entering return")
}

func (l *calcListener) ExitReturn(c *parser.ReturnContext) {
	DebugPrintln("Exiting return")
	l.nodeStack.Push(&ast.ReturnExp{l.nodeStack.Pop()})
}

func (l *calcListener) EnterYield(c *parser.YieldContext) {
	DebugPrintln("Entering yield")
}

func (l *calcListener) ExitYield(c *parser.YieldContext) {
	DebugPrintln("Exiting yield")
	l.nodeStack.Push(&ast.YieldExp{l.nodeStack.Pop()})
}

func (l *calcListener) EnterAssign(c *parser.AssignContext) {
	DebugPrintln("Enter assign")
}

func (l *calcListener) ExitAssign(c *parser.AssignContext) {
	DebugPrintln("Exit assign")
	assignNode := &ast.Assign{}
	assignNode.Expr = l.nodeStack.Pop()
	assignNode.Target = l.nodeStack.Pop()
	l.nodeStack.Push(assignNode)
}

func (l *calcListener) EnterCompExp(c *parser.CompExpContext) {
	DebugPrintln("Enter comp exp")
}

func (l *calcListener) ExitCompExp(c *parser.CompExpContext) {
	DebugPrintln("exit comp exp")

	compNode := &ast.CompNode{}
	compNode.Op = c.GetOp().GetText()
	compNode.Right = l.nodeStack.Pop()
	compNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(compNode)
}

func (l *calcListener) EnterArray(c *parser.ArrayContext) {
	DebugPrintln("Entering array literal")
}

func (l *calcListener) ExitArray(c *parser.ArrayContext) {
	DebugPrintln("Exiting array literal")

	newArr := &ast.ArrayLiteral{}
	newArr.Length = len(filterCommas(c.GetElems().GetChildren()))

	for i := 0; i < newArr.Length; i++ {
		newArr.Exprs = append([]ast.Node{l.nodeStack.Pop()}, newArr.Exprs...)
	}

	l.nodeStack.Push(newArr)
}

func (l *calcListener) EnterTuple(c *parser.TupleContext) {
	DebugPrintln("Entering tuple")
}

func (l *calcListener) ExitTuple(c *parser.TupleContext) {
	DebugPrintln("Exiting tuple")

	newTup := &ast.TupleLiteral{}
	elemCount := len(filterCommas(c.GetElems().GetChildren()))

	for i := 0; i < elemCount; i++ {
		newTup.Exprs = append([]ast.Node{l.nodeStack.Pop()}, newTup.Exprs...)
	}

	l.nodeStack.Push(newTup)
}

func (l *calcListener) EnterSliceExp(c *parser.SliceExpContext) {
	DebugPrintln("Entering slice exp")
}

func (l *calcListener) ExitSliceExp(c *parser.SliceExpContext) {
	DebugPrintln("Exiting slice exp")

	sliceNode := &ast.SliceNode{}
	sliceNode.Index = l.nodeStack.Pop()
	sliceNode.Arr = l.nodeStack.Pop()

	l.nodeStack.Push(sliceNode)
}

func (l *calcListener) EnterCommandExp(c *parser.CommandExpContext) {
	DebugPrintln("Entering command exp")
}

func (l *calcListener) ExitCommandExp(c *parser.CommandExpContext) {
	DebugPrintln("Exiting command exp")

	command := &ast.CommandExp{}
	splitCommand := strings.Split(c.GetText()[1:len(c.GetText())-1], " ")
	command.Command = splitCommand[0]

	// TODO: Support more advanced command syntax
	for i := 1; i < len(splitCommand); i++ {
		command.Args = append(command.Args, splitCommand[i])
	}

	l.nodeStack.Push(command)
}

func (l *calcListener) EnterPipeExp(c *parser.PipeExpContext) {
	DebugPrintln("Entering pipe exp")
}

func (l *calcListener) ExitPipeExp(c *parser.PipeExpContext) {
	DebugPrintln("Exiting pipe exp")

	pipeNode := &ast.PipeExp{}
	pipeNode.Right = l.nodeStack.Pop()
	pipeNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(pipeNode)
}

func (l *calcListener) EnterStrExp(c *parser.StrExpContext) {
	DebugPrintln("Entering string")
}

func (l *calcListener) ExitStrExp(c *parser.StrExpContext) {
	DebugPrintln("Exiting string", c.GetText())
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
	newProg.Funcs = make(map[string]*ast.FunDef)
	newProg.Structs = make(map[string]*ast.StructDef)

	newProg.Funcs["main"] = mainFunc
	return newProg
}

func ParseProgram(text string) *ast.Program {
	is := antlr.NewInputStream(text)
	lexer := parser.NewCalcLex(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewCalc(stream)

	l := &calcListener{}
	l.typeStack = &TypeStack{}
	antlr.ParseTreeWalkerDefault.Walk(l, p.Start())

	prog := NewProgram(l.mainFunc)
	return prog
}
