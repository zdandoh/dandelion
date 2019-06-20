package main

import (
	"ahead/parsing/parser"
	"fmt"
	"reflect"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type visitorFunc func(antlr.Tree) interface{}

var visitFuncs = make(map[string]visitorFunc)

type MathListener struct {
	parser.BaseMathVisitor
	lines    []MathOp
	currLine *MathOp
}

func registerFunc(name string, fun visitorFunc) {
	visitFuncs["*parser."+name] = fun
}

func visitMulDiv(node antlr.Tree) interface{} {
	ctx := node.(*parser.MulDivContext)
	//fmt.Println(ctx.GetText())
	return nil
}

func visitNum(node antlr.Tree) interface{} {
	ctx := node.(*parser.IntContext)
	//fmt.Println(ctx.GetText())
	return nil
}

func init() {
	registerFunc("IntContext", visitNum)
	registerFunc("MulDivContext", visitMulDiv)
}

func visitTree(node antlr.Tree) interface{} {
	nodeType := reflect.TypeOf(node).String()
	visitFunc, ok := visitFuncs[nodeType]
	if ok {
		newNode := visitFunc(node)
		if newNode != nil {
			return newNode
		}
	}

	children := node.GetChildren()

	childResults := make([]interface{}, len(children))
	for i := 0; i < len(children); i++ {
		childResults[i] = visitTree(children[0])
	}

	var notNull interface{}
	for _, result := range childResults {
		if result != nil {
			if notNull != nil {
				panic("AST walk error, too many children")
			}
			notNull = result
		}
	}

	return notNull
}

// hugs
func main() {
	is := antlr.NewInputStream("1 * 7 + 32\n")

	lexer := parser.NewMathLexer(is)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewMathParser(tokenStream)

	tree := p.Prog().(antlr.Tree)
	fmt.Println(tree.(*parser.ProgContext).Accept())
	visitTree(tree)
}
