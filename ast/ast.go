package ast

import (
	"fmt"
	"reflect"
	"strings"
)

type AstWalker interface {
	WalkNode(Node) Node
	WalkBlock(*Block) *Block
}

func WalkAst(astNode Node, w AstWalker) Node {
	result := w.WalkNode(astNode)
	if result != nil {
		return result
	}
	var retVal Node

	switch node := astNode.(type) {
	case *Assign:
		retVal = &Assign{node.Ident, WalkAst(node.Expr, w)}
	case *Num:
		retVal = node
	case *Ident:
		retVal = node
	case *AddSub:
		retVal = &AddSub{WalkAst(node.Left, w), WalkAst(node.Right, w), node.Op}
	case *MulDiv:
		retVal = &MulDiv{WalkAst(node.Left, w), WalkAst(node.Right, w), node.Op}
	case *FunDef:
		newBlock := WalkBlock(node.Body, w)
		retVal = &FunDef{newBlock, node.Args}
	case *FunApp:
		retVal = &FunApp{WalkAst(node.Fun, w), WalkList(node.Args, w)}
	case *While:
		retVal = &While{WalkAst(node.Cond, w), WalkBlock(node.Body, w)}
	case *If:
		retVal = &If{WalkAst(node.Cond, w), WalkBlock(node.Body, w)}
	case *CompNode:
		retVal = &CompNode{node.Op, WalkAst(node.Left, w), WalkAst(node.Right, w)}
	case *ArrayLiteral:
		retVal = &ArrayLiteral{node.Length, WalkList(node.Exprs, w)}
	case *SliceNode:
		retVal = &SliceNode{WalkAst(node.Index, w), WalkAst(node.Arr, w)}
	case *StrExp:
		retVal = node
	default:
		panic("WalkAst not defined for type: " + reflect.TypeOf(astNode).String())
	}

	return retVal
}

func WalkList(arr []Node, w AstWalker) []Node {
	newArr := make([]Node, 0)

	for _, line := range arr {
		newArr = append(newArr, w.WalkNode(line))
	}

	return newArr
}

func WalkBlock(block *Block, w AstWalker) *Block {
	newBlock := w.WalkBlock(block)
	if newBlock != nil {
		return newBlock
	}

	newBlock = &Block{}
	newLines := make([]Node, 0)
	for _, line := range block.Lines {
		newLines = append(newLines, WalkAst(line, w))
	}
	newBlock.Lines = newLines

	return newBlock
}

type Program struct {
	Funcs    map[string]*FunDef
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
