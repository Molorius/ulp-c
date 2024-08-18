/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package asm

import (
	"testing"

	"github.com/Molorius/ulp-c/pkg/emu"
	"github.com/Molorius/ulp-c/pkg/usb"
)

const TEST_PRELUDE = `
	.boot
move r3, __stack_end - 32 // initialize stack with a bit of space
jump main

	.boot.data
.int 0, 0, 0 // mutex
.int 0, 0 // send to esp

// void ulp_mutex_take(void) {
//     flag[0] = true;
//     turn = 1;
//     while ((flag[1] > 0) && (turn > 0)) {
//     }
// }
	.text
ulp_mutex_take:
	st r0, r3, -1 // store value on stack
	move r0, 1
	st r0, r0, (__boot_data_start-1) // flag[0] = true
	st r0, r0, (__boot_data_start-1)+1 // turn = 1
	// while (flag[1] && turn == 1) { }
ulp_mutex_take.loop:
	// this loop assumes that 'flag[1]' and 'turn' can only be 0 or 1
	ld r0, r0, (__boot_data_start-1)+1 // r0 = flag[1]
	jumpr ulp_mutex_take.end, 1, lt
	ld r0, r0, (__boot_data_start-1)+2 // r0 = turn
	jumpr ulp_mutex_take.loop, 0, gt
ulp_mutex_take.end:
	ld r0, r3, -1 // restore value
	jump r2

// void ulp_mutex_give(void) {
//     flag[0] = false;
// }
	.text
ulp_mutex_give:
	st r0, r3, -1
	move r0, 0
	st r0, r0, __boot_data_start // flag[0] = false
	ld r0, r3, -1
	jump r2

// void send_esp(uint16_t fn, uint16_t param)
// {
//     uint16_t ack;
//     ulp_mutex_take();
//     for (;;) {
//         ack = esp_write[0];
//         if (ack == 0) {
//             break;
//         }
//         ulp_mutex_give();
//         ulp_mutex_take();
//     }
//     esp_write[0] = fn;
//     esp_write[1] = param;
//     ulp_mutex_give();
// }
	.text
send_esp:
	sub r3, r3, 2
	st r0, r3, 0
	st r2, r3, 1

	move r2, send_esp.loop // ulp_mutex_take()
	jump ulp_mutex_take
send_esp.loop:
	ld r0, r2, __boot_data_start - send_esp.loop + 3
	jumpr send_esp.end, 0, le
	move r2, send_esp.loop.1 // ulp_mutex_give()
	jump ulp_mutex_give
send_esp.loop.1:
	move r2, send_esp.loop // ulp_mutex_take()
	jump ulp_mutex_take
send_esp.end:
	ld r0, r3, 2 // esp_write[0] = fn
	st r0, r2, __boot_data_start - send_esp.loop + 3
	ld r0, r3, 3 // esp_write[1] = param
	st r0, r2, __boot_data_start - send_esp.loop + 4
	move r2, .+2 // ulp_mutex_give()
	jump ulp_mutex_give

	ld r2, r3, 1
	ld r0, r3, 0
	add r3, r3, 2
	jump r2

// __noreturn void done(void)
// {
//     send_esp(ESP_DONE, ); // not valid c
//     halt();
// }
    .text
done:
	sub r3, r3, 1 // increase stack
	move r0, 1 // set to DONE
	st r0, r3, 0
	move r2, .+2
	jump send_esp
    halt

// void print_u16(uint16_t c)
// {
//     send_esp(ESP_PRINT_U16, c);
// }
    .text
print_u16:
    sub r3, r3, 3 // increase stack for temp and function call
    st r2, r3, 2 // store r2 in temp slot
    move r2, 2
print_u16.call:
    st r2, r3, 0 // store constant for send_esp
	ld r2, r3, 3 // load c
    st r2, r3, 1 // store c for send_esp
	move r2, .+2 // send_esp()
    jump send_esp
    ld r2, r3, 2 // reload r2
    add r3, r3, 3 // restore stack
    jump r2

// void print_char(uint16_t c)
// {
//     send_esp(ESP_PRINT_CHAR, c);
// }
    .text
print_char:
    sub r3, r3, 3 // increase stack for temp and function call
	st r2, r3, 2
	move r2, 3
	jump print_u16.call

	.text
main:
`

const TEST_POSTLUDE = `
jump done
`

func CompileTestWithHeader(asm string) ([]byte, error) {
	a := Assembler{}
	filename := "test.S"
	reserved := 8176
	return a.BuildFile(asm, filename, reserved)
}

func RunTestWithHeader(t *testing.T, asm string, expect string) {
	content := TEST_PRELUDE + asm + TEST_POSTLUDE
	RunTest(t, content, expect)
}

func RunTest(t *testing.T, asm string, expect string) {
	// compile the binary
	a := Assembler{}
	filename := "test.S"
	reserved := 8176
	bin, err := a.BuildFile(asm, filename, reserved)
	if err != nil {
		t.Fatalf("Failed to compile: %s", err)
	}

	// run the test on hardware
	t.Run("hardware", func(t *testing.T) {
		hw := usb.Hardware{}
		port, err := hw.EnvPort()
		if err != nil {
			t.Skipf("Skipping test: %v", err)
		}
		got, err := hw.Execute(port, bin)
		if err != nil {
			t.Fatalf("Execution failed: %s", err)
		}
		if got != expect {
			t.Errorf("expected \"%s\" got \"%s\"", expect, got)
		}
	})

	// run the test on emulator
	t.Run("emulator", func(t *testing.T) {
		u := emu.UlpEmu{}
		maxSeconds := 1 // maximum emulated seconds
		maxCycles := uint64(8_000_000 * maxSeconds)
		err := u.LoadBinary(bin)
		if err != nil {
			t.Fatalf("Loading binary failed: %s", err)
		}
		got, err := u.RunWithSystem(maxCycles)
		if err != nil {
			t.Fatalf("Execution failed: %s", err)
		}
		if got != expect {
			t.Errorf("expected \"%s\" got \"%s\"", expect, got)
		}
	})
}
