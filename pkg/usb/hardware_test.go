/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package usb_test

import (
	"testing"
	"time"

	"github.com/Molorius/ulp-c/pkg/asm"
	"github.com/Molorius/ulp-c/pkg/usb"
)

const reservedBytes = 8176
const reduce = true

func TestSimpleHardware(t *testing.T) {
	h := usb.Hardware{}
	assembly := `
	.boot
	halt

	.boot.data
	.int 0, 0, 0 // the mutex
	.int 1, 0 // DONE, 0
	`

	// compile binary
	assembler := asm.Assembler{}
	bin, err := assembler.BuildFile(assembly, "testSimpleHardware.S", reservedBytes, reduce)
	if err != nil {
		t.Fatalf("Failed to compile: %s", err)
	}

	// open port if the environment variable is set
	timeout := 2 * time.Second
	err = h.OpenPortFromEnv(timeout)
	if err != nil {
		t.Fatal(err)
	}
	if !h.PortSet() {
		t.Skipf("Port not set, skipping")
	}
	defer h.Close()

	// test that it works repeatedly
	testRuns := 5
	for i := 0; i < testRuns; i++ {
		_, err = h.Execute(bin, t)
		if err != nil {
			t.Fatalf("Test %d failed: %s", i, err)
		}
	}
}

func TestMutex(t *testing.T) {
	h := usb.Hardware{}
	assembly := `
	.boot
	move r2, mutex
	move r1, 1
	st r1, r2, 0 // flag[0] = true
	st r1, r2, 2 // turn = 1
	// while (flag[1] && turn == 1) { }
loop:
	ld r0, r2, 1
	jumpr loop, 1, lt
	ld r0, r2, 2
	jumpr loop, 0, gt
end:
	st r1, r2, 3 // set to DONE
	move r0, 0
	st r0, r2, 0 // flag[0] = false
	halt

	.boot.data
mutex:
	.int 0, 0, 0 // the mutex
	.int 0, 0 // ESP_ACK, 0
	`

	// compile binary
	assembler := asm.Assembler{}
	bin, err := assembler.BuildFile(assembly, "testSimpleHardware.S", reservedBytes, reduce)
	if err != nil {
		t.Fatalf("Failed to compile: %s", err)
	}

	timeout := 2 * time.Second
	err = h.OpenPortFromEnv(timeout)
	if err != nil {
		t.Fatal(err)
	}
	if !h.PortSet() {
		t.Skipf("Port not set, skipping")
	}
	defer h.Close()

	// test that it works repeatedly
	testRuns := 5
	for i := 0; i < testRuns; i++ {
		_, err = h.Execute(bin, t)
		if err != nil {
			t.Fatalf("Test %d failed: %s", i, err)
		}
	}
}
