/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
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
	CanReduce() bool     // can this statement be reduced?
	IsFinalReduce() bool // is this a final statement (a jump) in a reduction?
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

func (s StmntDirective) CanReduce() bool {
	return false
}

func (s StmntDirective) IsFinalReduce() bool {
	return false
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

func (s StmntGlobal) CanReduce() bool {
	return false
}

func (s StmntGlobal) IsFinalReduce() bool {
	return false
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

func (s StmntInt) CanReduce() bool {
	return false
}

func (s StmntInt) IsFinalReduce() bool {
	return false
}

type StmntInstr struct {
	Instruction Token
	Args        []Arg
	labels      *map[string]*Label
}

func (s StmntInstr) String() string {
	return fmt.Sprintf("{%s %s}", s.Instruction, s.Args)
}

// Is the instruction reducable? Does not include arguments.
func (s StmntInstr) reducableIns() bool {
	t := s.Instruction.TokenType
	if t.IsInstruction() {
		if t == token.Jumpr || t == token.Jumps {
			return false // jumpr and jumps can't reduce
		}
		return true // other instructions can reduce
	}
	return false // all other token types can't reduce
}

func (s StmntInstr) CanReduce() bool {
	if s.reducableIns() {
		// check every argument to see if it uses a relative
		for _, arg := range s.Args {
			if arg.IsRelative() { // if any are relative
				return false // then no reducing!
			}
		}
		return true
	}
	return false
}

func (s StmntInstr) IsFinalReduce() bool {
	// the only finishing instruction is a definite jump
	if s.Instruction.TokenType == token.Jump {
		isDefiniteJump := len(s.Args) <= 1 // "jump x, ov" and "jump x, eq" can fallthrough
		return isDefiniteJump
	}
	return false
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
	case token.Call:
		return 8
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

func (s StmntLabel) CanReduce() bool {
	return false
}

func (s StmntLabel) IsFinalReduce() bool {
	return false
}

// expressions

type Expr interface {
	Evaluate(map[string]*Label) (int, error)
	IsRelative() bool
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
	case token.RightRight:
		return left >> right, nil
	case token.LeftLeft:
		return left << right, nil
	default:
		return 0, GenericTokenError{e.Operator, "unknown binary token, please file a bug report"}
	}
}

func (exp ExprBinary) String() string {
	return fmt.Sprintf("(%s%s%s)", exp.Left, exp.Operator, exp.Right)
}

func (exp ExprBinary) IsRelative() bool {
	return exp.Left.IsRelative() || exp.Right.IsRelative()
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

func (exp ExprUnary) IsRelative() bool {
	return exp.Expression.IsRelative()
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

func (exp ExprLiteral) IsRelative() bool {
	return exp.Operator.TokenType == token.Here
}

// arguments

type Arg interface {
	IsReg() bool
	IsJump() bool
	IsExpr() bool
	IsRelative() bool
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

func (a ArgReg) IsRelative() bool {
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

func (a ArgJump) IsRelative() bool {
	return false // we can fallthrough and still continue
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

func (a ArgExpr) IsRelative() bool {
	return a.Expr.IsRelative()
}
