/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package hlp

type HlpNumber uint16

type StaticStatement interface {
}

type StaticStatementFunction struct {
	Ident      Token
	NoReturn   bool
	Static     bool
	Extern     bool
	Parameters []Definition
	Returns    int
}

type StaticStatementAsm struct {
	Ident      Token
	Parameters []Definition
	Returns    int
	Asm        []string
}

type StaticStatementArray struct {
	Ident Token
	N     HlpNumber
}

type GlobalVar struct {
	Ident  Token     // the identity of this global variable
	Array  bool      // true if this is an array (uses the "@" operator)
	Extern bool      // true if this is tagged with the "extern" modifier
	Value  []Primary // the initial values
}

type Var struct {
	Ident     Token // the identity of the variable access
	Array     bool  // is this an array? Uses the "#" operator
	Offset    int   // the offset after "#" in an array
	AddressOf bool  // uses the "&" operator
}

type Primary interface {
}

type PrimaryNumber struct {
	N HlpNumber
}

type PrimaryVar struct {
	V Var
}

type Definition interface {
}

type DefinitionInt struct {
	Ident   Token
	Initial HlpNumber
}

type DefinitionArray struct {
	Ident   Token
	Size    int
	Initial []HlpNumber
}