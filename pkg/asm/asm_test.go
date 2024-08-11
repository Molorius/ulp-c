/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package asm

import "testing"

func TestSimple(t *testing.T) {
	RunTestWithHeader(t, "simple", "", "")
}

func TestPrintU16(t *testing.T) {
	RunTestWithHeader(t, "printU16", `
	move r0, 123
	st r0, r3, 0
	move r2, .+2
	jump print_u16

	move r0, 456
	st r0, r3, 0
	move r2, .+2
	jump print_u16
	`, "123 456 ")
}

func TestPrintChar(t *testing.T) {
	RunTestWithHeader(t, "printU16", `
	move r0, 65
	st r0, r3, 0
	move r2, .+2
	jump print_char

	move r0, 66
	st r0, r3, 0
	move r2, .+2
	jump print_char
	`, "AB")
}
