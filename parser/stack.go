package parser

import (
	"ahead/ast"
	"ahead/types"
)

type NodeStack struct {
	arr []ast.Node
}

func (s *NodeStack) Push(node ast.Node) {
	s.arr = append(s.arr, node)
}

func (s *NodeStack) Pop() ast.Node {
	last := s.arr[len(s.arr)-1]
	s.arr = s.arr[:len(s.arr)-1]

	return last
}

func (s *NodeStack) Peek() ast.Node {
	return s.arr[len(s.arr)-1]
}

func (s *NodeStack) Clear() {
	s.arr = make([]ast.Node, 0)
}

type BlockStack struct {
	arr []*ast.Block
	Top *ast.Block
}

func (s *BlockStack) Push(block *ast.Block) {
	s.arr = append(s.arr, block)
	s.Top = block
}

func (s *BlockStack) Pop() *ast.Block {
	last := s.arr[len(s.arr)-1]
	s.arr = s.arr[:len(s.arr)-1]

	if len(s.arr) > 0 {
		s.Top = s.arr[len(s.arr)-1]
	} else {
		s.Top = nil
	}

	return last
}

func (s *BlockStack) Clear() {
	s.arr = make([]*ast.Block, 0)
	s.Top = nil
}

type TypeStack struct {
	arr []types.Type
}

func (s *TypeStack) Push(t types.Type) {
	s.arr = append(s.arr, t)
}

func (s *TypeStack) Pop() types.Type {
	last := s.arr[len(s.arr)-1]
	s.arr = s.arr[:len(s.arr)-1]

	return last
}

func (s *TypeStack) Peek() types.Type {
	if len(s.arr) > 0 {
		return s.arr[len(s.arr)-1]
	}

	return nil
}
