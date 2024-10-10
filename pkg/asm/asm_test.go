/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package asm

import (
	"fmt"
	"testing"
)

func TestJumpr(t *testing.T) {
	tests := []struct {
		name   string
		asm    string
		expect string
	}{
		{
			name: "less than or equal",
			asm: `
			move r0, 0
			jumpr t.0, 0, le
			jump done
		t.0:
			jumpr t.1, 0xFFFF, le
			jump done
		t.1:

			move r0, 0xFFFE
			jumpr t.2, 0xFFFE, le
			jump done
		t.2:
			jumpr done, 0xFFFD, le

			move r0, 0xFFFF
			jumpr t.3, 0xFFFF, le
			jump done
		t.3:
			jumpr done, 0xFFFE, le

			move r0, 1
			st r0, r3, 0
			call print_u16
			`,
			expect: "1 ",
		},
	}
	r := Runner{}
	r.SetDefaults()
	err := r.SetupPort()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.RunTestWithHeader(t, tt.asm, tt.expect)
		})
	}
}

func buildOrderOfOps(ops string) string {
	s := `
	move r0, %s
	st r0, r3, 0
	call print_u16
	`
	return fmt.Sprintf(s, ops)
}

func TestOrderOfOperations(t *testing.T) {
	tests := []struct {
		name   string
		ops    string
		expect string
	}{
		{
			name:   "0",
			ops:    "0",
			expect: "0 ",
		},
		{
			name:   "mult add",
			ops:    "1 + 2*3",
			expect: "7 ",
		},
		{
			name:   "add mult",
			ops:    "2*3 + 1",
			expect: "7 ",
		},
		{
			name:   "add shift",
			ops:    "1<<3 + 5",
			expect: "256 ",
		},
		{
			name:   "add shift paren",
			ops:    "(1<<3) + 5",
			expect: "13 ",
		},
	}
	r := Runner{}
	r.SetDefaults()
	err := r.SetupPort()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.RunTestWithHeader(t, buildOrderOfOps(tt.ops), tt.expect)
		})
	}
}

func buildAlu(ops string) string {
	s := `
	%s
	st r0, r3, 0
	call print_u16
	`
	return fmt.Sprintf(s, ops)
}

func TestAlu(t *testing.T) {
	tests := []struct {
		name   string
		ops    string
		expect string
	}{
		{
			name:   "| immediate",
			ops:    "move r0, 0b1100\r\n or r0, r0, 0b01",
			expect: "13 ",
		},
	}
	r := Runner{}
	r.SetDefaults()
	err := r.SetupPort()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r.RunTestWithHeader(t, buildAlu(tt.ops), tt.expect)
		})
	}
}
