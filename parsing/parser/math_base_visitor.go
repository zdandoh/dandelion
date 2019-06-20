// Code generated from Math.g4 by ANTLR 4.7.2. DO NOT EDIT.

package parser // Math
import "github.com/antlr/antlr4/runtime/Go/antlr"

type BaseMathVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseMathVisitor) VisitProg(ctx *ProgContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMathVisitor) VisitLine(ctx *LineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMathVisitor) VisitParens(ctx *ParensContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMathVisitor) VisitMulDiv(ctx *MulDivContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMathVisitor) VisitAddSub(ctx *AddSubContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMathVisitor) VisitInt(ctx *IntContext) interface{} {
	return v.VisitChildren(ctx)
}
