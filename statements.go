package main

import (
	"fmt"
	"time"
)

func (f *Function) Run() {
	t := time.Now()
	iter := 100000000
	for i := 0; i < iter; i++ {
		f.registers[0] = f.statements[0].Run()
	}
	fmt.Println(float64(time.Since(t).Nanoseconds()) / float64(iter))
}

type IntOpStatement struct {
	a    VarInt
	b    VarInt
	kind TokenType
}

func (s *IntOpStatement) Run() Var {
	switch s.kind {
	case AddOpToken:
		return s.a + s.b
	case SubOpToken:
		return s.a - s.b
	case MultOpToken:
		return s.a * s.b
	case DivideOpToken:
		return s.a / s.b
	case ModOpToken:
		return s.a % s.b
	default:
		panic("invalid op")
	}
}

type FloatOpStatement struct {
	a    VarInt
	b    VarInt
	kind TokenType
}

func (s *FloatOpStatement) Run() Var {
	switch s.kind {
	case AddOpToken:
		return s.a + s.b
	case SubOpToken:
		return s.a - s.b
	case MultOpToken:
		return s.a * s.b
	case DivideOpToken:
		return s.a / s.b
	case ModOpToken:
		return s.a % s.b
	default:
		panic("invalid op")
	}
}
