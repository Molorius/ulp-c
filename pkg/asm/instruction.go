package asm

import (
	"fmt"

	"github.com/Molorius/ulp-c/pkg/asm/token"
)

type validateInsHelperStruct struct {
	isReg  bool
	isExpr bool
	isJump bool
}

func validateIns(s StmntInstr, v []validateInsHelperStruct) error {
	if len(s.Args) != len(v) {
		return InstrArgCountError{s.Instruction, fmt.Sprintf("%d", len(s.Args)), len(v)}
	}
	errs := error(nil)
	for i := range v {
		reg := v[i].isReg && s.Args[i].IsReg()
		exp := v[i].isExpr && s.Args[i].IsExpr()
		jmp := v[i].isJump && s.Args[i].IsJump()
		if !(reg || exp || jmp) {
			return InstrArgTypeError{s, i}
		}
	}

	return errs
}

func (s *StmntInstr) validate() error {
	// nArgMsg := "incorrect number of arguments"
	switch s.Instruction.TokenType {
	case token.Add, token.Sub, token.And, token.Or, token.Lsh, token.Rsh:
		return validateIns(*s, []validateInsHelperStruct{
			{isReg: true},
			{isReg: true},
			{isReg: true, isExpr: true},
		})
	case token.Move:
		return validateIns(*s, []validateInsHelperStruct{
			{isReg: true},
			{isReg: true, isExpr: true},
		})
	case token.St, token.Ld:
		return validateIns(*s, []validateInsHelperStruct{
			{isReg: true},
			{isExpr: true},
		})
	case token.Jump:
		// jump has an optional parameter, try with that first
		e := validateIns(*s, []validateInsHelperStruct{
			{isExpr: true, isReg: true},
			{isJump: true},
		})
		if e != nil {
			// then try without it
			e = validateIns(*s, []validateInsHelperStruct{
				{isExpr: true, isReg: true},
			})
		}
		return e
	case token.Jumpr, token.Jumps:
		return validateIns(*s, []validateInsHelperStruct{
			{isExpr: true},
			{isExpr: true},
			{isJump: true},
		})
	case token.StageInc, token.StageDec, token.Sleep, token.Wait:
		return validateIns(*s, []validateInsHelperStruct{
			{isExpr: true},
		})
	case token.Adc:
		return validateIns(*s, []validateInsHelperStruct{
			{isReg: true},
			{isExpr: true},
			{isExpr: true},
		})
	case token.I2cRd, token.RegWr:
		return validateIns(*s, []validateInsHelperStruct{
			{isExpr: true},
			{isExpr: true},
			{isExpr: true},
			{isExpr: true},
		})
	case token.I2cWr:
		return validateIns(*s, []validateInsHelperStruct{
			{isExpr: true},
			{isExpr: true},
			{isExpr: true},
			{isExpr: true},
			{isExpr: true},
		})
	case token.RegRd:
		return validateIns(*s, []validateInsHelperStruct{
			{isExpr: true},
			{isExpr: true},
			{isExpr: true},
		})
	case token.StageRst, token.Halt, token.Wake:
		return validateIns(*s, []validateInsHelperStruct{})
	default:
		return UnknownTokenError{s.Instruction}
	}
}
