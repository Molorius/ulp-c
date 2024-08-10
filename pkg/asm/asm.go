package asm

import (
	"errors"
	"fmt"
)

type Assembler struct {
}

func (asm *Assembler) BuildFile(content string, name string, reservedBytes int) ([]byte, error) {
	s := scanner{}
	tokens, err := s.scanFile(content, name)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error while scanning"), err)
	}
	p := parser{}
	stmnts, err := p.parseTokens(tokens)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("error while parsing"), err)
	}
	c := Compiler{}
	bin, err := c.Compile(stmnts, reservedBytes)
	if err != nil {
		return nil, err
	}
	return bin, nil
}
