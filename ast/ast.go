package ast

import (
	"fmt"
	"reflect"
	"strings"
)

func ApplyFunc(astNode Node, fun func(Node) Node) Node {
	result := fun(astNode)
	if result != nil {
		return result
	}
	var retVal Node

	switch node := astNode.(type) {
	case *Assign:
		retVal = &Assign{node.Ident, ApplyFunc(node.Expr, fun)}
	case *Num:
		retVal = node
	case *Ident:
		retVal = node
	case *AddSub:
		retVal = &AddSub{ApplyFunc(node.Left, fun), ApplyFunc(node.Right, fun), node.Op}
	case *MulDiv:
		retVal = &MulDiv{ApplyFunc(node.Left, fun), ApplyFunc(node.Right, fun), node.Op}
	case *FunDef:
		newBlock := &Block{ApplyBlock(node.Body.Lines, fun)}
		retVal = &FunDef{newBlock, node.Args}
	case *FunApp:
		retVal = &FunApp{ApplyFunc(node.Fun, fun), ApplyBlock(node.Args, fun)}
	case *While:
		retVal = &While{ApplyFunc(node.Cond, fun), &Block{ApplyBlock(node.Body.Lines, fun)}}
	case *If:
		retVal = &If{ApplyFunc(node.Cond, fun), &Block{ApplyBlock(node.Body.Lines, fun)}}
	case *CompNode:
		retVal = &CompNode{node.Op, ApplyFunc(node.Left, fun), ApplyFunc(node.Right, fun)}
	case *ArrayLiteral:
		retVal = &ArrayLiteral{node.Length, ApplyBlock(node.Exprs, fun)}
	case *SliceNode:
		retVal = &SliceNode{ApplyFunc(node.Index, fun), ApplyFunc(node.Arr, fun)}
	case *StrExp:
		retVal = node
	default:
		panic("ApplyFunc not defined for type: " + reflect.TypeOf(astNode).String())
	}

	return retVal
}

func ApplyBlock(lines []Node, fun func(Node) Node) []Node {
	newLines := make([]Node, 0)
	for _, line := range lines {
		newLines = append(newLines, ApplyFunc(line, fun))
	}

	return newLines
}

type Program struct {
	Funcs    []*FunDef
	MainFunc *FunDef
	Output   string
}

type Block struct {
	Lines []Node
}

func (b *Block) String() string {
	lines := ""

	for _, expr := range b.Lines {
		exprLines := strings.Split(fmt.Sprintf("%v", expr), "\n")
		for _, line := range exprLines {
			lines += fmt.Sprintf("    %v\n", line)
		}
	}

	return lines
}

type Node interface {
}

type Line interface {
}

type AddSub struct {
	Left  Node
	Right Node
	Op    string
}

func (n *AddSub) String() string {
	return fmt.Sprintf("%v %s %v", n.Left, n.Op, n.Right)
}

type MulDiv struct {
	Left  Node
	Right Node
	Op    string
}

func (n *MulDiv) String() string {
	return fmt.Sprintf("%v %s %v", n.Left, n.Op, n.Right)
}

type Num struct {
	Value string
}

func (n *Num) String() string {
	return fmt.Sprintf("%s", n.Value)
}

type Assign struct {
	Ident string
	Expr  Node
}

func (n *Assign) String() string {
	return fmt.Sprintf("%s = %v", n.Ident, n.Expr)
}

type Ident struct {
	Value string
}

func (n *Ident) String() string {
	return n.Value
}

type FunDef struct {
	Body *Block
	Args []string
}

func NewFunDef() *FunDef {
	newFun := &FunDef{}
	newFun.Args = make([]string, 0)

	return newFun
}

func (n *FunDef) String() string {
	lines := "f"
	if len(n.Args) > 0 {
		lines += "(" + strings.Join(n.Args, ",") + ")"
	}
	lines += "{\n"
	lines += n.Body.String()

	lines += "}"
	return lines
}

type FunApp struct {
	Fun  Node
	Args []Node
}

func (n *FunApp) String() string {
	argStrings := make([]string, 0)
	for _, arg := range n.Args {
		argStrings = append(argStrings, fmt.Sprintf("%v", arg))
	}
	return fmt.Sprintf("%v(%s)", n.Fun, strings.Join(argStrings, ", "))
}

type While struct {
	Cond Node
	Body *Block
}

func (n *While) String() string {
	lines := fmt.Sprintf("while %v {\n", n.Cond)
	lines += n.Body.String()
	lines += "}"

	return lines
}

type If struct {
	Cond Node
	Body *Block
}

func (n *If) String() string {
	lines := fmt.Sprintf("if %v {\n", n.Cond)
	lines += n.Body.String()
	lines += "}"

	return lines
}

type CompNode struct {
	Op    string
	Left  Node
	Right Node
}

func (n *CompNode) String() string {
	return fmt.Sprintf("%v %s %v", n.Left, n.Op, n.Right)
}

type ArrayLiteral struct {
	Length int
	Exprs  []Node
}

func (n *ArrayLiteral) String() string {
	arrStr := "["

	exprStrings := make([]string, 0)
	for _, expr := range n.Exprs {
		exprStrings = append(exprStrings, fmt.Sprintf("%v", expr))
	}

	arrStr += strings.Join(exprStrings, ", ") + "]"
	return arrStr
}

type SliceNode struct {
	Index Node
	Arr   Node
}

func (n *SliceNode) String() string {
	return fmt.Sprintf("%v[%v]", n.Arr, n.Index)
}

type StrExp struct {
	Value string
}

func (n *StrExp) String() string {
	return fmt.Sprintf("\"%s\"", n.Value)
}
