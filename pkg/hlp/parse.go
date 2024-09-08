/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package hlp

import (
	"errors"
	"fmt"
	"os"

	"github.com/Molorius/ulp-c/pkg/hlp/token"
)

type parser struct {
	tokens   []Token
	position int
}

func (p *parser) parseTokens(tokens []Token) ([]StaticStatement, error) {
	p.tokens = tokens
	p.position = 0
	stmnt, err := p.program()
	return stmnt, err
}

func (p *parser) program() ([]StaticStatement, error) {
	ret := make([]StaticStatement, 0)
	errs := error(nil)

	for {
		if p.endOfStream() {
			break
		}
		switch p.peak().TokenType {
		case token.Semicolon:
			// skip empty statements
			p.advancePointer()
			continue
		case token.EndOfFile:
			// go to next file
			p.advancePointer()
			continue
		}
		// get the next statement
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

func (p *parser) statement() (StaticStatement, error) {
	t := p.peak()
	switch t.TokenType {
	case token.Noreturn, token.Asm, token.Static:
		p.advancePointer()
		return p.function(&t, false)
	case token.Func:
		return p.function(nil, false)
	case token.Extern:
		p.advancePointer()
		t := p.peak()
		switch t.TokenType {
		case token.Func:
			return p.function(nil, true)
		case token.Identifier:
			return p.globalVariable(true)
		default:
			return nil, ExpectedError{t, "function or variable definition without a modifier"}
		}
	case token.Identifier:
		return p.globalVariable(false)
	default:
		p.advancePointer()
		return nil, ExpectedError{t, "function or variable definition"}
	}
}

func (p *parser) function(modifier *Token, extern bool) (StaticStatement, error) {
	if !p.match(token.Func) {
		p.advancePointer()
		return nil, ExpectedError{p.previous(), "\"func\""}
	}
	ident, params, returnSize, err := p.functionHeader()
	if err != nil {
		return nil, err
	}
	noReturn := false
	static := false
	if modifier != nil {
		switch modifier.TokenType {
		case token.Noreturn:
			noReturn = true
		case token.Static:
			static = true
		case token.Asm:
		default:
			return nil, GenericTokenError{*modifier, "unknown modifier, please file a bug report"}
		}
	}

	if modifier != nil && modifier.TokenType != token.Asm {
		return StaticStatementFunction{
			Ident:      ident,
			NoReturn:   noReturn,
			Static:     static,
			Extern:     extern,
			Parameters: params,
			Returns:    returnSize,
		}, nil
	}
	asm, err := p.functionAssemblyBody()
	if err != nil {
		return nil, err
	}
	return StaticStatementAsm{
		Ident:      ident,
		Parameters: params,
		Returns:    returnSize,
		Asm:        asm,
	}, nil
}

func (p *parser) functionAssemblyBody() ([]string, error) {
	out := make([]string, 0)
	if !p.match(token.LeftBrace) {
		return nil, ExpectedError{p.peak(), "a \"{\""}
	}
	for p.match(token.String) {
		s := p.previous()
		out = append(out, s.StringVal)
		if !p.match(token.Semicolon) {
			return nil, ExpectedError{p.peak(), "a semicolon \";\""}
		}
	}
	if !p.match(token.RightBrace) {
		return nil, ExpectedError{p.peak(), "a string or a \"}\""}
	}
	return out, nil
}

func (p *parser) functionHeader() (Token, []Definition, int, error) {
	ident := p.next()
	if ident.TokenType != token.Identifier {
		return Token{}, nil, 0, ExpectedError{ident, "name of a function"}
	}
	params, err := p.functionParameters()
	if err != nil {
		return Token{}, nil, 0, err
	}
	_ = params
	returnSize := p.next()
	if returnSize.TokenType != token.Number {
		return Token{}, nil, 0, ExpectedError{returnSize, "a number for the return amount"}
	}
	return ident, params, returnSize.Number, nil
}

func (p *parser) functionParameter() (Definition, error) {
	ident := p.peak()
	if !p.match(token.Identifier) {
		return nil, ExpectedError{p.peak(), "a parameter name"}
	}
	if p.match(token.At) {
		num := p.peak()
		if !p.match(token.Number) {
			return nil, ExpectedError{p.peak(), "an integer size for the array"}
		}
		return DefinitionArray{
			Ident: ident,
			Size:  num.Number,
		}, nil
	} else {
		t := p.peak()
		switch t.TokenType {
		case token.RightParen, token.Comma:
		default:
			p.advancePointer()
			return nil, ExpectedError{t, "a \")\", \"@\", or \",\""}
		}
		return DefinitionInt{Ident: ident}, nil
	}
}

func (p *parser) functionParameters() ([]Definition, error) {
	defs := make([]Definition, 0)
	if !p.match(token.LeftParen) {
		return nil, ExpectedError{p.peak(), "left paranthese ("}
	}

	for p.match(token.Identifier) {
		p.reversePointer()
		d, err := p.functionParameter()
		if err != nil {
			return nil, err
		}
		defs = append(defs, d)
		if p.match(token.Comma) {
			continue
		}
	}

	if !p.match(token.RightParen) {
		p.advancePointer()
		return nil, ExpectedError{p.previous(), "right paranthese )"}
	}

	return defs, nil
}

func (p *parser) globalVariable(extern bool) (StaticStatement, error) {
	_ = extern
	ident := p.peak()
	if ident.TokenType != token.Identifier {
		p.advancePointer()
		return nil, ExpectedError{ident, "an identifier"}
	}
	p.advancePointer()
	nxt := p.peak()
	gVar := GlobalVar{
		Ident:  ident,
		Array:  false,
		Extern: extern,
		Value:  make([]Primary, 0),
	}

	switch nxt.TokenType {
	case token.Semicolon:
		return nxt, nil
	case token.Equal:
		if extern {
			return nil, ExpectedError{nxt, "a semicolon \";\""}
		}
		p.advancePointer()
		if p.match(token.Number) {
			n := p.previous()
			gVar.Value = append(gVar.Value, n.Number)
		} else if p.match(token.Ampersand) {

		} else {
			return nil, GenericTokenError{p.peak(), "a value"}
		}
	case token.At:
		if !p.match(token.Number) {
			return nil, ExpectedError{p.peak(), "a number for the size of the array"}
		}
		n := p.previous()
		_ = n
		gVar.Array = true

		nxt2 := p.peak()
		switch nxt2.TokenType {
		case token.Semicolon:

		case token.Equal:

		default:
			return nil, ExpectedError{nxt2, "a semicolon or equals"}
		}
	default:
		return nil, UnknownTokenError{nxt}
	}
	return gVar, nil
}

func (p *parser) ArrayValues() []Primary {
	return nil
}

func (p *parser) Primary() (Primary, error) {
	val := p.next()
	addressOf := false
	switch val.TokenType {
	case token.Number:
		return PrimaryNumber{
			N: HlpNumber(val.Number),
		}, nil
	case token.Ampersand:
		addressOf = true
		val = p.next()
		if val.TokenType != token.Identifier {
			return nil, ExpectedError{val, "an identifier"}
		}
		fallthrough
	case token.Identifier:
		offset := 0
		array := false
		nxt := p.peak()
		if nxt.TokenType == token.Pound {
			p.advancePointer()
			n := p.next()
			if n.TokenType != token.Number {
				return nil, ExpectedError{n, "a numerical offset"}
			}
			offset = n.Number
			array = true
		}
		v := Var{
			Ident:     val,
			Array:     array,
			Offset:    offset,
			AddressOf: addressOf,
		}
		return PrimaryVar{v}, nil
	default:
		return nil, ExpectedError{val, "a number or identifier"}
	}
}

func (p *parser) endOfStream() bool {
	return p.position >= len(p.tokens)
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

func (p *parser) check(t token.Type) bool {
	return p.peak().TokenType == t
}

func (p *parser) advancePointer() {
	p.position += 1
	p.fixPointer()
}

func (p *parser) reversePointer() {
	p.position -= 1
	p.fixPointer()
}

func (p *parser) fixPointer() {
	if p.position < 0 {
		p.position = 0
	}
	if p.position > len(p.tokens) {
		p.position = len(p.tokens)
	}
}

func (p *parser) peak() Token {
	p.fixPointer()
	if p.endOfStream() {
		fmt.Println("peaked past end of stream, please file a bug report")
		os.Exit(1)
	}
	return p.tokens[p.position]
}

func (p *parser) previous() Token {
	return p.tokens[p.position-1]
}

func (p *parser) next() Token {
	t := p.peak()
	p.advancePointer()
	return t
}
