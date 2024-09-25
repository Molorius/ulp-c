/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package hlp

import (
	"testing"

	"github.com/Molorius/ulp-c/pkg/hlp/token"
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

func eof() Token {
	return tok(token.EndOfFile)
}

func str(s string) Token {
	return Token{TokenType: token.String, StringVal: s}
}

func Test_scanner_scanFile(t *testing.T) {
	f := "test.hlp"
	tests := []struct {
		name    string
		hlp     string
		want    []Token
		wantErr bool
	}{
		{
			name: "basic",
			hlp:  "{ ( + )}",
			want: []Token{
				tok(token.LeftBrace),
				tok(token.LeftParen),
				tok(token.Plus),
				tok(token.RightParen),
				tok(token.RightBrace),
				eof(),
			},
		},
		{
			name: "<<",
			hlp:  "<<",
			want: []Token{
				tok(token.Lsh),
				eof(),
			},
		},
		{
			name: "<=",
			hlp:  "<=",
			want: []Token{
				tok(token.LessEqual),
				eof(),
			},
		},
		{
			name: ">>",
			hlp:  ">>",
			want: []Token{
				tok(token.Rsh),
				eof(),
			},
		},
		{
			name: ">=",
			hlp:  ">=",
			want: []Token{
				tok(token.GreaterEqual),
				eof(),
			},
		},
		{
			name: "!=",
			hlp:  "!=",
			want: []Token{
				tok(token.NotEqual),
				eof(),
			},
		},
		{
			name: "><",
			hlp:  "><",
			want: []Token{
				tok(token.Greater),
				tok(token.Less),
				eof(),
			},
		},
		{
			name: "string",
			hlp:  "\"test0\"\"test1\" \"test2\"",
			want: []Token{
				str("test0"),
				str("test1"),
				str("test2"),
				eof(),
			},
		},
		{
			name:    "incomplete string",
			hlp:     "\"test",
			wantErr: true,
		},
		{
			name: "ignore comment",
			hlp:  "-// ignore me\n+",
			want: []Token{
				tok(token.Minus),
				tok(token.Plus),
				eof(),
			},
		},
		{
			name: "ignore whitespace",
			hlp:  " \n\t\r",
			want: []Token{eof()},
		},
		{
			name: "ignore multiline comment",
			hlp:  "-/*\n\nstill ignore me\n\n*/+",
			want: []Token{
				tok(token.Minus),
				tok(token.Plus),
				eof(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Scanner{}
			got, err := s.ScanFile(tt.hlp, f)
			if (err != nil) != tt.wantErr {
				t.Errorf("scanner.scanFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !tokens_equal(got, tt.want) {
				t.Errorf("scanner.scanFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
