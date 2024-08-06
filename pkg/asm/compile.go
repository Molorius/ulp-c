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

func (c *Compiler) Compile(program []Stmnt, reservedBytes int) ([]byte, error) {
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
		return nil, err
	}
	err = c.genLabels(reservedBytes)
	if err != nil {
		return nil, err
	}
	err = c.genGlobals()
	if err != nil {
		return nil, err
	}
	err = c.compileAll()
	if err != nil {
		return nil, err
	}
	err = c.validateSections()
	if err != nil {
		return nil, err
	}
	bin, err := c.buildBinary()
	return bin, err
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

func (c *Compiler) genLabels(reservedBytes int) error {
	// resolve section offsets
	c.Boot.Offset = 0
	c.Text.Offset = c.Boot.Offset + c.Boot.Size
	c.BootData.Offset = c.Text.Offset + c.Text.Size
	c.Data.Offset = c.BootData.Offset + c.BootData.Size
	c.Bss.Offset = c.Data.Offset + c.Data.Size
	// data is never placed in stack, calculate remaining memory
	c.Stack.Offset = c.Bss.Offset + c.Bss.Size
	c.Stack.Size = reservedBytes - c.Stack.Offset
	if c.Stack.Size < 0 {
		total := c.Stack.Offset
		return fmt.Errorf("overflowing the %d reserved bytes: .boot=%d .boot.data=%d .text=%d .data=%d .bss=%d total=%d",
			reservedBytes, c.Boot.Size, c.BootData.Size, c.Text.Size, c.Data.Size, c.Bss.Size, total)
	}

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
