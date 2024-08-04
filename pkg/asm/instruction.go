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
		return InstrArgCountError{s.Instruction, fmt.Sprintf("%d", len(v)), len(s.Args)}
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
			{isReg: true},
			{isExpr: true},
		})
	case token.Jump:
		// jump has an optional parameter, try without that first
		e := validateIns(*s, []validateInsHelperStruct{
			{isExpr: true, isReg: true},
		})

		if e != nil {
			// then try with it
			e = validateIns(*s, []validateInsHelperStruct{
				{isExpr: true, isReg: true},
				{isJump: true},
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

func (s StmntInstr) Compile(labels map[string]*Label) ([]byte, error) {
	s.labels = &labels
	switch s.Instruction.TokenType {
	case token.Add:
		return s.compileAlu(0)
	case token.Sub:
		return s.compileAlu(1)
	case token.And:
		return s.compileAlu(2)
	case token.Or:
		return s.compileAlu(3)
	case token.Move:
		return s.compileMove()
	case token.Lsh:
		return s.compileAlu(5)
	case token.Rsh:
		return s.compileAlu(6)
	case token.St:
		return s.compileMemory(6, 0b100)
	case token.Ld:
		return s.compileMemory(13, 0)
	case token.Jump:
		return s.compileJump()
	default:
		return nil, GenericTokenError{s.Instruction, "instruction not implemented for compile, please file a bug report"}
	}
}

func (s *StmntInstr) compileAlu(aluSel int) ([]byte, error) {
	rdst, err := s.Args[0].(ArgReg).Evaluate()
	if err != nil {
		return nil, err
	}
	rsrs1, err := s.Args[1].(ArgReg).Evaluate()
	if err != nil {
		return nil, err
	}
	// imm := 0
	subOp := 0
	imm, isReg, err := evalArgOrReg(s.Args[2], *s.labels)
	if !isReg {
		subOp = 1
	}
	if err != nil {
		return nil, err
	}
	return insStandard(7, subOp, aluSel, imm, rsrs1, rdst), nil
}

func evalArgOrReg(arg Arg, labels map[string]*Label) (int, bool, error) {
	isReg := false
	imm := 0
	err := error(nil)
	switch third := arg.(type) {
	case ArgReg:
		imm, err = third.Evaluate()
		if err != nil {
			return 0, false, err
		}
		isReg = true
	case ArgExpr:
		imm, err = third.Expr.Evaluate(labels)
		if err != nil {
			return 0, false, err
		}
	default:
		return 0, false, fmt.Errorf("could not evaluate in evalArgOrReg(), please file a bug report")
	}
	return imm, isReg, err
}

func (s *StmntInstr) compileMove() ([]byte, error) {
	rdst, err := s.Args[0].(ArgReg).Evaluate()
	if err != nil {
		return nil, err
	}
	val, isReg, err := evalArgOrReg(s.Args[1], *s.labels)
	if err != nil {
		return nil, err
	}
	val = bitMask(val, 16)
	imm := 0
	rsrc1 := 0
	subOp := 0
	if !isReg {
		subOp = 1
		imm = val
	} else {
		rsrc1 = val
	}
	imm = bitMask(imm, 16)
	return insStandard(7, subOp, 4, imm, rsrc1, rdst), nil
}

func (s *StmntInstr) compileMemory(op int, subOp int) ([]byte, error) {
	rA, err := s.Args[0].(ArgReg).Evaluate()
	if err != nil {
		return nil, err
	}
	rB, err := s.Args[1].(ArgReg).Evaluate()
	if err != nil {
		return nil, err
	}
	offset, err := s.Args[2].(ArgExpr).Expr.Evaluate(*s.labels)
	if err != nil {
		return nil, err
	}
	return insMemory(op, subOp, offset, rA, rB), nil
}

func (s *StmntInstr) compileJump() ([]byte, error) {
	val, isReg, err := evalArgOrReg(s.Args[0], *s.labels)
	if err != nil {
		return nil, err
	}
	sel := 1 // register
	if !isReg {
		sel = 0 // immediate
	}
	jumpType := 0 // undonditional jump
	// TODO: add other jump types
	if len(s.Args) > 1 {
		return nil, GenericTokenError{s.Instruction, "condition operand not supported yet"}
	}
	op := 8
	subOp := 0
	return insJump(op, subOp, jumpType, sel, val), nil
}

func insStandard(op int, subOp int, aluSel int, imm int, rA int, rB int) []byte {
	ins := bitMask(op, 4) << 28
	ins |= bitMask(subOp, 3) << 25
	ins |= bitMask(aluSel, 4) << 21
	ins |= bitMask(imm, 17) << 4
	ins |= bitMask(rA, 2) << 2
	ins |= bitMask(rB, 2)
	return byteInt(ins)
}

func insJump(op int, subOp int, jumpType int, sel int, arg int) []byte {
	ins := bitMask(op, 4) << 28
	ins |= bitMask(subOp, 3) << 25
	ins |= bitMask(jumpType, 3) << 22
	ins |= bitMask(sel, 1) << 21
	if sel == 0 {
		ins |= bitMask(arg, 11) << 2
	} else {
		ins |= bitMask(arg, 2)
	}
	return byteInt(ins)
}

func insMemory(op int, subOp int, offset int, rA int, rB int) []byte {
	ins := bitMask(op, 4) << 28
	ins |= bitMask(subOp, 3) << 25
	ins |= bitMask(offset, 11) << 10
	ins |= bitMask(rB, 2) << 2
	ins |= bitMask(rA, 2)
	return byteInt(ins)
}

func bitMask(val int, bits int) int {
	// create the bit mask
	mask := 0
	for i := 0; i < bits; i++ {
		mask <<= 1
		mask |= 1
	}
	return val & mask
}
