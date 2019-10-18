package ast

import (
	"ahead/types"
	"fmt"
	"strings"
)

type Program struct {
	Funcs    map[string]*FunDef
	Structs  map[string]*StructDef
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

type AddSub struct {
	Left  Node
	Right Node
	Op    string
}

func (n *AddSub) String() string {
	return fmt.Sprintf("%v %s %v", n.Left, n.Op, n.Right)
}

type Mod struct {
	Left  Node
	Right Node
}

func (n *Mod) String() string {
	return fmt.Sprintf("%v %% %v", n.Left, n.Right)
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
	Target Node
	Expr   Node
}

func (n *Assign) String() string {
	return fmt.Sprintf("%v = %v", n.Target, n.Expr)
}

type Ident struct {
	Value string
}

func (n *Ident) String() string {
	return n.Value
}

type FunDef struct {
	Body *Block
	Args []Node
	Type types.Type
}

func NewFunDef() *FunDef {
	newFun := &FunDef{}
	newFun.Args = make([]Node, 0)

	return newFun
}

func (n *FunDef) String() string {
	lines := "f"

	argStrings := make([]string, 0)
	for _, arg := range n.Args {
		argStrings = append(argStrings, arg.(*Ident).Value)
	}

	if len(n.Args) > 0 {
		lines += "(" + strings.Join(argStrings, ",") + ")"
	}
	lines += "{\n"
	lines += n.Body.String()

	lines += "}"
	return lines
}

type StructMember struct {
	Name     *Ident
	TypeName *Ident
}

func (n *StructMember) String() string {
	return fmt.Sprintf("%s %s", n.TypeName, n.Name)
}

type StructDef struct {
	Members []*StructMember
}

func (n *StructDef) String() string {
	members := make([]string, 0)
	for _, member := range n.Members {
		members = append(members, "    "+member.String())
	}

	return fmt.Sprintf("struct {\n%s\n}", strings.Join(members, "\n"))
}

type StructInstance struct {
	Name          string
	DefaultValues map[*Ident]Node
}

func (n *StructInstance) String() string {
	return fmt.Sprintf("<struct instance '%s'>", n.Name)
}

type StructAccess struct {
	Field  Node // Must be ident
	Target Node
}

func (n *StructAccess) String() string {
	return fmt.Sprintf("%s.%s", n.Target, n.Field)
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

type PipeExp struct {
	Left  Node
	Right Node
}

func (n *PipeExp) String() string {
	return fmt.Sprintf("%v -> %v", n.Left, n.Right)
}

type CommandExp struct {
	Command string
	Args    []string
}

func (n *CommandExp) String() string {
	return fmt.Sprintf("`%s`", n.Command)
}

type ReturnExp struct {
	Target Node
}

func (n *ReturnExp) String() string {
	return fmt.Sprintf("return %s", n.Target)
}

type YieldExp struct {
	Target Node
}

func (n *YieldExp) String() string {
	return fmt.Sprintf("yield %s", n.Target)
}
