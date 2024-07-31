package main

import (
	"fmt"

	"github.com/Molorius/ulp-c/pkg/asm"
)

func main() {
	asm := asm.Assembler{}
	err := asm.BuildFile("0-(4+3) * 7\n", "quick_test.S")
	if err != nil {
		fmt.Println(err)
	}
}
