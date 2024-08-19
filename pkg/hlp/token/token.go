/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package token

type Type int

const (
	// punctuation

	Semicolon    Type = iota // token for ;
	Colon                    // token for :
	Ampersand                // token for &
	Equal                    // token for =
	LeftParen                // token for (
	RightParen               // token for )
	LeftBrace                // token for {
	RightBrace               // token for }
	LeftBracket              // token for [
	RightBracket             // token for ]

	// gotos

	Goto // token for "goto"
	If   // token for "if"
	IfEq // token for low level "ifEq"
	IfOv // token for low level "ifOv"

	// operations

	Plus  // token for + operator
	Minus // token for - operator
	Or    // token for | operator
	Lsh   // token for << operator
	Rsh   // token for >> operator

	// comparison

	EqualEqual   // token for ==
	NotEqual     // token for !=
	Greater      // token for >
	Less         // token for <
	GreaterEqual // token for >=
	LessEqual    // token for <=

	// other

	Asm // token for __asm__ modifier

	// literals
	Identifier
	Number

	EndOfFile
	Unknown
)

var toToken = map[string]Type{
	";":       Semicolon,
	":":       Colon,
	"&":       Ampersand,
	"=":       Equal,
	"(":       LeftParen,
	")":       RightParen,
	"[":       LeftBracket,
	"]":       RightBracket,
	"{":       LeftBrace,
	"}":       RightBrace,
	"goto":    Goto,
	"if":      If,
	"ifEq":    IfEq,
	"ifOv":    IfOv,
	"+":       Plus,
	"-":       Minus,
	"|":       Or,
	"<<":      Lsh,
	">>":      Rsh,
	"==":      EqualEqual,
	"!=":      NotEqual,
	">":       Greater,
	"<":       Less,
	">=":      GreaterEqual,
	"<=":      LessEqual,
	"__asm__": Asm,
}
var toString map[Type]string

func init() {
	toString = make(map[Type]string)
	for s, t := range toToken {
		toString[t] = s
	}
}

func (t Type) String() string {
	val, ok := toString[t]
	if ok {
		return val
	}
	switch t {
	case Identifier:
		return "Identifier"
	case Number:
		return "Number"
	case EndOfFile:
		return "EOF"
	default:
		return "UNKNOWN"
	}
}

func ToType(str string) Type {
	val, ok := toToken[str]
	if ok {
		return val
	}
	return Unknown
}
