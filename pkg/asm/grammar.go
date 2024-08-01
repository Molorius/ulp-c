package asm

import "fmt"

// statements

type Stmnt interface {
	Size() int // the size of this statement after compilation, in bytes
	Compile(int, map[string]*Label) ([]byte, error)
}

type StmntDirective struct {
	Directive Token
}

func (s StmntDirective) Size() int {
	return 0
}

func (s StmntDirective) Compile(here int, labels map[string]*Label) ([]byte, error) {
	return nil, nil
}

type StmntGlobal struct {
	Label Token
}

func (s StmntGlobal) Size() int {
	return 0
}

func (s StmntGlobal) Compile(here int, labels map[string]*Label) ([]byte, error) {
	return nil, nil
}

func (s StmntGlobal) String() string {
	return fmt.Sprintf(".global(%s)", s.Label)
}

type StmntInstr struct {
	Instruction Token
	Args        []Arg
}

func (s StmntInstr) Size() int {
	return 4
}

func (s StmntInstr) Compile(here int, labels map[string]*Label) ([]byte, error) {
	return []byte{0, 0, 0, 0}, nil
}

type StmntLabel struct {
	Label Token
}

func (s StmntLabel) Size() int {
	return 0
}

func (s StmntLabel) Compile(here int, labels map[string]*Label) ([]byte, error) {
	return nil, nil
}

func (s StmntLabel) String() string {
	return fmt.Sprintf("Label(%s)", s.Label)
}

// expressions

type Expr interface {
}

type ExprBinary struct {
	Left     Expr
	Right    Expr
	Operator Token
}

func (exp ExprBinary) String() string {
	return fmt.Sprintf("(%s %s %s)", exp.Operator, exp.Left, exp.Right)
}

type ExprUnary struct {
	Expression Expr
	Operator   Token
}

func (exp ExprUnary) String() string {
	return fmt.Sprintf("(%s %s)", exp.Operator, exp.Expression)
}

type ExprLiteral struct {
	Operator Token
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
