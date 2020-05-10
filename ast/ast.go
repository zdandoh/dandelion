package ast

import (
	"bytes"
	"crypto/sha256"
	"dandelion/types"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/llir/llvm/ir/enum"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

func init() {
	gob.Register(Ident{})
	gob.Register(AddSub{})
	gob.Register(MulDiv{})
	gob.Register(Num{})
	gob.Register(StructDef{})
	gob.Register(BoolExp{})
	gob.Register(StrExp{})
	gob.Register(Assign{})
	gob.Register(ReturnExp{})
	gob.Register(YieldExp{})
	gob.Register(FloatExp{})
	gob.Register(FunApp{})
	gob.Register(FunDef{})
	gob.Register(While{})
	gob.Register(If{})
	gob.Register(Mod{})
	gob.Register(CompNode{})
	gob.Register(TupleLiteral{})
	gob.Register(ArrayLiteral{})
	gob.Register(SliceNode{})
	gob.Register(StructDef{})
	gob.Register(StructAccess{})
	gob.Register(StructInstance{})
	gob.Register(Closure{})
	gob.Register(ParenExp{})
	gob.Register(NullExp{})
	gob.Register(Pipeline{})
	gob.Register(BlockExp{})
	gob.Register(For{})
	gob.Register(FlowControl{})
	gob.Register(BuiltinExp{})
	gob.Register(TypeAssert{})
	gob.Register(IsExp{})
	gob.Register(ForIter{})
	gob.Register(PipeExp{})
}

type NodeID int

var NoID NodeID = -1

func (n NodeID) ID() NodeID {
	return n
}

type BuiltinName string

const (
	BuiltinLen  BuiltinName = "len"
	BuiltinNext BuiltinName = "next"
	BuiltinSend BuiltinName = "send"
	BuiltinAny  BuiltinName = "any"
	BuiltinDone BuiltinName = "done"
	BuiltinType BuiltinName = "type"
)

var BuiltinArgs = map[BuiltinName]int{
	BuiltinLen:  1,
	BuiltinNext: 1,
	BuiltinSend: 2,
	BuiltinAny:  1,
	BuiltinDone: 1,
	BuiltinType: 1,
}

type Program struct {
	Funcs       map[string]*FunDef
	structs     map[string]*StructDef
	structOrder []*StructDef
	Metadata    map[NodeID]*Meta
	RefTypes    map[types.TypeHash]types.Type // Types that are referenced in the program, even if no expression has that type
	CurrNodeID  NodeID
	Output      string
}

func NewProgram() *Program {
	newProg := &Program{}
	newProg.Funcs = make(map[string]*FunDef)
	newProg.structs = make(map[string]*StructDef)
	newProg.Metadata = make(map[NodeID]*Meta)
	newProg.RefTypes = make(map[types.TypeHash]types.Type)

	return newProg
}

func (p *Program) Struct(name string) *StructDef {
	for _, sDef := range p.structs {
		if sDef.Type.Name == name {
			return sDef
		}
	}

	return nil
}

func (p *Program) StructNo(index int) *StructDef {
	return p.structOrder[index]
}

func (p *Program) StructCount() int {
	return len(p.structs)
}

func (p *Program) AddStruct(name string, newStruct *StructDef) {
	p.structs[name] = newStruct
	p.structOrder = append(p.structOrder, newStruct)
}

func (p *Program) Meta(node Node) *Meta {
	return p.Metadata[node.ID()]
}

func (p *Program) NewNodeID() NodeID {
	p.CurrNodeID++

	newMeta := &Meta{}
	p.Metadata[p.CurrNodeID] = newMeta

	return p.CurrNodeID
}

type Meta struct {
	LineNo int
	Hint   types.Type
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
	String() string
	ID() NodeID
}

type AddSub struct {
	Left  Node
	Right Node
	Op    string
	NodeID
}

func (n *AddSub) String() string {
	return fmt.Sprintf("%v %s %v", n.Left, n.Op, n.Right)
}

type Mod struct {
	Left  Node
	Right Node
	NodeID
}

func (n *Mod) String() string {
	return fmt.Sprintf("%v %% %v", n.Left, n.Right)
}

type ParenExp struct {
	Exp Node
	NodeID
}

func (n *ParenExp) String() string {
	return fmt.Sprintf("(%v)", n.Exp)
}

type MulDiv struct {
	Left  Node
	Right Node
	Op    string
	NodeID
}

func (n *MulDiv) String() string {
	return fmt.Sprintf("%v %s %v", n.Left, n.Op, n.Right)
}

type Num struct {
	Value int64
	NodeID
}

func (n *Num) String() string {
	return fmt.Sprintf("%d", n.Value)
}

// TODO create an "Assignable" interface so you don't always need to do a type switch on the target
type Assign struct {
	Target Node
	Expr   Node
	NodeID
}

func (n *Assign) String() string {
	return fmt.Sprintf("%v = %v", n.Target, n.Expr)
}

type IsExp struct {
	CheckNode Node
	CheckType types.Type
	NodeID
}

func (n *IsExp) String() string {
	return fmt.Sprintf("%s is %s", n.CheckNode, n.CheckType.TypeString())
}

type Ident struct {
	Value string
	NodeID
}

func (n *Ident) String() string {
	return n.Value
}

type FunDef struct {
	Body     *Block
	Args     []Node
	TypeHint *types.FuncType
	IsCoro   *bool
	NodeID
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
		if n.TypeHint != nil && len(n.TypeHint.ArgTypes) > i {
			argString = fmt.Sprintf("%s %s", n.TypeHint.ArgTypes[i].TypeString(), argString)
		}
		argStrings = append(argStrings, argString)
	}

	if len(n.Args) > 0 {
		lines += "(" + strings.Join(argStrings, ",") + ")"
	}
	if n.TypeHint != nil {
		lines += " " + n.TypeHint.RetType.TypeString() + " "
	}

	lines += "{\n"
	lines += n.Body.String()

	lines += "}"
	return lines
}

