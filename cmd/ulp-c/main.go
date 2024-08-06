package main

import (
	"fmt"
	"os"

	"github.com/Molorius/ulp-c/pkg/asm"
)

func main() {
	asm := asm.Assembler{}
	s := `
	.boot
	.global boot
boot:
	move r0, 0
	move r1, 0
	move r3, 8172/4
	move r2, .+2
	jump main
finished:
	jump finished

	.text
	.global main
main:
	st r0, r3, 0
	st r1, r3, 1
	st r2, r3, 2

	ld r2, r3, 2
	ld r1, r3, 1
	ld r0, r3, 0
	jump r2
	.int 0, 1, 2, main+1
	.int 0x1234
	jumpr . - 126, 0xFFFF, eq
	jumps main, 123, eq
	jumps main, 0xFE, gt
	stage_rst
	stage_inc 5
	stage_dec 4
	halt
	wake
	sleep 0
	wait 10
`
	bin, err := asm.BuildFile(s, "quick_test.S")
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create("out.bin")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = f.Write(bin)
	if err != nil {
		fmt.Println(err)
		return
	}
}
