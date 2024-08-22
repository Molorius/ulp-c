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

type parser struct {
	tokens   []Token
	position int
}

func (p *parser) parseTokens(tokens []Token) ([]Stmnt, error) {
	p.tokens = tokens
	p.position = 0
	stmnt, err := p.program()
	return stmnt, err
}

func (p *parser) program() ([]Stmnt, error) {
	ret := make([]Stmnt, 0)
	errs := error(nil)
	for {
		if p.match(token.EndOfFile) {
			break
		}
		s, err := p.statement()
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf(""), err)
			continue
		}
		if s != nil {
			ret = append(ret, s)
		}
	}
	return ret, errs
}

func (p *parser) statement() (Stmnt, error) {
	if p.match(token.NewLine) {
		return p.statement()
	}
	if p.match(token.EndOfFile) {
		return nil, nil
	}
	t := p.peak()
	if t.TokenType == token.Identifier {
		return p.label()
	}
	s := Stmnt(nil)
	err := error(nil)
	if t.TokenType.IsDirective() {
		s, err = p.directive()
	} else if t.TokenType.IsInstruction() {
		s, err = p.instruction()
	} else {
		err = GenericTokenError{t, "expected a statement"}
	}

	if err != nil {
		p.nextLine()
	} else {
		// directives and instructions expect a newline/eof
		e := p.consumeEndline()
		if e != nil {
			err = errors.Join(GenericTokenError{t, "error while building statement"}, e)
			s = nil
			p.nextLine()
		}
	}
	return s, err
}

func (p *parser) directive() (Stmnt, error) {
	t := p.next()
	if !t.TokenType.IsDirective() {
		return nil, GenericTokenError{t, "compiler bug in parser.directive(), please file a bug report"}
	}
	switch t.TokenType {
	case token.Global:
		if p.match(token.Identifier) {
			s := StmntGlobal{Label: p.previous()}
			return s, nil
		}
		return nil, ExpectedTokenError{token.Identifier, p.next()}
	case token.Int:
		return p.directiveInt(t)
	default:
		return StmntDirective{t}, nil
	}
}

func (p *parser) directiveInt(t Token) (Stmnt, error) {
	args, err := p.arguments()
	if err != nil {
		return nil, errors.Join(GenericTokenError{t, "could not parse arguments for .int"}, err)
	}
	if len(args) == 0 {
		return nil, GenericTokenError{t, "expected at least 1 argument"}
	}
	argsExpr := make([]ArgExpr, len(args))
	for i := range args {
		switch a := args[i].(type) {
		case ArgExpr:
			argsExpr[i] = a
		default:
			return nil, GenericTokenError{t, fmt.Sprintf("expected an expression on argument %d", i)}
		}
	}
	return StmntInt{argsExpr}, nil
}

func (p *parser) instruction() (Stmnt, error) {
	t := p.next()
	if !t.TokenType.IsInstruction() {
		return nil, GenericTokenError{t, "compiler bug in parser.instruction(), please file a bug report"}
	}
	args, err := p.arguments()
	if err != nil {
		return nil, errors.Join(GenericTokenError{t, "could not parse arguments"}, err)
	}
	s := StmntInstr{
		Instruction: t,
		Args:        args,
	}
	err = s.validate()
	if err != nil {
		return nil, err
	}
	return StmntInstr{
		Instruction: t,
		Args:        args,
	}, nil
}

func (p *parser) label() (Stmnt, error) {
	if !p.match(token.Identifier) {
		return nil, GenericTokenError{p.peak(), "compiler bug in parser.label(), please file a bug report"}
	}
	t := p.previous()
	err := p.consume(token.Colon)
	if err != nil {
		return nil, errors.Join(GenericTokenError{t, "is this supposed to be a label?"}, err)
	}
	return StmntLabel{t}, nil
}

func (p *parser) arguments() ([]Arg, error) {
	args := make([]Arg, 0)
	errs := error(nil)
	if p.isAtEndOfLine() {
		return args, nil
	}
	a, err := p.argument()
	if err != nil {
		return nil, err
	}
	args = append(args, a)
	for p.match(token.Comma) {
		op := p.previous()
		a, err := p.argument()
		if err != nil {
			return nil, errors.Join(UnfinishedError{op, "an argument"}, err)
		}
		args = append(args, a)
	}
	return args, errs
}

