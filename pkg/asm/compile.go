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

	"github.com/Molorius/ulp-c/pkg/asm/token"
)

type Label struct {
	Name    string
	Value   int
	Global  bool
	section *Section
}

type Section struct {
	Size   int // size in bytes
	Bin    []byte
	Offset int // offset from start of program
}

func (s *Section) Validate(name string) error {
	if s.Size != len(s.Bin) {
		return fmt.Errorf("Section %s was size %d but expected %d, please file a bug report", name, len(s.Bin), s.Size)
	}
	return nil
}

type Compiler struct {
	program        []Stmnt
	position       int // position within the program
	Labels         map[string]*Label
	preLabels      map[string]int // these contain offsets relative to the section
	Boot           Section
	BootData       Section
	Text           Section
	Data           Section
	Bss            Section
	Stack          Section // data not placed here
	CurrentSection *Section
}

func (c *Compiler) compile(program []Stmnt, reservedBytes int) error {
	c.program = program
	c.Labels = make(map[string]*Label)
	c.preLabels = make(map[string]int)
	c.Boot = Section{}
	c.BootData = Section{}
	c.Text = Section{}
	c.Data = Section{}
	c.Bss = Section{}
	c.Stack = Section{}
	err := c.genPreLabels()
	if err != nil {
		return err
	}
	err = c.genLabels(reservedBytes)
	if err != nil {
		return err
	}
	err = c.genGlobals()
	if err != nil {
		return err
	}
	err = c.compileAll()
	if err != nil {
		return err
	}
	err = c.validateSections()
	return err
}

func (c *Compiler) CompileToBin(program []Stmnt, reservedBytes int) ([]byte, error) {
	err := c.compile(program, reservedBytes)
	if err != nil {
		return nil, err
	}
	bin, err := c.buildBinary()
	return bin, err
}

func (c *Compiler) CompileToAsm(program []Stmnt, reservedBytes int) ([]byte, error) {
	// find all of the addresses we need
	err := c.compile(program, reservedBytes)
	if err != nil {
		return nil, err
	}
	addresses := make(map[int]*Label)
	for _, label := range c.Labels {
		addresses[label.Value] = label
	}
	s := ".text\n"
	start := 0
	b := make([]byte, 0)
	b = append(b, c.Boot.Bin...)
	b = append(b, c.Text.Bin...)
	s = c.buildAsm(start, s, b, addresses)

	start += len(b)
	s += ".data\n"
	b = make([]byte, 0)
	b = append(b, c.BootData.Bin...)
	b = append(b, c.Data.Bin...)
	s = c.buildAsm(start, s, b, addresses)

	start += len(b)
	s += ".bss\n"
	b = make([]byte, 0)
	b = append(b, c.Bss.Bin...)
	s = c.buildAsm(start, s, b, addresses)
	s += fmt.Sprintf(".skip %d", c.Stack.Size)

	return []byte(s), nil
}

func (c *Compiler) buildAsm(start int, s string, bin []byte, addr map[int]*Label) string {
	for pos, b := range bin {
		i := start + pos
		label, ok := addr[i]
		if ok {
			if label.Global {
				s += fmt.Sprintf(".global %s\n%s:\n", label.Name, label.Name)
			}
		}
		s += fmt.Sprintf("    .byte 0x%02X\n", b)
	}
	return s
}

func (c *Compiler) genPreLabels() error {
	c.position = 0
	c.CurrentSection = &c.Text
	for _, stmnt := range c.program {
		c.CurrentSection.Size += stmnt.Size()
		switch s := stmnt.(type) {
		case StmntDirective:
			c.setSection(s.Directive.TokenType)
		case StmntLabel:
			offset := c.CurrentSection.Size
			name := s.Label.Lexeme
			c.preLabels[name] = offset
			l := Label{
				Name:    name,
				section: c.CurrentSection,
			}
			c.Labels[name] = &l
		}
	}
	return nil
}

func (c *Compiler) FormatSections() string {
	stackStr := ""
	if c.Stack.Size != 0 {
		stackStr = fmt.Sprintf(" .stack=%d", c.Stack.Size)
	}
	total := c.Boot.Size + c.BootData.Size + c.Text.Size + c.Data.Size + c.Bss.Size + c.Stack.Size
	return fmt.Sprintf(".boot=%d .boot.data=%d .text=%d .data=%d .bss=%d%s total=%d",
		c.Boot.Size, c.BootData.Size, c.Text.Size, c.Data.Size, c.Bss.Size, stackStr, total)
}

