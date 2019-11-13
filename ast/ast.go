package ast

import (
	"ahead/types"
	"fmt"
	"github.com/llir/llvm/ir/enum"
	"strings"
)

type Program struct {
	Funcs   map[string]*FunDef
	Structs map[string]*StructDef
	Output  string
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

type ParenExp struct {
	Exp Node
}

func (n *ParenExp) String() string {
	return fmt.Sprintf("(%v)", n.Exp)
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
	Value int64
}

func (n *Num) String() string {
	return fmt.Sprintf("%d", n.Value)
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
	Body    *Block
	Args    []Node
	Unbound []Node
	Type    types.FuncType
}

func NewFunDef() *FunDef {
	newFun := &FunDef{}
	newFun.Args = make([]Node, 0)

	return newFun
}

func (n *FunDef) String() string {
	lines := "f"

	argStrings := make([]string, 0)
	for i := 0; i < len(n.Args); i++ {
		argString := n.Args[i].(*Ident).Value
		if len(n.Type.ArgTypes) == len(n.Args) {
			argString = fmt.Sprintf("%s %s", n.Type.ArgTypes[i].TypeString(), argString)
		}
		argStrings = append(argStrings, argString)
	}

	if len(n.Args) > 0 {
		lines += "(" + strings.Join(argStrings, ",") + ")"
	}
	if len(n.Type.ArgTypes) == len(n.Args) {
		lines += " " + n.Type.RetType.TypeString() + " "
	}

	lines += "{\n"
	lines += n.Body.String()

	lines += "}"
	return lines
}

type LineBundle struct {
	Lines []Node
}

func (n *LineBundle) String() string {
	return "__UNRESOLVED_LINE_BUNDLE__"
}

type StructMember struct {
	Name *Ident
	Type types.Type
}

func (n *StructMember) String() string {
	return fmt.Sprintf("%s %s", n.Type.TypeString(), n.Name)
}

type StructDef struct {
	Members []*StructMember
	Type    types.StructType
}

func (n *StructDef) String() string {
	members := make([]string, 0)
	for _, member := range n.Members {
		members = append(members, "    "+member.String())
	}

	return fmt.Sprintf("struct {\n%s\n}", strings.Join(members, "\n"))
}

type Closure struct {
	Name   string
	Target Node
}

func (n *Closure) String() string {
	return fmt.Sprintf("<closure of '%s'>", n.Target)
}

type StructInstance struct {
	Values []Node
	DefRef *StructDef
}

func (n *StructInstance) String() string {
	memberVals := make([]string, 0)
	for _, member := range n.Values {
		memberVals = append(memberVals, fmt.Sprintf("    %s", member))
	}

	return fmt.Sprintf("struct instance {\n%s\n}", strings.Join(memberVals, "\n"))
}

type StructAccess struct {
	Field      Node // Must be ident
	Target     Node
	TargetType types.StructType
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

func (n *CompNode) LLPred() enum.IPred {
	switch n.Op {
	case "<":
		return enum.IPredSLT
	case ">":
		return enum.IPredSGT
	case ">=":
		return enum.IPredSGE
	case "<=":
		return enum.IPredSLE
	case "==":
		return enum.IPredEQ
	case "!=":
		return enum.IPredNE
	default:
		panic("Unsupported CompNode operator")
	}
}

type ArrayLiteral struct {
	Length int
	Exprs  []Node
	Type   types.ArrayType
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

type TupleLiteral struct {
	Exprs []Node
	Type  types.TupleType
}

func (n *TupleLiteral) String() string {
	tupStr := "("

	exprStrings := make([]string, 0)
	for _, expr := range n.Exprs {
		exprStrings = append(exprStrings, fmt.Sprintf("%v", expr))
	}

	tupStr += strings.Join(exprStrings, ", ") + ")"
	return tupStr
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

type Pipeline struct {
	Ops []Node
}

func (n *Pipeline) String() string {
	segStrs := make([]string, 0)

	for _, op := range n.Ops {
		segStrs = append(segStrs, fmt.Sprintf("%v", op))
	}

	return "(" + strings.Join(segStrs, " -> ") + ")"
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
