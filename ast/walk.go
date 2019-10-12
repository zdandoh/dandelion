package ast

import (
	"reflect"
)

type AstWalker interface {
	WalkNode(Node) Node
	WalkBlock(*Block) *Block
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
	case *Assign:
		retVal = &Assign{WalkAst(node.Target, w), WalkAst(node.Expr, w)}
	case *Num:
		retVal = node
	case *Ident:
		retVal = node
	case *AddSub:
		retVal = &AddSub{WalkAst(node.Left, w), WalkAst(node.Right, w), node.Op}
	case *PipeExp:
		retVal = &PipeExp{WalkAst(node.Left, w), WalkAst(node.Right, w)}
	case *CommandExp:
		retVal = &CommandExp{node.Command, node.Args}
	case *MulDiv:
		retVal = &MulDiv{WalkAst(node.Left, w), WalkAst(node.Right, w), node.Op}
	case *Mod:
		retVal = &Mod{WalkAst(node.Left, w), WalkAst(node.Right, w)}
	case *FunDef:
		walkedArgs := WalkList(node.Args, w)
		newBlock := WalkBlock(node.Body, w)
		retVal = &FunDef{newBlock, walkedArgs}
	case *FunApp:
		walkedArgs := WalkList(node.Args, w)
		retVal = &FunApp{WalkAst(node.Fun, w), walkedArgs}
	case *While:
		retVal = &While{WalkAst(node.Cond, w), WalkBlock(node.Body, w)}
	case *StructInstance:
		newDefaults := make(map[*Ident]Node)
		for key, value := range node.DefaultValues {
			newDefaults[WalkAst(key, w).(*Ident)] = WalkAst(value, w)
		}

		retVal = &StructInstance{node.Name, newDefaults}
	case *StructDef:
		retVal = &StructDef{Members: node.Members}
	case *StructAccess:
		retVal = &StructAccess{WalkAst(node.Field, w), WalkAst(node.Target, w)}
	case *If:
		retVal = &If{WalkAst(node.Cond, w), WalkBlock(node.Body, w)}
	case *ReturnExp:
		retVal = &ReturnExp{WalkAst(node.Target, w)}
	case *YieldExp:
		retVal = &YieldExp{WalkAst(node.Target, w)}
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
		newLines = append(newLines, WalkAst(line, w))
	}
	newBlock.Lines = newLines

	return newBlock
}
