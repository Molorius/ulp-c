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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Scanner{}
			got, err := s.ScanFile(tt.hlp, f)
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