func (c *Compiler) genLabels(reservedBytes int) error {
	// resolve section offsets
	c.Boot.Offset = 0
	c.Text.Offset = c.Boot.Offset + c.Boot.Size
	c.BootData.Offset = c.Text.Offset + c.Text.Size
	c.Data.Offset = c.BootData.Offset + c.BootData.Size
	c.Bss.Offset = c.Data.Offset + c.Data.Size
	// data is never placed in stack, calculate remaining memory
	c.Stack.Offset = c.Bss.Offset + c.Bss.Size
	stackSize := reservedBytes - c.Stack.Offset
	if stackSize < 0 {
		return fmt.Errorf("overflowing the %d reserved bytes: %s", reservedBytes, c.FormatSections())
	}
	c.Stack.Size = stackSize

	// resolve label offsets
	for name, offset := range c.preLabels {
		c.Labels[name].Value = c.Labels[name].section.Offset + offset
	}

	// generate section labels
	c.Labels["__boot_start"] = &Label{
		Name:  "__boot_start",
		Value: c.Boot.Offset,
	}
	c.Labels["__boot_end"] = &Label{
		Name:  "__boot_end",
		Value: c.Boot.Offset + c.Boot.Size,
	}
	c.Labels["__text_start"] = &Label{
		Name:  "__text_start",
		Value: c.Text.Offset,
	}
	c.Labels["__text_end"] = &Label{
		Name:  "__text_end",
		Value: c.Text.Offset + c.Text.Size,
	}
	c.Labels["__boot_data_start"] = &Label{
		Name:  "__boot_data_start",
		Value: c.BootData.Offset,
	}
	c.Labels["__boot_data_end"] = &Label{
		Name:  "__boot_data_end",
		Value: c.BootData.Offset + c.BootData.Size,
	}
	c.Labels["__data_start"] = &Label{
		Name:  "__data_start",
		Value: c.Data.Offset,
	}
	c.Labels["__data_end"] = &Label{
		Name:  "__data_end",
		Value: c.Data.Offset + c.Data.Size,
	}
	c.Labels["__bss_start"] = &Label{
		Name:  "__bss_start",
		Value: c.Bss.Offset,
	}
	c.Labels["__bss_end"] = &Label{
		Name:  "__bss_end",
		Value: c.Bss.Offset + c.Bss.Size,
	}
	c.Labels["__stack_start"] = &Label{
		Name:  "__stack_start",
		Value: c.Stack.Offset,
	}
	c.Labels["__stack_end"] = &Label{
		Name:  "__stack_end",
		Value: c.Stack.Offset + c.Stack.Size,
	}

	return nil
}

func (c *Compiler) genGlobals() error {
	for _, stmnt := range c.program {
		switch s := stmnt.(type) {
		case StmntGlobal:
			name := s.Label.Lexeme
			label, ok := c.Labels[name]
			if !ok {
				return GenericTokenError{s.Label, ""}
			}
			label.Global = true
		}
	}
	return nil
}

func (c *Compiler) compileAll() error {
	// create binaries for each section
	c.Boot.Bin = make([]byte, 0)
	c.Text.Bin = make([]byte, 0)
	c.Data.Bin = make([]byte, 0)
	c.Bss.Bin = make([]byte, 0)
	c.CurrentSection = &c.Text

	for _, stmnt := range c.program {
		switch s := stmnt.(type) {
		case StmntDirective:
			c.setSection(s.Directive.TokenType)
		}
		hereVal := c.CurrentSection.Offset + len(c.CurrentSection.Bin)
		here := Label{
			Name:  ".",
			Value: hereVal,
		}
		c.Labels["."] = &here
		bin, err := stmnt.Compile(c.Labels)
		if err != nil {
			return err
		}
		c.CurrentSection.Bin = append(c.CurrentSection.Bin, bin...)
	}

	return nil
}

func (c *Compiler) validateSections() error {
	errs := error(nil)
	err := c.Boot.Validate(".boot")
	if err != nil {
		errs = errors.Join(errs, err)
	}
	err = c.Text.Validate(".text")
	if err != nil {
		errs = errors.Join(errs, err)
	}
	err = c.BootData.Validate(".boot.data")
	if err != nil {
		errs = errors.Join(errs, err)
	}
	err = c.Data.Validate(".data")
	if err != nil {
		errs = errors.Join(errs, err)
	}
	err = c.Bss.Validate(".bss")
	if err != nil {
		errs = errors.Join(errs, err)
	}

	for _, b := range c.Bss.Bin {
		if b != 0 {
			errs = errors.Join(errs, fmt.Errorf(".bss section contains non-zero data"))
			break
		}
	}
	return errs
}

func (c *Compiler) setSection(t token.Type) {
	switch t {
	case token.Boot:
		c.CurrentSection = &c.Boot
	case token.Text:
		c.CurrentSection = &c.Text
	case token.BootData:
		c.CurrentSection = &c.BootData
	case token.Data:
		c.CurrentSection = &c.Data
	case token.Bss:
		c.CurrentSection = &c.Bss
	}
}

func (c *Compiler) buildBinary() ([]byte, error) {
	b := make([]byte, 0)

	// build header
	magic := 0x00706c75
	// the ".text" section starts at 12 within the binary,
	// the section will be loaded at 0 within ram though
	textAddr := 12
	textSize := c.Boot.Size + c.Text.Size
	dataSize := c.BootData.Size + c.Data.Size
	bssSize := c.Bss.Size + c.Stack.Size
	b = append(b, byteInt(magic)...)
	b = append(b, byteShort(textAddr)...)
	b = append(b, byteShort(textSize)...)
	b = append(b, byteShort(dataSize)...)
	b = append(b, byteShort(bssSize)...)

	// append the rest
	b = append(b, c.Boot.Bin...)
	b = append(b, c.Text.Bin...)
	b = append(b, c.BootData.Bin...)
	b = append(b, c.Data.Bin...)
	// b = append(b, c.Bss.Bin...)
	// b = append(b, c.Stack.Bin...) // data not actually placed here

	return b, nil
}

func byteShort(i int) []byte {
	return []byte{
		byte(i),
		byte(i >> 8),
	}
}

func byteInt(i int) []byte {
	return []byte{
		byte(i),
		byte(i >> 8),
		byte(i >> 16),
		byte(i >> 24),
	}
}

// func ulpAddr(addr int) int {
// 	return addr / 4
// }
