package asm

import (
	"fmt"

	"github.com/Molorius/ulp-c/pkg/asm/token"
)

type ExpectedTokenError struct {
	expected token.Type
	got      Token
}

func (e ExpectedTokenError) Error() string {
	if e.got.TokenType == token.EndOfFile {
		return fmt.Sprintf("%s: expected \"%s\" but hit end of file", e.got.Ref, e.expected)
	}
	return fmt.Sprintf("%s: expected \"%s\" got \"%s\"", e.got.Ref, e.expected, e.got.Lexeme)
}

type GenericTokenError struct {
	token   Token
	message string
}

func (e GenericTokenError) Error() string {
	return fmt.Sprintf("%s: %s", e.token.Ref, e.message)
}

type UnknownTokenError struct {
	token Token
}

func (e UnknownTokenError) Error() string {
	return fmt.Sprintf("%s: unknown token \"%s\"", e.token.Ref, e.token.Lexeme)
}

type UnfinishedError struct {
	token    Token
	expected string
}

func (e UnfinishedError) Error() string {
	return fmt.Sprintf("%s: \"%s\" is unfinished, expected %s", e.token.Ref, e.token.Lexeme, e.expected)
}
