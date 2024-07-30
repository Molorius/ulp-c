package asm

import (
	"errors"
	"fmt"

	"github.com/Molorius/ulp-c/pkg/asm/token"
)

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

type parser struct {
	tokens   []Token
	position int
}

func (p *parser) parseTokens(tokens []Token) (Expr, error) {
	p.tokens = tokens
	p.position = 0
	exp, err := p.expression()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error while parsing assembly"), err)
	}
	return exp, nil
}

func (p *parser) expression() (Expr, error) {
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
		e, err := p.expression()
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
		message: "Expected an expression.",
	}
}

// below are methods to interact with the token stream

func (p *parser) peak() Token {
	p.fixPointer()
	return p.tokens[p.position]
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

// func (p *parser) isAtEnd() bool {
// 	return p.check(token.EndOfFile)
// }

func (p *parser) previous() Token {
	return p.tokens[p.position-1]
}

func (p *parser) match(toks ...token.Type) bool {
	for _, t := range toks {
		if p.check(t) {
			p.advancePointer()
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