func (p *parser) argument() (Arg, error) {
	t := p.peak()
	if t.TokenType.IsRegister() {
		p.advancePointer()
		return ArgReg{t}, nil
	}
	if t.TokenType.IsJump() {
		p.advancePointer()
		return ArgJump{t}, nil
	}

	expr, err := p.Expression()
	if err != nil {
		return nil, err
	}
	return ArgExpr{expr}, nil
}

func (p *parser) Additive() (Expr, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}
	for p.match(token.Minus, token.Plus) {
		op := p.previous()
		f, err := p.factor()
		if err != nil {
			return nil, errors.Join(UnfinishedError{op, "a factor"}, err)
		}
		expr = ExprBinary{
			Left:     expr,
			Operator: op,
			Right:    f,
		}
	}
	return expr, nil
}

func (p *parser) Expression() (Expr, error) {
	expr, err := p.Additive()
	if err != nil {
		return nil, err
	}
	for p.match(token.RightRight, token.LeftLeft) {
		op := p.previous()
		f, err := p.Additive()
		if err != nil {
			return nil, errors.Join(UnfinishedError{op, "an additive"}, err)
		}
		expr = ExprBinary{
			Left:     expr,
			Operator: op,
			Right:    f,
		}
	}
	return expr, nil
}

func (p *parser) factor() (Expr, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.match(token.Slash, token.Star) {
		op := p.previous()
		u, err := p.unary()
		if err != nil {
			return nil, errors.Join(UnfinishedError{op, "a unary"}, err)
		}
		expr = ExprBinary{
			Left:     expr,
			Operator: op,
			Right:    u,
		}
	}
	return expr, nil
}

func (p *parser) unary() (Expr, error) {
	if p.match(token.Minus) {
		op := p.previous()
		u, err := p.unary()
		if err != nil {
			return nil, errors.Join(UnfinishedError{op, "a unary"}, err)
		}
		return ExprUnary{
			Operator:   op,
			Expression: u,
		}, nil
	}
	return p.primary()
}

func (p *parser) primary() (Expr, error) {
	if p.match(token.Number, token.Here, token.Identifier) {
		return ExprLiteral{
			Operator: p.previous(),
		}, nil
	}
	if p.match(token.LeftParen) {
		left := p.previous()
		e, err := p.Expression()
		if err != nil {
			return nil, errors.Join(UnfinishedError{left, "an expression"}, err)
		}
		err = p.consume(token.RightParen)
		if err != nil {
			return nil, errors.Join(
				UnfinishedError{left, "\")\""},
				err,
			)
		}
		return e, nil
	}
	return nil, GenericTokenError{
		token:   p.peak(),
		message: "expected an expression",
	}
}

// below are methods to interact with the token stream

func (p *parser) peak() Token {
	p.fixPointer()
	return p.tokens[p.position]
}

func (p *parser) next() Token {
	t := p.peak()
	p.advancePointer()
	return t
}

func (p *parser) fixPointer() {
	if p.position < 0 {
		p.position = 0
	}
	if p.position >= len(p.tokens) {
		p.position = len(p.tokens) - 1
	}
}

func (p *parser) advancePointer() {
	p.position += 1
	p.fixPointer()
}

func (p *parser) check(t token.Type) bool {
	return p.peak().TokenType == t
}

func (p *parser) isAtEnd() bool {
	return p.check(token.EndOfFile)
}

func (p *parser) consumeEndline() error {
	if p.isAtEnd() {
		return nil
	}
	return p.consume(token.NewLine)
}

func (p *parser) nextLine() {
	for {
		err := p.consumeEndline()
		if err == nil {
			return
		}
		p.advancePointer()
	}
}

func (p *parser) isAtEndOfLine() bool {
	t := p.peak()
	return t.TokenType == token.EndOfFile || t.TokenType == token.NewLine
}

func (p *parser) previous() Token {
	return p.tokens[p.position-1]
}

func (p *parser) match(toks ...token.Type) bool {
	for _, t := range toks {
		if p.check(t) {
			if t != token.EndOfFile {
				p.advancePointer()
			}
			return true
		}
	}
	return false
}

func (p *parser) consume(t token.Type) error {
	if p.match(t) {
		return nil
	}
	return ExpectedTokenError{
		expected: t,
		got:      p.peak(),
	}
}
