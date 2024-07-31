package main

import (
	"fmt"

	"github.com/Molorius/ulp-c/pkg/asm"
)

func main() {
	asm := asm.Assembler{}
	err := asm.BuildFile("jump 0, ov \n.global test  \ntest: add r0, r0, 0+5*(.+(4*16))\ntest2: sub r0, r0, 1+2\nwait 123", "quick_test.S")
	if err != nil {
		fmt.Println(err)
	}
}
