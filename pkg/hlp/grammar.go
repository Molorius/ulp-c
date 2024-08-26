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
	NoReturn bool
	Inputs   []PrimaryVar
	Outputs  int
}

type StaticStatementAsm struct {
	Inputs  []PrimaryVar
	Outputs int
	Asm     []string
}

type StaticStatementArray struct {
	Ident Token
	N     HlpNumber
}

type Var struct {
	Ident   Token
	Offset  int
	IsArray bool
}

type Primary interface {
}

type PrimaryNumber struct {
	N HlpNumber
}

type PrimaryVar struct {
	V Var
}