type StructMember struct {
	Name *Ident
	Type types.Type
	NodeID
}

type StructMethod struct {
	Name       string
	TargetName string
}

func (n *StructMember) String() string {
	return fmt.Sprintf("%s %s", n.Type.TypeString(), n.Name)
}

type StructDef struct {
	Members []*StructMember
	Methods []*StructMethod // Methods are discovered during function removal
	Type    types.StructType
	NodeID
}

func (d *StructDef) Method(name string) *StructMethod {
	for _, method := range d.Methods {
		if method.Name == name {
			return method
		}
	}

	return nil
}

func (d *StructDef) HasMethod(name string) bool {
	if d.Methods == nil {
		return false
	}

	for _, method := range d.Methods {
		if method.Name == name {
			return true
		}
	}

	return false
}

func (d *StructDef) HasMember(name string) bool {
	for _, member := range d.Members {
		if member.Name.Value == name {
			return true
		}
	}

	return false
}

func (d *StructDef) Has(name string) bool {
	return d.HasMember(name) || d.HasMethod(name)
}

func (n *StructDef) String() string {
	members := make([]string, 0)
	for _, member := range n.Members {
		members = append(members, "    "+member.String())
	}
	for _, method := range n.Methods {
		members = append(members, fmt.Sprintf("    %s()", method.Name))
	}

	return fmt.Sprintf("struct {\n%s\n}", strings.Join(members, "\n"))
}

func (n *StructDef) MemberType(memberName string) types.Type {
	for _, member := range n.Members {
		if member.Name.Value == memberName {
			return member.Type
		}
	}

	panic("Unknown member name: " + memberName)
}

func (n *StructDef) Offset(offsetName string) int {
	structOffset := -1
	for i, member := range n.Members {
		if member.Name.Value == offsetName {
			structOffset = i
			break
		}
	}

	return structOffset
}

type LineBundle struct {
	Lines []Node
	NodeID
}

func (n *LineBundle) String() string {
	return "__UNRESOLVED_LINE_BUNDLE__"
}

type TypeAssert struct {
	Target     Node
	TargetType types.Type
	NodeID
}

