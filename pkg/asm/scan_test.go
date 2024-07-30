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
			asm:  "add // this is a test",
			want: []Token{tok(token.Add), tok(token.EndOfFile)},
		},
		{
			name: "pound comment",
			asm:  "add # this is a test",
			want: []Token{tok(token.Add), tok(token.EndOfFile)},
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
