package main

import (
	"fmt"
	"strings"
)

type AstNode interface {
}

type Line interface {
}

type AddSub struct {
	left  AstNode
	right AstNode
	op    string
}

func (n *AddSub) String() string {
	return fmt.Sprintf("%v %s %v", n.left, n.op, n.right)
}

type MulDiv struct {
	left  AstNode
	right AstNode
	op    string
}

func (n *MulDiv) String() string {
	return fmt.Sprintf("%v %s %v", n.left, n.op, n.right)
}

type Num struct {
	value string
}

func (n *Num) String() string {
	return fmt.Sprintf("%s", n.value)
}

type Assign struct {
	ident string
	expr  AstNode
}

func (n *Assign) String() string {
	return fmt.Sprintf("%s = %v", n.ident, n.expr)
}

type Ident struct {
	value string
}

func (n *Ident) String() string {
	return n.value
}

type FunDef struct {
	body *Block
	args []string
}

func NewFunDef() *FunDef {
	newFun := &FunDef{}
	newFun.args = make([]string, 0)

	return newFun
}

func (n *FunDef) String() string {
	lines := "f"
	if len(n.args) > 0 {
		lines += "(" + strings.Join(n.args, ",") + ")"
	}
	lines += "{\n"
	lines += n.body.String()

	lines += "}"
	return lines
}

type FunApp struct {
	fun  AstNode
	args []AstNode
}

func (n *FunApp) String() string {
	argStrings := make([]string, 0)
	for _, arg := range n.args {
		argStrings = append(argStrings, fmt.Sprintf("%v", arg))
	}
	return fmt.Sprintf("%v(%s)", n.fun, strings.Join(argStrings, ", "))
}

type While struct {
	cond AstNode
	body *Block
}

func (n *While) String() string {
	lines := fmt.Sprintf("while %v {\n", n.cond)
	lines += n.body.String()
	lines += "}"

	return lines
}

type If struct {
	cond AstNode
	body *Block
}

func (n *If) String() string {
	lines := fmt.Sprintf("if %v {\n", n.cond)
	lines += n.body.String()
	lines += "}"

	return lines
}

type CompNode struct {
	op    string
	left  AstNode
	right AstNode
}

func (n *CompNode) String() string {
	return fmt.Sprintf("%v %s %v", n.left, n.op, n.right)
}

type ArrayLiteral struct {
	length int
	exprs  []AstNode
}

func (n *ArrayLiteral) String() string {
	arrStr := "["

	exprStrings := make([]string, 0)
	for _, expr := range n.exprs {
		exprStrings = append(exprStrings, fmt.Sprintf("%v", expr))
	}

	arrStr += strings.Join(exprStrings, ", ") + "]"
	return arrStr
}

type SliceNode struct {
	index AstNode
	arr   AstNode
}

func (n *SliceNode) String() string {
	return fmt.Sprintf("%v[%v]", n.arr, n.index)
}