func (n *TypeAssert) String() string {
	return fmt.Sprintf("%s.(%s)", n.Target, n.TargetType.TypeString())
}

type Closure struct {
	Target  Node
	ArgTup  Node
	NewFunc Node
	Unbound []string
	NodeID
}

func (n *Closure) String() string {
	return fmt.Sprintf("<closure of '%v'>", n.Target)
}

type StructInstance struct {
	Values []Node
	DefRef *StructDef
	NodeID
}

func (n *StructInstance) String() string {
	memberVals := make([]string, 0)
	for _, member := range n.Values {
		memberVals = append(memberVals, fmt.Sprintf("    %s", member))
	}

	return fmt.Sprintf("struct instance {\n%s\n}", strings.Join(memberVals, "\n"))
}

type StructAccess struct {
	Field  Node // Must be ident
	Target Node
	NodeID
}

func (n *StructAccess) String() string {
	return fmt.Sprintf("%s.%s", n.Target, n.Field)
}

type FunApp struct {
	Fun    Node
	Args   []Node
	Extern bool
	NodeID
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
	NodeID
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
	NodeID
}

func (n *If) String() string {
	lines := fmt.Sprintf("if %v {\n", n.Cond)
	lines += n.Body.String()
	lines += "}"

	return lines
}

type For struct {
	Init Node
	Cond Node
	Step Node
	Body *Block
	NodeID
}

func (n *For) String() string {
	lines := fmt.Sprintf("for %v; %v; %v {\n", n.Init, n.Cond, n.Step)
	lines += "    " + n.Body.String()
	lines += "    }"

	return lines
}

type ForIter struct {
	Item Node
	Iter Node
	Body *Block
	NodeID
}

func (n *ForIter) String() string {
	lines := fmt.Sprintf("for %s in %s {\n", n.Item, n.Iter)
	lines += n.Body.String()
	lines += "}"

	return lines
}

type CompNode struct {
	Op    string
	Left  Node
	Right Node
	NodeID
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
	// Empty arrays are the only ast node that don't have a distinct type. We need to number each empty
	// array to distinguish them for the type checker.
	EmptyNo int
	NodeID
}

func (n *ArrayLiteral) String() string {
	arrStr := "["

	exprStrings := make([]string, 0)
	for _, expr := range n.Exprs {
		exprStrings = append(exprStrings, fmt.Sprintf("%v", expr))
	}

	arrStr += strings.Join(exprStrings, ", ") + "]"

	if n.EmptyNo > 0 {
		arrStr += fmt.Sprintf("#%d", n.EmptyNo)
	}

	return arrStr
}

type BuiltinExp struct {
	Args []Node
	Type BuiltinName
	NodeID
}

func (s *BuiltinExp) String() string {
	argStrings := make([]string, 0)
	for _, arg := range s.Args {
		argStrings = append(argStrings, arg.String())
	}

	return fmt.Sprintf("%s(%s)", s.Type, strings.Join(argStrings, ", "))
}

type TupleLiteral struct {
	Exprs []Node
	NodeID
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
	NodeID
}

func (n *SliceNode) String() string {
	return fmt.Sprintf("%v[%v]", n.Arr, n.Index)
}

type StrExp struct {
	Value string
	NodeID
}

func (n *StrExp) String() string {
	return fmt.Sprintf("\"%s\"", n.Value)
}

type BlockExp struct {
	Block *Block
	NodeID
}

func (n *BlockExp) String() string {
	block := "{\n"
	for _, line := range n.Block.Lines {
		block += "    " + line.String() + "\n"
	}
	block += "}"
	return block
}

type PipeExp struct {
	Left  Node
	Right Node
	NodeID
}

func (n *PipeExp) String() string {
	return fmt.Sprintf("%v -> %v", n.Left, n.Right)
}

type Pipeline struct {
	Ops []Node
	NodeID
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
	NodeID
}

func (n *CommandExp) String() string {
	return fmt.Sprintf("`%s`", n.Command)
}

