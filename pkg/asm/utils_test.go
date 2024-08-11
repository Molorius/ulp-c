/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package asm

import (
	"testing"
)

func TestTheRunnerWithHeaders(t *testing.T) {
	tests := []struct {
		name   string
		asm    string
		expect string
	}{
		{
			name:   "simple",
			asm:    "",
			expect: "",
		},
		{
			name: "print_u16",
			asm: `
			move r0, 123
			st r0, r3, 0
			move r2, .+2
			jump print_u16

			move r0, 456
			st r0, r3, 0
			move r2, .+2
			jump print_u16
			`,
			expect: "123 456 ",
		},
		{
			name: "print_char",
			asm: `
			move r0, 65
			st r0, r3, 0
			move r2, .+2
			jump print_char

			move r0, 66
			st r0, r3, 0
			move r2, .+2
			jump print_char
			`,
			expect: "AB",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RunTestWithHeader(t, tt.name, tt.asm, tt.expect)
		})
	}
}
