package main

type Expr interface{}

type MathOp struct {
	left  Expr
	right Expr
	op    string
}
