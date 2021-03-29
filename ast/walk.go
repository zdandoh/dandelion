package ast

import (
	"reflect"
)

type AstWalker interface {
	WalkNode(Node) Node
	WalkBlock(*Block) *Block
}

type PostAstWalker interface {
	WalkNode(Node) Node
	WalkBlock(*Block) *Block
	PostWalk(Node)
}

type BaseWalker struct {
	WalkN func(Node) Node
	WalkB func(*Block) *Block
}

func (b *BaseWalker) WalkNode(node Node) Node {
	return b.WalkN(node)
}

func (b *BaseWalker) WalkBlock(block *Block) *Block {
	return b.WalkB(block)
}

func WalkAst(astNode Node, w AstWalker) Node {
	result := w.WalkNode(astNode)
	if result != nil {
		return result
	}
	var retVal Node

	switch node := astNode.(type) {
	case *ParenExp:
		retVal = &ParenExp{WalkAst(node.Exp, w), node.NodeID}
	case *Assign:
		retVal = &Assign{WalkAst(node.Target, w), WalkAst(node.Expr, w), node.NodeID}
	case *Num:
		retVal = node
	case *Ident:
		retVal = node
	case *AddSub:
		retVal = &AddSub{WalkAst(node.Left, w), WalkAst(node.Right, w), node.Op, node.NodeID}
	case *PipeExp:
		retVal = &PipeExp{WalkAst(node.Left, w), WalkAst(node.Right, w), node.NodeID}
	case *Pipeline:
		retVal = &Pipeline{WalkList(node.Ops, w), node.NodeID}
	case *TupleLiteral:
		retVal = &TupleLiteral{WalkList(node.Exprs, w), node.NodeID}
	case *CommandExp:
		retVal = &CommandExp{node.Command, node.Args, node.NodeID}
	case *MulDiv:
		retVal = &MulDiv{WalkAst(node.Left, w), WalkAst(node.Right, w), node.Op, node.NodeID}
	case *Mod:
		retVal = &Mod{WalkAst(node.Left, w), WalkAst(node.Right, w), node.NodeID}
	case *FunDef:
		walkedArgs := WalkList(node.Args, w)
		newBlock := WalkBlock(node.Body, w)
		retVal = &FunDef{newBlock, walkedArgs, node.TypeHint, node.IsCoro, node.NodeID}
	case *Closure:
		retVal = &Closure{WalkAst(node.Target, w), WalkAst(node.ArgTup, w), WalkAst(node.NewFunc, w), node.Unbound, node.NodeID}
	case *TypeAssert:
		retVal = &TypeAssert{WalkAst(node.Target, w), node.TargetType, node.NodeID}
	case *FunApp:
		walkedArgs := WalkList(node.Args, w)
		retVal = &FunApp{WalkAst(node.Fun, w), walkedArgs, node.Extern, node.NodeID}
	case *While:
		retVal = &While{WalkAst(node.Cond, w), WalkBlock(node.Body, w), node.NodeID}
	case *For:
		retVal = &For{WalkAst(node.Init, w), WalkAst(node.Cond, w), WalkAst(node.Step, w), WalkBlock(node.Body, w), node.NodeID}
	case *ForIter:
		retVal = &ForIter{WalkAst(node.Item, w), WalkAst(node.Iter, w), WalkBlock(node.Body, w), node.NodeID}
	case *BlockExp:
		retVal = &BlockExp{WalkBlock(node.Block, w), node.NodeID}
	case *BeginExp:
		retVal = &BeginExp{WalkList(node.Nodes, w), node.NodeID}
	case *IsExp:
		retVal = &IsExp{WalkAst(node.CheckNode, w), node.CheckType, node.NodeID}
	case *StructInstance:
		newDefaults := make([]Node, len(node.Values))
		for i, value := range node.Values {
			newDefaults[i] = WalkAst(value, w)
		}

		retVal = &StructInstance{newDefaults, node.DefRef, node.NodeID}
	case *StructDef:
		retVal = &StructDef{node.Members, node.Methods, node.Type, node.NodeID}
	case *StructAccess:
		field := node.Field.(*Ident)
		retVal = &StructAccess{&Ident{field.Value, field.NodeID}, WalkAst(node.Target, w), node.NodeID}
	case *BuiltinExp:
		retVal = &BuiltinExp{WalkList(node.Args, w), node.Type, node.NodeID}
	case *If:
		retVal = &If{WalkAst(node.Cond, w), WalkBlock(node.Body, w), node.NodeID}
	case *ReturnExp:
		retVal = &ReturnExp{WalkAst(node.Target, w), node.SourceFunc, node.NodeID}
	case *YieldExp:
		retVal = &YieldExp{WalkAst(node.Target, w), node.SourceFunc, node.NodeID}
	case *CompNode:
		retVal = &CompNode{node.Op, WalkAst(node.Left, w), WalkAst(node.Right, w), node.NodeID}
	case *ArrayLiteral:
		retVal = &ArrayLiteral{node.Length, WalkList(node.Exprs, w), node.EmptyNo, node.NodeID}
	case *SliceNode:
		retVal = &SliceNode{WalkAst(node.Index, w), WalkAst(node.Arr, w), node.NodeID}
	case *TupleAccess:
		retVal = &TupleAccess{node.Index,WalkAst(node.Tup, w), node.NodeID}
	case *StrExp:
		retVal = node
	case *BoolExp:
		retVal = node
	case *NullExp:
		retVal = node
	case *ByteExp:
		retVal = node
	case *FloatExp:
		retVal = node
	case *FlowControl:
		retVal = node
	default:
		panic("WalkAst not defined for type: " + reflect.TypeOf(astNode).String())
	}

	postWalker, isPost := w.(PostAstWalker)
	if isPost {
		postWalker.PostWalk(astNode)
	}

	return retVal
}

func WalkList(arr []Node, w AstWalker) []Node {
	newArr := make([]Node, 0)

	for _, line := range arr {
		newArr = append(newArr, WalkAst(line, w))
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
		result := WalkAst(line, w)
		bundle, isBundle := result.(*LineBundle)
		if isBundle {
			newLines = append(newLines, bundle.Lines...)
		} else {
			newLines = append(newLines, result)
		}
	}
	newBlock.Lines = newLines

	return newBlock
}
