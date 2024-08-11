package emu

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

type UlpEmu struct {
	R          [4]uint16        // registers R0 through R3
	Overflow   bool             // overflow flag
	Zero       bool             // zero flag
	SC         uint8            // stage count register
	Memory     [8176 / 4]uint32 // memory visible to the ulp
	IP         uint16           // instruction pointer
	Wake       bool             // esp32 wake indicator
	cycles     uint64           // number of cycles executed
	dataOffset int
}

func (u *UlpEmu) LoadBinary(bin []uint8) error {
	// check header
	header := []uint8{'u', 'l', 'p', 0}
	if !reflect.DeepEqual(header, bin[0:4]) {
		return fmt.Errorf("invalid header: %v", header)
	}
	text_offset_binary := int(bin[4]) | (int(bin[5]) << 8)
	text_size := int(bin[6]) | (int(bin[7]) << 8)
	u.dataOffset = text_size / 4
	// clear memory
	for i := 0; i < len(u.Memory); i++ {
		u.Memory[i] = 0
	}
	// load binary
	code := bin[text_offset_binary:]
	for i := 0; i < len(code)/4; i++ {
		j := i * 4
		u.Memory[i] = binary.LittleEndian.Uint32(code[j : j+4])
	}
	u.IP = 0     // this is just a convention
	u.cycles = 0 // reset the cycles
	return nil
}

func (u *UlpEmu) Tick() error {
	instr := u.Fetch()
	return u.DecodeExecute(instr)
}

func (u *UlpEmu) Fetch() uint32 {
	intsr := u.Memory[u.IP]
	return intsr
}

