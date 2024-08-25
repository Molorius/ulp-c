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
	Compiler Compiler
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
	asm.Compiler = Compiler{}
	bin, err := asm.Compiler.CompileToBin(stmnts, reservedBytes)
	if err != nil {
		return nil, err
	}
	return bin, nil
}

func (asm *Assembler) BuildAssembly(content string, name string, reservedBytes int) ([]byte, error) {
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
	asm.Compiler = Compiler{}
	bin, err := asm.Compiler.CompileToAsm(stmnts, reservedBytes)
	if err != nil {
		return nil, err
	}
	return bin, nil
}
