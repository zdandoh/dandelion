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
		retVal = &ParenExp{WalkAst(node.Exp, w)}
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
	case *Pipeline:
		retVal = &Pipeline{WalkList(node.Ops, w)}
	case *TupleLiteral:
		retVal = &TupleLiteral{WalkList(node.Exprs, w)}
	case *CommandExp:
		retVal = &CommandExp{node.Command, node.Args}
	case *MulDiv:
		retVal = &MulDiv{WalkAst(node.Left, w), WalkAst(node.Right, w), node.Op}
	case *Mod:
		retVal = &Mod{WalkAst(node.Left, w), WalkAst(node.Right, w)}
	case *FunDef:
		walkedArgs := WalkList(node.Args, w)
		newBlock := WalkBlock(node.Body, w)
		retVal = &FunDef{newBlock, walkedArgs, node.TypeHint}
	case *Closure:
		retVal = &Closure{WalkAst(node.Target, w), WalkAst(node.ArgTup, w), WalkAst(node.NewFunc, w), node.Unbound}
	case *FunApp:
		walkedArgs := WalkList(node.Args, w)
		retVal = &FunApp{WalkAst(node.Fun, w), walkedArgs}
	case *While:
		retVal = &While{WalkAst(node.Cond, w), WalkBlock(node.Body, w)}
	case *StructInstance:
		newDefaults := make([]Node, len(node.Values))
		for i, value := range node.Values {
			newDefaults[i] = WalkAst(value, w)
		}

		retVal = &StructInstance{newDefaults, node.DefRef}
	case *StructDef:
		retVal = &StructDef{Members: node.Members}
	case *StructAccess:
		retVal = &StructAccess{&Ident{node.Field.(*Ident).Value}, WalkAst(node.Target, w)}
	case *If:
		retVal = &If{WalkAst(node.Cond, w), WalkBlock(node.Body, w)}
	case *ReturnExp:
		retVal = &ReturnExp{WalkAst(node.Target, w)}
	case *YieldExp:
		retVal = &YieldExp{WalkAst(node.Target, w)}
	case *CompNode:
		retVal = &CompNode{node.Op, WalkAst(node.Left, w), WalkAst(node.Right, w)}
	case *ArrayLiteral:
		retVal = &ArrayLiteral{node.Length, WalkList(node.Exprs, w), node.EmptyNo}
	case *SliceNode:
		retVal = &SliceNode{WalkAst(node.Index, w), WalkAst(node.Arr, w)}
	case *StrExp:
		retVal = node
	case *BoolExp:
		retVal = node
	case *ByteExp:
		retVal = node
	case *FloatExp:
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
