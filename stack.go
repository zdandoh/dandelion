package main

import (
	"fmt"
	"strings"
)

type NodeStack struct {
	arr []AstNode
}

func (s *NodeStack) Push(node AstNode) {
	s.arr = append(s.arr, node)
}

func (s *NodeStack) Pop() AstNode {
	last := s.arr[len(s.arr)-1]
	s.arr = s.arr[:len(s.arr)-1]

	return last
}

func (s *NodeStack) Peek() AstNode {
	return s.arr[len(s.arr)-1]
}

func (s *NodeStack) Clear() {
	s.arr = make([]AstNode, 0)
}

type Block struct {
	lines []AstNode
}

func (b *Block) String() string {
	lines := ""

	for _, expr := range b.lines {
		exprLines := strings.Split(fmt.Sprintf("%v", expr), "\n")
		for _, line := range exprLines {
			lines += fmt.Sprintf("    %v\n", line)
		}
	}

	return lines
}

type BlockStack struct {
	arr []*Block
	Top *Block
}

func (s *BlockStack) Push(block *Block) {
	s.arr = append(s.arr, block)
	s.Top = block
}

func (s *BlockStack) Pop() *Block {
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
	s.arr = make([]*Block, 0)
	s.Top = nil
}
