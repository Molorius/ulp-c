package asm

import (
	"errors"
	"fmt"

	"github.com/Molorius/ulp-c/pkg/asm/token"
)

// statements

type Stmnt interface {
	Size() int // the size of this statement after compilation, in bytes
	Compile(map[string]*Label) ([]byte, error)
}

type StmntDirective struct {
	Directive Token
}

func (s StmntDirective) Size() int {
	return 0
}

func (s StmntDirective) Compile(labels map[string]*Label) ([]byte, error) {
	return nil, nil
}

type StmntGlobal struct {
	Label Token
}

func (s StmntGlobal) Size() int {
	return 0
}

func (s StmntGlobal) Compile(labels map[string]*Label) ([]byte, error) {
	return nil, nil
}

func (s StmntGlobal) String() string {
	return fmt.Sprintf(".global(%s)", s.Label)
}

type StmntInt struct {
	Args []ArgExpr
}

func (s StmntInt) Size() int {
	return len(s.Args) * 4
}

func (s StmntInt) Compile(labels map[string]*Label) ([]byte, error) {
	errs := error(nil)
	out := make([]byte, 0)
	for _, a := range s.Args {
		val, err := a.Expr.Evaluate(labels)
		if err != nil {
			errs = errors.Join(errs, err)
		}
		b := byteInt(val)
		out = append(out, b...)
	}
	return out, errs
}

func (s StmntInt) String() string {
	return fmt.Sprintf("int{%s}", s.Args)
}

type StmntInstr struct {
	Instruction Token
	Args        []Arg
	labels      *map[string]*Label
}

func (s StmntInstr) String() string {
	return fmt.Sprintf("{%s %s}", s.Instruction, s.Args)
}

func (s StmntInstr) Size() int {
	switch s.Instruction.TokenType {
	case token.Jumpr, token.Jumps:
		switch s.Args[2].(ArgJump).Arg.TokenType {
		case token.Eq:
			return 8
		default:
			return 4
		}
	default:
		return 4
	}
}

type StmntLabel struct {
	Label Token
}

func (s StmntLabel) Size() int {
	return 0
}

func (s StmntLabel) Compile(labels map[string]*Label) ([]byte, error) {
	return nil, nil
}

func (s StmntLabel) String() string {
	return fmt.Sprintf("Label(%s)", s.Label)
}

// expressions

type Expr interface {
	Evaluate(map[string]*Label) (int, error)
}

type ExprBinary struct {
	Left     Expr
	Right    Expr
	Operator Token
}

func (e ExprBinary) Evaluate(labels map[string]*Label) (int, error) {
	errs := error(nil)
	left, err := e.Left.Evaluate(labels)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	right, err := e.Right.Evaluate(labels)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	if errs != nil {
		return 0, errs
	}
	switch e.Operator.TokenType {
	case token.Minus:
		return left - right, nil
	case token.Plus:
		return left + right, nil
	case token.Slash:
		return left / right, nil
	case token.Star:
		return left * right, nil
	default:
		return 0, GenericTokenError{e.Operator, "unknown binary token, please file a bug report"}
	}
}

func (exp ExprBinary) String() string {
	return fmt.Sprintf("(%s%s%s)", exp.Left, exp.Operator, exp.Right)
}

type ExprUnary struct {
	Expression Expr
	Operator   Token
}

func (e ExprUnary) Evaluate(labels map[string]*Label) (int, error) {
	val, err := e.Expression.Evaluate(labels)
	if err != nil {
		return 0, err
	}
	switch e.Operator.TokenType {
	case token.Minus:
		return -val, nil
	default:
		return 0, GenericTokenError{e.Operator, "unknown unary token, please file a bug report"}
	}
}

func (exp ExprUnary) String() string {
	return fmt.Sprintf("(%s%s)", exp.Operator, exp.Expression)
}

type ExprLiteral struct {
	Operator Token
}

func (e ExprLiteral) Evaluate(labels map[string]*Label) (int, error) {
	switch e.Operator.TokenType {
	case token.Number:
		return e.Operator.Number, nil
	case token.Identifier:
		l, ok := labels[e.Operator.Lexeme]
		if !ok {
			return 0, UnknownIdentifierError{e.Operator}
		}
		return l.Value / 4, nil
	case token.Here:
		l, ok := labels["."]
		if !ok {
			return 0, GenericTokenError{e.Operator, "the \"Here\" token \".\" not set, please file a bug report"}
		}
		return l.Value / 4, nil
	}
	return 0, nil
}

func (exp ExprLiteral) String() string {
	return exp.Operator.String()
}

// arguments

type Arg interface {
	IsReg() bool
	IsJump() bool
	IsExpr() bool
}

type ArgReg struct {
	Reg Token
}

func (a ArgReg) IsReg() bool {
	return true
}

func (a ArgReg) IsJump() bool {
	return false
}

func (a ArgReg) IsExpr() bool {
	return false
}

func (a ArgReg) ToExpr() ArgExpr {
	i, _ := a.Evaluate()
	t := a.Reg
	t.Number = i
	t.TokenType = token.Number
	return ArgExpr{Expr: ExprLiteral{t}}
}

func (a ArgReg) Evaluate() (int, error) {
	switch a.Reg.TokenType {
	case token.R0:
		return 0, nil
	case token.R1:
		return 1, nil
	case token.R2:
		return 2, nil
	case token.R3:
		return 3, nil
	default:
		return 0, GenericTokenError{a.Reg, "unknown register while evaluating, please file a bug report"}
	}
}

type ArgJump struct {
	Arg Token
}

func (a ArgJump) IsReg() bool {
	return false
}

func (a ArgJump) IsJump() bool {
	return true
}

func (a ArgJump) IsExpr() bool {
	return false
}

type ArgExpr struct {
	Expr Expr
}

func (a ArgExpr) IsReg() bool {
	return false
}

func (a ArgExpr) IsJump() bool {
	return false
}

func (a ArgExpr) IsExpr() bool {
	return true
}
