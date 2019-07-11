package main

type Func struct {
	lines []Line
	args  []Expr
}

type Line interface {
	Visit() interface{}
}

type Expr interface{}

type BinOp struct {
	left  Expr
	right Expr
	op    string
}
