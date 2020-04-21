package parser

import (
	parser "dandelion/aparser"
	"dandelion/ast"
	"dandelion/types"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type listener struct {
	*parser.BaseDandelionListener
	currNode   ast.Node
	rootNode   ast.Node
	nodeStack  NodeStack
	blockStack BlockStack
	typeStack  *TypeStack
	structNo   int
	nodeID     ast.NodeID
	emptyArrNo int
	nullNo     int
	prog       *ast.Program
}

const Debug = true
const ExternPrefix = "__extern_"

func DebugPrintln(more ...interface{}) {
	if Debug {
		fmt.Println(more...)
	}
}

func (l *listener) NewNodeID() ast.NodeID {
	l.nodeID++

	newMeta := &ast.Meta{parser.LineCounter, nil}
	l.prog.Metadata[l.nodeID] = newMeta

	return l.nodeID
}

func (l *listener) EnterParenExp(c *parser.ParenExpContext) {
	DebugPrintln("Enter paren exp")
}

func (l *listener) ExitParenExp(c *parser.ParenExpContext) {
	DebugPrintln("Exiting paren exp")

	l.nodeStack.Push(&ast.ParenExp{l.nodeStack.Pop(), l.NewNodeID()})
}

func (l *listener) EnterAddSub(c *parser.AddSubContext) {
	DebugPrintln("Enter addsub " + c.GetText())
}

func (l *listener) ExitAddSub(c *parser.AddSubContext) {
	DebugPrintln("Exit addsub " + c.GetText())

	addNode := &ast.AddSub{}
	addNode.Op = c.GetOp().GetText()
	addNode.Right = l.nodeStack.Pop()
	addNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(addNode)
}

func (l *listener) EnterModExp(c *parser.ModExpContext) {
	DebugPrintln("Enter mod exp")
}

func (l *listener) ExitModExp(c *parser.ModExpContext) {
	modNode := &ast.Mod{}
	modNode.Right = l.nodeStack.Pop()
	modNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(modNode)
}

func (l *listener) EnterMulDiv(c *parser.MulDivContext) {
	DebugPrintln("Enter multdiv " + c.GetText())
}

func (l *listener) ExitMulDiv(c *parser.MulDivContext) {
	DebugPrintln("Exit muldiv " + c.GetText())
	mulNode := &ast.MulDiv{}
	mulNode.Op = c.GetOp().GetText()
	mulNode.Right = l.nodeStack.Pop()
	mulNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(mulNode)
}

func (l *listener) EnterNumber(c *parser.NumberContext) {
	DebugPrintln("enter numb " + c.GetText())

	value, err := strconv.ParseInt(c.GetText(), 10, 64)
	if err != nil {
		panic("Invalid value for int")
	}

	l.nodeStack.Push(&ast.Num{value, l.NewNodeID()})
}

func (l *listener) ExitNumber(c *parser.NumberContext) {
	DebugPrintln("exit numb " + c.GetText())
}

func (l *listener) EnterIdent(c *parser.IdentContext) {
}

func (l *listener) ExitIdent(c *parser.IdentContext) {
	DebugPrintln("Exiting ident", c.GetText())

	newIdent := &ast.Ident{}
	newIdent.Value = c.GetId().GetText()
	newIdent.NodeID = l.NewNodeID()

	idType := c.GetIdtype()
	if idType != nil {
		newType := l.typeStack.Pop()
		meta := l.prog.Meta(newIdent)
		meta.Hint = newType
	}

	l.nodeStack.Push(newIdent)
}

func (l *listener) EnterStructDef(c *parser.StructDefContext) {
	DebugPrintln("Entering struct def")

	l.blockStack.Push(&ast.Block{})
}

func (l *listener) ExitStructDef(c *parser.StructDefContext) {
	DebugPrintln("Exiting struct def")

	structDef := l.PopStructDef()
	l.structNo++
	structDef.Type.Name = fmt.Sprintf("anon_struct%d", l.structNo)
	l.nodeStack.Push(structDef)
}

func (l *listener) EnterNamedStructDef(c *parser.NamedStructDefContext) {
	DebugPrintln("Entering named struct def")

	l.blockStack.Push(&ast.Block{})
}

func (l *listener) ExitNamedStructDef(c *parser.NamedStructDefContext) {
	DebugPrintln("Exiting named struct def")

	ident := fmt.Sprintf("%s", c.GetIdent().GetText())
	structDef := l.PopStructDef()
	structDef.Type.Name = ident
	l.nodeStack.Push(&ast.Assign{&ast.Ident{ident, l.NewNodeID()}, structDef, l.NewNodeID()})
}

func (l *listener) PopStructDef() *ast.StructDef {
	block := l.blockStack.Pop()
	newStruct := &ast.StructDef{}
	for _, member := range block.Lines {
		newStruct.Members = append(newStruct.Members, member.(*ast.StructMember))
	}

	return newStruct
}

func (l *listener) EnterTypeline(c *parser.TypelineContext) {
	DebugPrintln("Entering type line")
}

func (l *listener) ExitTypeline(c *parser.TypelineContext) {
	DebugPrintln("Exiting type line")

	memberName := fmt.Sprintf("%s", c.GetIdent().GetText())
	l.blockStack.Top.Lines = append(l.blockStack.Top.Lines, &ast.StructMember{&ast.Ident{memberName, l.NewNodeID()}, l.typeStack.Pop(), l.NewNodeID()})
}

func (l *listener) EnterBaseType(c *parser.BaseTypeContext) {
	DebugPrintln("Entering base type")
}

func (l *listener) ExitBaseType(c *parser.BaseTypeContext) {
	DebugPrintln("Exiting base type")
	text := c.GetText()
	var t types.Type
	switch text {
	case "string":
		t = types.StringType{}
	case "int":
		t = types.IntType{}
	case "bool":
		t = types.BoolType{}
	case "float":
		t = types.FloatType{}
	case "byte":
		t = types.ByteType{}
	case "void":
		t = types.NullType{}
	default:
		t = types.StructType{text}
	}

	l.typeStack.Push(t)
}

func (l *listener) EnterTypedFun(c *parser.TypedFunContext) {
	DebugPrintln("Entering typed fun")
}

func (l *listener) ExitTypedFun(c *parser.TypedFunContext) {
	DebugPrintln("Exiting typed fun")
	funType := types.FuncType{}

	// TODO figure out how to do this properly
	typeCount := int(math.Ceil(float64(c.GetFtypelist().GetChildCount()) / 2.0))
	funType.RetType = l.typeStack.Pop()
	for i := 0; i < typeCount; i++ {
		funType.ArgTypes = append([]types.Type{l.typeStack.Pop()}, funType.ArgTypes...)
	}
	l.typeStack.Push(funType)
}

func (l *listener) EnterTypedTup(c *parser.TypedTupContext) {
	DebugPrintln("Entering typed tuple")
}

func (l *listener) ExitTypedTup(c *parser.TypedTupContext) {
	DebugPrintln("Exiting typed tuple")

	tupType := types.TupleType{}
	typeCount := int(math.Ceil(float64(c.GetTuptypes().GetChildCount()) / 2.0))
	for i := 0; i < typeCount; i++ {
		tupType.Types = append([]types.Type{l.typeStack.Pop()}, tupType.Types...)
	}
	l.typeStack.Push(tupType)
}

func (l *listener) EnterTypedArr(c *parser.TypedArrContext) {
	DebugPrintln("Entering typed arr")
}

func (l *listener) ExitTypedArr(c *parser.TypedArrContext) {
	DebugPrintln("Exiting typed arr")
	l.typeStack.Push(types.ArrayType{l.typeStack.Pop()})
}

func (l *listener) EnterStructAccess(c *parser.StructAccessContext) {
	DebugPrintln("Entering struct access")
}

func (l *listener) ExitStructAccess(c *parser.StructAccessContext) {
	DebugPrintln("Exiting struct access")

	access := &ast.StructAccess{}
	access.Field = &ast.Ident{c.IDENT().GetText(), l.NewNodeID()}
	access.Target = l.nodeStack.Pop()
	l.nodeStack.Push(access)
}

func (l *listener) EnterNextExp(c *parser.NextExpContext) {
	DebugPrintln("Entering next exp")
}

func (l *listener) ExitNextExp(c *parser.NextExpContext) {
	DebugPrintln("Exiting next exp")

	nextExp := &ast.NextExp{l.nodeStack.Pop(), l.NewNodeID()}
	l.nodeStack.Push(nextExp)
}

func (l *listener) EnterSendExp(c *parser.SendExpContext) {
	DebugPrintln("Entering send exp")
}

func (l *listener) ExitSendExp(c *parser.SendExpContext) {
	DebugPrintln("Exiting send exp")

	val := l.nodeStack.Pop()
	target := l.nodeStack.Pop()
	nextExp := &ast.SendExp{target, val, l.NewNodeID()}
	l.nodeStack.Push(nextExp)
}

func (l *listener) EnterLine(c *parser.LineContext) {
	DebugPrintln("Entered line: " + c.GetText())
}

func (l *listener) ExitLine(c *parser.LineContext) {
	DebugPrintln("Exit line: " + c.GetText())
	l.blockStack.Top.Lines = append(l.blockStack.Top.Lines, l.nodeStack.Pop())
}

func (l *listener) EnterStart(c *parser.StartContext) {
	// Setup, basically
	mainBlock := &ast.Block{}
	l.blockStack.Push(mainBlock)
}

func (l *listener) ExitStart(c *parser.StartContext) {
	mainFunc := ast.NewFunDef()
	mainFunc.Args = []ast.Node{}
	mainFunc.TypeHint = &types.FuncType{[]types.Type{}, types.IntType{}}
	mainFunc.Body = l.blockStack.Pop()
	l.prog.Funcs["main"] = mainFunc
}

func (l *listener) EnterFunApp(c *parser.FunAppContext) {
	DebugPrintln("Entering funapp")
}

func (l *listener) ExitFunApp(c *parser.FunAppContext) {
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

	appIdent, isIdent := funApp.Fun.(*ast.Ident)
	if isIdent && strings.HasPrefix(appIdent.Value, ExternPrefix) {
		funApp.Extern = true
	}

	l.nodeStack.Push(funApp)
	DebugPrintln("Exiting funapp ", argCount)
}

func (l *listener) EnterFunDef(c *parser.FunDefContext) {
	DebugPrintln("Entering fun def")
	l.blockStack.Push(&ast.Block{})
}

func (l *listener) ExitFunDef(c *parser.FunDefContext) {
	DebugPrintln("Exiting fun def")

	funDef := ast.NewFunDef()

	// TODO figure out how to do this properly
	isPipeFunc := strings.HasPrefix(c.GetText(), "f{")

	funType := types.FuncType{}
	isFunTyped := c.GetReturntype() != nil

	if isFunTyped {
		funType.RetType = l.typeStack.Pop()
	}

	var args []ast.Node
	if isPipeFunc {
		args = []ast.Node{&ast.Ident{"i", l.NewNodeID()}, &ast.Ident{"e", l.NewNodeID()}, &ast.Ident{"a", l.NewNodeID()}}
	} else if c.GetTypedargs() != nil {
		argTypes := filterCommas(c.GetTypedargs().GetChildren())
		for _, arg := range argTypes {
			_, ok := arg.(*antlr.TerminalNodeImpl)
			if ok {
				args = append(args, &ast.Ident{fmt.Sprintf("%s", arg), l.NewNodeID()})
				funType.ArgTypes = append([]types.Type{l.typeStack.Pop()}, funType.ArgTypes...)
			}
		}
	} else if c.GetArgs() != nil {
		parsedArgs := c.GetArgs()
		argTokens := filterCommas(parsedArgs.GetChildren())
		for _, arg := range argTokens {
			argStr := fmt.Sprintf("%s", arg)
			args = append(args, &ast.Ident{argStr, l.NewNodeID()})
		}
	} else {
		args = []ast.Node{}
	}

	funDef.Args = args
	funDef.Body = l.blockStack.Pop()

	if isFunTyped {
		funDef.TypeHint = &funType
	}
	l.nodeStack.Push(funDef)
}

func (l *listener) EnterWhile(c *parser.WhileContext) {
	DebugPrintln("Entering while")

	l.blockStack.Push(&ast.Block{})
}

func (l *listener) ExitWhile(c *parser.WhileContext) {
	DebugPrintln("Exiting while")

	whileNode := &ast.While{}
	whileNode.Cond = l.nodeStack.Pop()
	whileNode.Body = l.blockStack.Pop()

	l.nodeStack.Push(whileNode)
}

func (l *listener) EnterBlockExp(c *parser.BlockExpContext) {
	DebugPrintln("Entering block exp")

	l.blockStack.Push(&ast.Block{})
}

func (l *listener) ExitBlockExp(c *parser.BlockExpContext) {
	l.nodeStack.Push(&ast.BlockExp{l.blockStack.Pop(), l.NewNodeID()})
}

func (l *listener) EnterIf(c *parser.IfContext) {
	DebugPrintln("Entering if")

	l.blockStack.Push(&ast.Block{})
}

func (l *listener) ExitIf(c *parser.IfContext) {
	DebugPrintln("Exiting if")

	ifNode := &ast.If{}
	ifNode.Cond = l.nodeStack.Pop()
	ifNode.Body = l.blockStack.Pop()

	l.nodeStack.Push(ifNode)
}

func (l *listener) EnterReturn(c *parser.ReturnContext) {
	DebugPrintln("Entering return")
}

func (l *listener) ExitReturn(c *parser.ReturnContext) {
	DebugPrintln("Exiting return")
	l.nodeStack.Push(&ast.ReturnExp{l.nodeStack.Pop(), "", l.NewNodeID()})
}

func (l *listener) EnterYield(c *parser.YieldContext) {
	DebugPrintln("Entering yield")
}

func (l *listener) ExitYield(c *parser.YieldContext) {
	DebugPrintln("Exiting yield")
	l.nodeStack.Push(&ast.YieldExp{l.nodeStack.Pop(), "", l.NewNodeID()})
}

func (l *listener) EnterAssign(c *parser.AssignContext) {
	DebugPrintln("Enter assign")
}

func (l *listener) ExitAssign(c *parser.AssignContext) {
	DebugPrintln("Exit assign")
	assignNode := &ast.Assign{}
	assignNode.Expr = l.nodeStack.Pop()
	assignNode.Target = l.nodeStack.Pop()
	l.nodeStack.Push(assignNode)
}

func (l *listener) EnterCompExp(c *parser.CompExpContext) {
	DebugPrintln("Enter comp exp")
}

func (l *listener) ExitCompExp(c *parser.CompExpContext) {
	DebugPrintln("exit comp exp")

	compNode := &ast.CompNode{}
	compNode.Op = c.GetOp().GetText()
	compNode.Right = l.nodeStack.Pop()
	compNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(compNode)
}

func (l *listener) EnterBoolExp(c *parser.BoolExpContext) {
	DebugPrintln("Entering bool literal")
}

func (l *listener) ExitBoolExp(c *parser.BoolExpContext) {
	DebugPrintln("Exiting bool literal")

	boolExp := &ast.BoolExp{}
	boolExp.Value = c.GetText() == "true"

	l.nodeStack.Push(boolExp)
}

func (l *listener) EnterNullExp(c *parser.NullExpContext) {
	DebugPrintln("Entering null literal")
}

func (l *listener) ExitNullExp(c *parser.NullExpContext) {
	DebugPrintln("Exiting null literal")

	l.nullNo++
	nullExp := &ast.NullExp{}
	nullExp.NullID = l.nullNo

	l.nodeStack.Push(nullExp)
}

func (l *listener) EnterByteExp(c *parser.ByteExpContext) {
	DebugPrintln("Entering byte literal")
}

func (l *listener) ExitByteExp(c *parser.ByteExpContext) {
	DebugPrintln("Exiting byte literal")

	byteExp := &ast.ByteExp{}
	byteStr := c.GetText()
	byteExp.Value = byte(byteStr[1])

	l.nodeStack.Push(byteExp)
}

func (l *listener) EnterFloatExp(c *parser.FloatExpContext) {
	DebugPrintln("Entering float literal")
}

func (l *listener) ExitFloatExp(c *parser.FloatExpContext) {
	DebugPrintln("Exiting float literal")

	var err error
	floatExp := &ast.FloatExp{}
	floatExp.Value, err = strconv.ParseFloat(c.GetText(), 64)
	if err != nil {
		panic("error parsing float: " + err.Error())
	}

	l.nodeStack.Push(floatExp)
}

func (l *listener) EnterArray(c *parser.ArrayContext) {
	DebugPrintln("Entering array literal")
}

func (l *listener) ExitArray(c *parser.ArrayContext) {
	DebugPrintln("Exiting array literal")

	newArr := &ast.ArrayLiteral{}
	newArr.Length = len(filterCommas(c.GetElems().GetChildren()))

	for i := 0; i < newArr.Length; i++ {
		newArr.Exprs = append([]ast.Node{l.nodeStack.Pop()}, newArr.Exprs...)
	}

	if len(newArr.Exprs) == 0 {
		l.emptyArrNo++
		newArr.EmptyNo = l.emptyArrNo
	} else {
		newArr.EmptyNo = -1
	}

	l.nodeStack.Push(newArr)
}

func (l *listener) EnterTuple(c *parser.TupleContext) {
	DebugPrintln("Entering tuple")
}

func (l *listener) ExitTuple(c *parser.TupleContext) {
	DebugPrintln("Exiting tuple")

	newTup := &ast.TupleLiteral{}
	elemCount := len(filterCommas(c.GetElems().GetChildren()))

	for i := 0; i < elemCount; i++ {
		newTup.Exprs = append([]ast.Node{l.nodeStack.Pop()}, newTup.Exprs...)
	}

	l.nodeStack.Push(newTup)
}

func (l *listener) EnterSliceExp(c *parser.SliceExpContext) {
	DebugPrintln("Entering slice exp")
}

func (l *listener) ExitSliceExp(c *parser.SliceExpContext) {
	DebugPrintln("Exiting slice exp")

	sliceNode := &ast.SliceNode{}
	sliceNode.Index = l.nodeStack.Pop()
	sliceNode.Arr = l.nodeStack.Pop()

	l.nodeStack.Push(sliceNode)
}

func (l *listener) EnterCommandExp(c *parser.CommandExpContext) {
	DebugPrintln("Entering command exp")
}

func (l *listener) ExitCommandExp(c *parser.CommandExpContext) {
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

func (l *listener) EnterPipeExp(c *parser.PipeExpContext) {
	DebugPrintln("Entering pipe exp")
}

func (l *listener) ExitPipeExp(c *parser.PipeExpContext) {
	DebugPrintln("Exiting pipe exp")

	pipeNode := &ast.PipeExp{}
	pipeNode.Right = l.nodeStack.Pop()
	pipeNode.Left = l.nodeStack.Pop()

	l.nodeStack.Push(pipeNode)
}

func (l *listener) EnterStrExp(c *parser.StrExpContext) {
	DebugPrintln("Entering string")
}

func (l *listener) ExitStrExp(c *parser.StrExpContext) {
	DebugPrintln("Exiting string", c.GetText())
	text := c.GetText()[1 : len(c.GetText())-1]
	l.nodeStack.Push(&ast.StrExp{text, l.NewNodeID()})
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

func ParseProgram(text string) *ast.Program {
	is := antlr.NewInputStream(text)
	lexer := parser.NewDandelionLex(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewDandelion(stream)

	l := &listener{}
	l.typeStack = &TypeStack{}
	l.prog = ast.NewProgram()
	antlr.ParseTreeWalkerDefault.Walk(l, p.Start())
	l.prog.CurrNodeID = l.nodeID + 1

	return l.prog
}
