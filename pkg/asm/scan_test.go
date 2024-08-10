/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package asm

import (
	"testing"

	"github.com/Molorius/ulp-c/pkg/asm/token"
)

func tokens_equal(a []Token, b []Token) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(&b[i]) {
			return false
		}
	}
	return true
}

func tok(t token.Type) Token {
	return Token{TokenType: t}
}

func num(i int) Token {
	return Token{TokenType: token.Number, Number: i}
}

func ident(s string) Token {
	return Token{TokenType: token.Identifier, Lexeme: s}
}

func unknown(s string) Token {
	return Token{TokenType: token.Unknown, Lexeme: s}
}

func newline() Token {
	return Token{TokenType: token.NewLine}
}

func Test_scanner_scanFile(t *testing.T) {
	f := "test.S"
	tests := []struct {
		name    string
		asm     string
		want    []Token
		wantErr bool
	}{
		{
			name: "basic",
			asm:  "move r0, r1",
			want: []Token{
				tok(token.Move),
				tok(token.R0),
				tok(token.Comma),
				tok(token.R1),
				tok(token.EndOfFile),
			},
		},
		{
			name: "slash comment",
			asm:  "add // this is a test\n1",
			want: []Token{tok(token.Add), newline(), num(1), tok(token.EndOfFile)},
		},
		{
			name: "pound comment",
			asm:  "add # this is a test\n1",
			want: []Token{tok(token.Add), newline(), num(1), tok(token.EndOfFile)},
		},
		{
			name: "multiline comment",
			asm:  "add /* this\nis\na\ntest */1",
			want: []Token{tok(token.Add), num(1), tok(token.EndOfFile)},
		},
		{
			name: "inline comment",
			asm:  "add/* this is a test */1",
			want: []Token{tok(token.Add), num(1), tok(token.EndOfFile)},
		},
		{
			// I'm not sure if I want an error here instead, but
			// currently this will silently exit
			name: "unfinished comment",
			asm:  "123/*",
			want: []Token{num(123), tok(token.EndOfFile)},
		},
		{
			name: "adjacent characters",
			asm:  ".+42*7+TEST",
			want: []Token{
				tok(token.Here),
				tok(token.Plus),
				num(42),
				tok(token.Star),
				num(7),
				tok(token.Plus),
				ident("TEST"),
				tok(token.EndOfFile),
			},
		},
		{
			name: "sections",
			asm:  ".boot .text .data .bss",
			want: []Token{
				tok(token.Boot),
				tok(token.Text),
				tok(token.Data),
				tok(token.Bss),
				tok(token.EndOfFile),
			},
		},
		{
			name: "numbers",
			asm:  "0 123 0x123 0o15 -50",
			want: []Token{
				num(0),
				num(123),
				num(0x123),
				num(0o15),
				tok(token.Minus),
				num(50),
				tok(token.EndOfFile),
			},
		},
		{
			name: "error assorted unknown chars",
			asm:  "!~@$%",
			want: []Token{
				unknown("!"),
				unknown("~"),
				unknown("@"),
				unknown("$"),
				unknown("%"),
				tok(token.EndOfFile),
			},
			wantErr: true,
		},
		{
			name: "error identifier starts with number",
			asm:  "1test",
			want: []Token{
				unknown("1test"),
				tok(token.EndOfFile),
			},
			wantErr: true,
		},
		{
			name: "error unknown macro",
			asm:  ".int .boot.data",
			want: []Token{
				unknown(".int"),
				unknown(".boot.data"),
				tok(token.EndOfFile),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scanner{}
			got, err := s.scanFile(tt.asm, f)
			if (err != nil) != tt.wantErr {
				t.Errorf("scanner.scanFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tokens_equal(got, tt.want) {
				t.Errorf("scanner.scanFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
