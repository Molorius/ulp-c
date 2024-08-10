/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
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
