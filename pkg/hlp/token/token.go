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
	At                       // token for @
	Pound                    // token for #
	Comma                    // token for ,

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

	Asm      // token for __asm__ function modifier
	Noreturn // token for noreturn function modifier
	Func     // token for function definition
	Static   // token for static global variable modifier
	Extern   // token for extern modifier

	// hardware instructions

	RegWr // token for reg_wr instruction
	RegRd // token for reg_rd instruction
	Wait  // token for wait instruction
	I2cWr // token for i2c_wr instruction
	I2cRd // token for i2c_rd instruction
	Halt  // token for halt instruction
	Wake  // token for wake instruction
	Sleep // token for sleep instruction
	Adc   // token for adc instruction

	// literals
	Identifier
	Number
	String // token for a string

	EndOfFile
	Unknown
)

var toToken = map[string]Type{
	";":        Semicolon,
	":":        Colon,
	"&":        Ampersand,
	"=":        Equal,
	"(":        LeftParen,
	")":        RightParen,
	"[":        LeftBracket,
	"]":        RightBracket,
	"{":        LeftBrace,
	"}":        RightBrace,
	"@":        At,
	"#":        Pound,
	",":        Comma,
	"goto":     Goto,
	"if":       If,
	"ifEq":     IfEq,
	"ifOv":     IfOv,
	"+":        Plus,
	"-":        Minus,
	"|":        Or,
	"<<":       Lsh,
	">>":       Rsh,
	"==":       EqualEqual,
	"!=":       NotEqual,
	">":        Greater,
	"<":        Less,
	">=":       GreaterEqual,
	"<=":       LessEqual,
	"__asm__":  Asm,
	"noreturn": Noreturn,
	"func":     Func,
	"static":   Static,
	"extern":   Extern,
	"reg_wr":   RegWr,
	"reg_rd":   RegRd,
	"wait":     Wait,
	"i2c_wr":   I2cWr,
	"i2c_rd":   I2cRd,
	"halt":     Halt,
	"wake":     Wake,
	"sleep":    Sleep,
	"adc":      Adc,
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
	case String:
		return "String"
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
