package asm

import (
	"errors"
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
	case token.StageInc:
		return s.compileStage(0)
	case token.StageDec:
		return s.compileStage(1)
	case token.StageRst:
		return s.compileStage(2)
	case token.St:
		return s.compileMemory(6, 0b100)
	case token.Ld:
		return s.compileMemory(13, 0)
	case token.Jump:
		return s.compileJump()
	case token.Jumpr:
		return s.compileJumpr()
	case token.Jumps:
		return s.compileJumps()
	case token.Halt:
		return s.compileNoParams(11, 0, 0)
	case token.Wake:
		return s.compileNoParams(9, 0, 1)
	case token.Sleep:
		return s.compileSingleParam(9, 1)
	case token.Wait:
		return s.compileSingleParam(4, 0)
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

func (s *StmntInstr) compileStage(aluSel int) ([]byte, error) {
	imm := 0
	err := error(nil)
	if aluSel != 2 {
		imm, err = s.Args[0].(ArgExpr).Expr.Evaluate(*s.labels)
		if err != nil {
			return nil, err
		}
	}
	return insStage(aluSel, imm), nil
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
	if len(s.Args) > 1 {
		argToken := s.Args[1].(ArgJump).Arg
		switch argToken.TokenType {
		case token.Eq:
			jumpType = 1
		case token.Ov:
			jumpType = 2
		default:
			return nil, GenericTokenError{argToken, "unsupported jump type for jump instruction"}
		}
	}
	op := 8
	subOp := 0
	return insJump(op, subOp, jumpType, sel, val), nil
}

func (s *StmntInstr) compileJumpr() ([]byte, error) {
	step, threshold, argToken, err := s.argsJumpRS()
	if err != nil {
		return nil, err
	}
	err = s.stepValidate(step)
	if err != nil {
		return nil, err
	}
	threshold &= 0xFFFF // mask it off for later
	ge := 1
	lt := 0
	switch argToken.TokenType {
	case token.Eq:
		err = s.stepValidate(step - 1)
		if err != nil {
			return nil, errors.Join(GenericTokenError{argToken, "step-1 outside of bounds for this condition"}, err)
		}
		if threshold == 0xFFFF { // rollover causes problems, only check ge
			// we will mess up addresses if we return 1 instruction so instead
			// use a slightly faster 2-instruction sequence
			return append(insJumpr(2, lt, 0xFFFF), insJumpr(step-1, ge, 0xFFFF)...), nil
		}
		// `step-1` because we need to account for the first instruction's offset
		return append(insJumpr(2, ge, threshold+1), insJumpr(step-1, ge, threshold)...), nil
	case token.Lt:
		return insJumpr(step, lt, threshold), nil
	case token.Le:
		if threshold == 0xFFFF { // always true
			return insJumpr(step, ge, 0), nil
		}
		return insJumpr(step, lt, threshold+1), nil
	case token.Gt:
		if threshold == 0xFFFF { // always false
			return insJumpr(step, lt, 0), nil
		}
		return insJumpr(step, ge, threshold+1), nil
	case token.Ge:
		return insJumpr(step, ge, threshold), nil
	default:
		return nil, GenericTokenError{argToken, "unsupported jump type for jumpr instruction"}
	}
}

func stepFix(step int) int {
	if step < 0 {
		step = -step
		step |= 1 << 7
	}
	return step
}

func (s *StmntInstr) argsJumpRS() (int, int, Token, error) {
	dest, err := s.Args[0].(ArgExpr).Expr.Evaluate(*s.labels)
	if err != nil {
		return 0, 0, Token{}, err
	}
	l, ok := (*s.labels)["."]
	if !ok {
		return 0, 0, Token{}, GenericTokenError{s.Instruction, "the \"Here\" token \".\" not set to get offset, please file a bug report"}
	}
	step := dest - (l.Value / 4)
	threshold, err := s.Args[1].(ArgExpr).Expr.Evaluate(*s.labels)
	if err != nil {
		return 0, 0, Token{}, err
	}
	argToken := s.Args[2].(ArgJump).Arg
	return step, threshold, argToken, nil
}

func (s *StmntInstr) stepValidate(step int) error {
	stepMax := 0b1111111
	if step > stepMax {
		return GenericTokenError{s.Instruction, fmt.Sprintf("step of %d is outside the maximum of %d", step, stepMax)}
	}
	if step < -stepMax {
		return GenericTokenError{s.Instruction, fmt.Sprintf("step of %d is outside the minimum of %d", step, -stepMax)}
	}
	return nil
}

func (s *StmntInstr) compileJumps() ([]byte, error) {
	step, threshold, argToken, err := s.argsJumpRS()
	if err != nil {
		return nil, err
	}
	err = s.stepValidate(step)
	if err != nil {
		return nil, err
	}
	threshold &= 0xFF // mask it off for later
	le := 2
	lt := 0
	ge := 1
	switch argToken.TokenType {
	case token.Eq:
		err = s.stepValidate(step - 1)
		if err != nil {
			return nil, errors.Join(GenericTokenError{argToken, "step-1 outside of bounds for this condition"}, err)
		}
		// `step-1` because we need to account for the first instruction's offset
		return append(insJumps(2, lt, threshold), insJumps(step-1, le, threshold)...), nil
	case token.Lt:
		return insJumps(step, lt, threshold), nil
	case token.Le:
		return insJumps(step, le, threshold), nil
	case token.Gt:
		if threshold == 0xFF {
			return insJumps(step, lt, 0), nil // never jump
		}
		return insJumps(step, ge, threshold+1), nil
	case token.Ge:
		return insJumps(step, ge, threshold), nil
	default:
		return nil, GenericTokenError{argToken, "unsupported jump type for jumpr instruction"}
	}
}

func (s *StmntInstr) compileNoParams(op int, subOp int, param int) ([]byte, error) {
	return insSingleParam(op, subOp, param), nil
}

func (s *StmntInstr) compileSingleParam(op int, subOp int) ([]byte, error) {
	imm, err := s.Args[0].(ArgExpr).Expr.Evaluate(*s.labels)
	if err != nil {
		return nil, err
	}
	return insSingleParam(op, subOp, imm), nil
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

func insStage(aluSel int, imm int) []byte {
	op := 7
	subOp := 2
	ins := bitMask(op, 4) << 28
	ins |= bitMask(subOp, 3) << 25
	ins |= bitMask(aluSel, 4) << 21
	ins |= bitMask(imm, 8) << 4
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

func insJumpr(step int, cond int, threshold int) []byte {
	op := 8
	subOp := 1
	step = stepFix(step)
	ins := bitMask(op, 4) << 28
	ins |= bitMask(subOp, 3) << 25
	ins |= bitMask(step, 8) << 17
	ins |= bitMask(cond, 1) << 16
	ins |= bitMask(threshold, 16)
	return byteInt(ins)
}

func insJumps(step int, cond int, threshold int) []byte {
	op := 8
	subOp := 2
	step = stepFix(step)
	ins := bitMask(op, 4) << 28
	ins |= bitMask(subOp, 3) << 25
	ins |= bitMask(step, 8) << 17
	ins |= bitMask(cond, 2) << 15
	ins |= bitMask(threshold, 8)
	return byteInt(ins)
}

func insSingleParam(op int, subOp int, param int) []byte {
	ins := bitMask(op, 4) << 28
	ins |= bitMask(subOp, 3) << 25
	ins |= bitMask(param, 16)
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
