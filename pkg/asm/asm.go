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
		return err
	}
	fmt.Printf("tokens: %q\n", tokens)
	p := parser{}
	expr, err := p.parseTokens(tokens)
	if err != nil {
		return errors.Join(fmt.Errorf("parse error"), err)
	}
	fmt.Printf("expression: %s\n", expr)
	return err
}
