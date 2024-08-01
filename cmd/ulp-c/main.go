package main

import (
	"fmt"

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
	move r2, .+8
	jump main
finished:
	jump finished

	.text
	.global main
main:
	sub r3, r3, 3
	st r0, r3, 4*0
	st r1, r3, 4*1
	st r2, r3, 4*2

	ld r2, r3, 4*2
	ld r1, r3, 4*1
	ld r0, r3, 4*0
	add r3, r3, 3
	jump r2

	.data
	move r0, 0
	move r0, 0

	.bss
	move r0, 0
`
	err := asm.BuildFile(s, "quick_test.S")
	if err != nil {
		fmt.Println(err)
	}
}
