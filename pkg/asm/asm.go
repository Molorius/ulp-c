package asm

import (
	"errors"
	"fmt"
)

type Assembler struct {
}

func (asm *Assembler) BuildFile(content string, name string) ([]byte, error) {
	s := scanner{}
	tokens, err := s.scanFile(content, name)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error while scanning"), err)
	}
	fmt.Printf("tokens: %q\n", tokens)
	p := parser{}
	stmnts, err := p.parseTokens(tokens)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error while parsing"), err)
	}
	fmt.Printf("statements: %s\n", stmnts)
	c := Compiler{}
	bin, err := c.Compile(stmnts)
	if err != nil {
		return nil, err
	}
	fmt.Printf("binary: %v\n", bin)
	return bin, nil
}
