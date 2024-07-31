package asm

import "fmt"

type Expr interface {
}

type ExprBinary struct {
	Left     Expr
	Right    Expr
	Operator Token
}

func (exp ExprBinary) String() string {
	return fmt.Sprintf("%s %s %s", exp.Left, exp.Right, exp.Operator)
}

type ExprUnary struct {
	Expression Expr
	Operator   Token
}

func (exp ExprUnary) String() string {
	return fmt.Sprintf("%s %s", exp.Expression, exp.Operator)
}

type ExprLiteral struct {
	Operator Token
}

func (exp ExprLiteral) String() string {
	return exp.Operator.String()
}