type ReturnExp struct {
	Target     Node
	SourceFunc string
	NodeID
}

func (n *ReturnExp) String() string {
	return fmt.Sprintf("return %s", n.Target)
}

type YieldExp struct {
	Target     Node
	SourceFunc string
	NodeID
}

func (n *YieldExp) String() string {
	return fmt.Sprintf("yield %s", n.Target)
}

type FlowStatement string

const FlowContinue = "continue"
const FlowBreak = "break"

type FlowControl struct {
	Type FlowStatement
	NodeID
}

func (n *FlowControl) String() string {
	return string(n.Type)
}

type BoolExp struct {
	Value bool
	NodeID
}

func (n *BoolExp) String() string {
	return fmt.Sprintf("%t", n.Value)
}

type NullExp struct {
	NullID int
	NodeID
}

func (n *NullExp) String() string {
	return fmt.Sprintf("null#%d", n.NullID)
}

type ByteExp struct {
	Value byte
	NodeID
}

func (n *ByteExp) String() string {
	return fmt.Sprintf("%d", n.Value)
}

type FloatExp struct {
	Value float64
	NodeID
}

func (n *FloatExp) String() string {
	return fmt.Sprintf("%v", n.Value)
}

type NodeHash string

func HashNode(node Node) NodeHash {
	// Zero the node id to prevent it from affecting the hash
	nodeID := node.ID()
	SetID(node, 0)
	defer SetID(node, nodeID)

	b := bytes.NewBuffer(nil)
	err := gob.NewEncoder(b).Encode(node)
	if err != nil {
		panic(errors.Wrap(err, "failed to hash ast node"))
	}

	hash := sha256.New()
	return NodeHash(hex.EncodeToString(hash.Sum(b.Bytes())))
}

func Statement(node Node) bool {
	switch node.(type) {
	case *Assign:
		return true
	case *ReturnExp:
		return true
	case *YieldExp:
		return true
	case *If:
		return true
	case *While:
		return true
	case *For:
		return true
	case *ForIter:
		return true
	case *BlockExp:
		return true
	case *FlowControl:
		return true
	}

	return false
}

// This is really dumb but I don't know of a better way
func SetID(astNode Node, newID NodeID) {
	switch node := astNode.(type) {
	case *Ident:
		node.NodeID = newID
	case *AddSub:
		node.NodeID = newID
	case *MulDiv:
		node.NodeID = newID
	case *Num:
		node.NodeID = newID
	case *StructDef:
		node.NodeID = newID
	case *BoolExp:
		node.NodeID = newID
	case *StrExp:
		node.NodeID = newID
	case *Assign:
		node.NodeID = newID
	case *ReturnExp:
		node.NodeID = newID
	case *YieldExp:
		node.NodeID = newID
	case *FloatExp:
		node.NodeID = newID
	case *FunApp:
		node.NodeID = newID
	case *FunDef:
		node.NodeID = newID
	case *While:
		node.NodeID = newID
	case *If:
		node.NodeID = newID
	case *Mod:
		node.NodeID = newID
	case *CompNode:
		node.NodeID = newID
	case *TupleLiteral:
		node.NodeID = newID
	case *ArrayLiteral:
		node.NodeID = newID
	case *SliceNode:
		node.NodeID = newID
	case *StructAccess:
		node.NodeID = newID
	case *StructInstance:
		node.NodeID = newID
	case *Closure:
		node.NodeID = newID
	case *ParenExp:
		node.NodeID = newID
	case *NullExp:
		node.NodeID = newID
	case *Pipeline:
		node.NodeID = newID
	case *BlockExp:
		node.NodeID = newID
	case *For:
		node.NodeID = newID
	case *FlowControl:
		node.NodeID = newID
	case *BuiltinExp:
		node.NodeID = newID
	case *TypeAssert:
		node.NodeID = newID
	case *IsExp:
		node.NodeID = newID
	case *ForIter:
		node.NodeID = newID
	case *PipeExp:
		node.NodeID = newID
	default:
		panic("SetID not defined for type:" + reflect.TypeOf(astNode).String())
	}
}
