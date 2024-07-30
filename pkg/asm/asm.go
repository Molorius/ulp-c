package asm

import "fmt"

type Assembler struct {
}

func (asm *Assembler) BuildFile(content string, name string) error {
	s := scanner{}
	tokens, err := s.scanFile(content, name)
	fmt.Printf("tokens: %s\n", tokens)
	return err
}
