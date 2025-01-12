package asm

import (
	"fmt"
	"testing"
)

const reduceInstructions int = 500

func testReduceHelper(instructions int) string {
	asm := ""
	for i := 0; i < instructions; i++ {
		line := fmt.Sprintf("move r0, %d\r\njump r0\r\n", i)
		asm = asm + line + line
	}
	return asm
}

func TestReduce(t *testing.T) {
	size := reduceInstructions * 4

	asm := testReduceHelper(reduceInstructions)
	a := Assembler{}
	bin, err := a.BuildFile(asm, "test.S", 8176, false)
	if err != nil {
		t.Fatalf("Compiling failed: %s", err)
	}
	binReduced, err := a.BuildFile(asm, "test.S", 8176, true)
	if err != nil {
		t.Fatalf("Compiling reduced failed: %s", err)
	}
	diff := len(bin) - len(binReduced)
	if diff != size {
		t.Fatalf("Wrong size reduction expected %d got %d", size, diff)
	}
}

func BenchmarkCompile(b *testing.B) {
	asm := testReduceHelper(reduceInstructions)
	a := Assembler{}
	for i := 0; i < b.N; i++ {
		_, err := a.BuildFile(asm, "test.S", 8176, false)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkCompileWithReduce(b *testing.B) {
	asm := testReduceHelper(reduceInstructions)
	a := Assembler{}
	for i := 0; i < b.N; i++ {
		_, err := a.BuildFile(asm, "test.S", 8176, true)
		if err != nil {
			b.Error(err)
		}
	}
}