func (u *UlpEmu) DecodeExecute(instr uint32) error {
	op := bitRead(instr, 28, 4)
	subOp := bitRead(instr, 25, 3)
	switch op {
	case 7: // operations
		u.cycles += 6 // 2 to execute, 4 to fetch next
		u.IP++
		rdst := bitRead(instr, 0, 2)
		rsrc1 := bitRead(instr, 2, 2)
		val1 := uint32(u.R[rsrc1])
		aluSel := bitRead(instr, 21, 4)
		switch subOp {
		case 0: // operations among registers
			rsrc2 := bitRead(instr, 4, 2)
			val2 := uint32(u.R[rsrc2])
			switch aluSel {
			case 0: // add
				out := val1 + val2
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
				u.Overflow = out > 0xFFFF
			case 1: // sub
				out := val1 - val2
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
				u.Overflow = out > 0xFFFF
			case 2: // and
				out := val1 & val2
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			case 3: // or
				out := val1 | val2
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			case 4: // move
				out := val1
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			case 5: // lsh
				out := val1 << val2
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			case 6: // rsh
				out := val1 >> val2
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			default:
				return fmt.Errorf("unknown ALU reg aluSel: %v", aluSel)
			}
		case 1: // operations with immediate
			imm := bitRead(instr, 4, 16)
			switch aluSel {
			case 0: // add
				out := val1 + imm
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
				u.Overflow = out > 0xFFFF
			case 1: // sub
				out := val1 - imm
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
				u.Overflow = out > 0xFFFF
			case 2: // and
				out := val1 & imm
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			case 3: // or
				out := val1 & imm
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			case 4: // move
				out := imm
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			case 5: // lsh
				out := val1 << imm
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			case 6: // rsh
				out := val1 >> imm
				out16 := uint16(out)
				u.R[rdst] = out16
				u.Zero = out16 == 0
			default:
				return fmt.Errorf("unknown ALU immediate aluSel: %v", aluSel)
			}
		case 2: // stage count
			imm := uint8(bitRead(instr, 4, 8))
			switch aluSel {
			case 0: // stage increase
				u.SC += imm
			case 1: // stage decrease
				u.SC -= imm
			case 2: // stage reset
				u.SC = 0
			default:
				return fmt.Errorf("unknown stage count aluSel: %v", aluSel)
			}
		default:
			return fmt.Errorf("unknown operation subOp %v", subOp)
		}
	case 6: // store
		rsrc := bitRead(instr, 0, 2)
		rdst := bitRead(instr, 2, 2)
		offset := bitRead(instr, 10, 11)
		upper := uint32(u.IP << 5)
		upper |= uint32(rdst)
		lower := uint32(u.R[rsrc])
		value := (upper << 16) | lower
		address := u.R[rdst] + uint16(offset)
		address = address & 0x7FF
		u.Memory[address] = value
		u.IP++
		u.cycles += 8
	case 13: // load
		rdst := bitRead(instr, 0, 2)
		rsrc := bitRead(instr, 2, 2)
		offset := bitRead(instr, 10, 11)
		address := u.R[rsrc] + uint16(offset)
		address = address & 0x7FF
		value := u.Memory[address]
		u.R[rdst] = uint16(value)
		u.IP++
		u.cycles += 8 // 4 execute + 4 fetch
	case 8: // jump
		u.cycles += 4 // 2 execute + 2 fetch
		switch subOp {
		case 0: // jump
			rdst := bitRead(instr, 0, 2)
			immAddr := bitRead(instr, 2, 12)
			sel := bitRead(instr, 21, 1)
			jumpType := bitRead(instr, 22, 3)
			jumpAddr := uint16(immAddr)
			if sel == 1 {
				jumpAddr = u.R[rdst]
			}
			switch jumpType {
			case 0: // unconditional jump
				u.IP = jumpAddr
			case 1: // jump if zero flag set
				if u.Zero {
					u.IP = jumpAddr
				} else {
					u.IP += 1
				}
			case 2: // jump if overflow flag set
				if u.Overflow {
					u.IP = jumpAddr
				} else {
					u.IP += 1
				}
			default:
				return fmt.Errorf("unknown jump type %v", jumpType)
			}
		case 1: // jumpr
			threshold := bitRead(instr, 0, 16)
			cond := bitRead(instr, 16, 1)
			step := bitRead(instr, 17, 8)
			newIp := u.IP + uint16(step)
			value := u.R[0]
			if (step >> 7) == 1 {
				newIp = u.IP - uint16(step&0x3F) // 6 bits
			}
			shouldJump := value < uint16(threshold)
			if cond == 1 {
				shouldJump = !shouldJump
			}
			if shouldJump {
				u.IP = newIp
			} else {
				u.IP += 1
			}
		case 2: // jumps
			threshold := uint8(bitRead(instr, 0, 8))
			cond := bitRead(instr, 15, 2)
			step := bitRead(instr, 17, 8)
			newIp := u.IP + uint16(step)
			if (step >> 7) == 1 {
				newIp = u.IP - uint16(step&0x3F) // 6 bits
			}
			var shouldJump bool
			switch cond {
			case 0:
				shouldJump = u.SC < threshold
			case 1:
				shouldJump = u.SC >= threshold
			default:
				shouldJump = u.SC <= threshold
			}
			if shouldJump {
				u.IP = newIp
			} else {
				u.IP += 1
			}
		default:
			return fmt.Errorf("unknown jump subOp %v", subOp)
		}
	case 9: // wake
		u.Wake = true
		u.IP++
		u.cycles += 6 // 2 execute 4 fetch
	default:
		return fmt.Errorf("unknown operation %v", op)
	}
	return nil
}

func bitRead(num uint32, offset uint, size uint) uint32 {
	val := num >> offset
	mask := uint32((1 << size) - 1)
	return val & mask
}

func (u *UlpEmu) RunWithSystem(maxCycles uint64) (string, error) {
	out := ""
	prev := uint32(0)
	for {
		if u.cycles >= maxCycles {
			return out, fmt.Errorf("exceeded max cycles: %s", out)
		}
		err := u.Tick()
		if err != nil {
			return out, fmt.Errorf("emulation error: %s", err)
		}
		mutexFlag0 := u.Memory[u.dataOffset+0] & 0xFFFF
		if prev == 1 && mutexFlag0 == 0 { // if the ulp just gave the mutex
			fn := uint16(u.Memory[u.dataOffset+3] & 0xFFFF)
			param := uint16(u.Memory[u.dataOffset+4] & 0xFFFF)
			u.Memory[u.dataOffset+3] = 0 // acknowledge it
			switch fn {
			case 0: // ack
				break
			case 1: // done
				return out, nil
			case 2: // printU16
				out += fmt.Sprintf("%d ", param)
			case 3: // printChar
				out += fmt.Sprintf("%c", param&0xFF)
			default:
				return out, fmt.Errorf("unknown ulp sys function: %d", fn)
			}
		}
		prev = mutexFlag0
	}
}
