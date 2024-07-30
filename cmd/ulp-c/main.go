package main

import "github.com/Molorius/ulp-c/pkg/asm"

func main() {
	asm := asm.Assembler{}
	asm.BuildFile("add r0, r0, 1 \n sub r0, r2, 47", "quick_test.S")
}
