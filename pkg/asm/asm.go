package asm

import (
	"errors"
	"fmt"
)

type Assembler struct {
}

func (asm *Assembler) BuildFile(content string, name string) error {
	s := scanner{}
	tokens, err := s.scanFile(content, name)
	if err != nil {
		return errors.Join(fmt.Errorf("error while scanning"), err)
	}
	fmt.Printf("tokens: %q\n", tokens)
	p := parser{}
	stmnts, err := p.parseTokens(tokens)
	if err != nil {
		return errors.Join(fmt.Errorf("error while parsing"), err)
	}
	fmt.Printf("statements: %s\n", stmnts)
	c := Compiler{}
	bin, err := c.Compile(stmnts)
	fmt.Printf("binary: %v\n", bin)
	return err
}
