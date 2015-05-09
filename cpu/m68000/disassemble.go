package m68000

import (
	"fmt"

	"github.com/marcopeereboom/byo/cpu"
)

// ccr translates the ccr into human readable form.  This should be moved into
// a monitor file.
func (c *m68k) ccr() string {
	b := make([]byte, 5)
	for i := 0; i < len(b); i++ {
		b[i] = '-'
	}

	if c.sr&carry == carry {
		b[4] = 'C'
	}
	if c.sr&overflow == overflow {
		b[3] = 'V'
	}
	if c.sr&zero == zero {
		b[2] = 'Z'
	}
	if c.sr&negative == negative {
		b[1] = 'N'
	}
	if c.sr&extend == extend {
		b[0] = 'X'
	}

	return string(b)
}

func (c *m68k) disassemble(address uint32) (string, int, error) {
	opcode := c.read16(address)
	i, found := opcodes[opcode]
	if !found {
		return "INVALID", 0, cpu.ErrInvalidOpcode
	}

	operand := i.fetchOperand(c, c.pc+2, i.operandSize)

	return i.disassemble(c, opcode, operand)
}

func disassembleSD(mrMode bool, opcode uint16, bits uint16) string {
	var m, r uint16
	if mrMode {
		m, r = mr(opcode, bits)
	} else {
		m, r = rm(opcode, bits)
	}

	switch m {
	case 0x00:
		// r = Dn
		return fmt.Sprintf("d%v", r)
	case 0x01:
		// r = An
		return fmt.Sprintf("a%v", r)
	case 0x02:
		// r = An -> (An)
		return fmt.Sprintf("(a%v)", r)
	case 0x03:
		// r = An -> (An)+
		return fmt.Sprintf("(a%v)+", r)
	case 0x04:
		// r = An -> -(An)
		return fmt.Sprintf("-(a%v)", r)
	case 0x05:
		// r = An -> (d16,An)
	case 0x06:
		// r = An -> (d8,An,Xn)
	case 0x07:
		// r == 0x00 -> (xxx).W
		// r == 0x01 -> (xxx).L
	}

	return fmt.Sprintf("unhandled %v 0x%x", mrMode, m)
}

func size2Text(size uint16) string {
	switch size {
	case 0x01:
		return "b"
	case 0x02:
		return "l"
	case 0x03:
		return "w"
	}
	return "invalid"
}

// decodeSize returns encoded size
func decodeSize(opc uint16, bits uint16) (s uint16) {
	s = opc >> (bits - 1) & 0x03
	return
}

func disassembleSize(opc uint16, bits uint16) string {
	return size2Text(decodeSize(opc, bits))
}

// rm decodes and returns Register Mode.
func rm(opc uint16, bits uint16) (m, r uint16) {
	// encoded as RRRMMM
	m = (opc >> (bits - 5)) & 0x07
	r = opc >> (bits - 2) & 0x07
	return
}

// mr decodes and returns Mode Register.
func mr(opc uint16, bits uint16) (m, r uint16) {
	// encoded as MMMRRR
	r = opc >> (bits - 5) & 0x07
	m = opc >> (bits - 2) & 0x07
	return
}

func dMove(c *m68k, opcode uint16, operand []byte) (string, int, error) {
	source := disassembleSD(true, opcode, 5)
	dest := disassembleSD(false, opcode, 11)
	size := "." + disassembleSize(opcode, 13)
	s := fmt.Sprintf("move%v\t%v,%v", size, source, dest)
	return s, 2 + len(operand), nil
}

func dAdd(c *m68k, opcode uint16, operand []byte) (string, int, error) {
	var sz, source, dest string

	register, opmode := mr(opcode, 11) // this is right
	// page 4-5
	switch opmode {
	// <ea> + Dn -> Dn
	case 0x0:
		// byte
		sz = ".b"
		source = disassembleSD(true, opcode, 5)
		dest = fmt.Sprintf("d%v", register)
	case 0x1:
		// word
		sz = ".w"
		source = disassembleSD(true, opcode, 5)
		dest = fmt.Sprintf("d%v", register)
	case 0x2:
		// long
		sz = ".l"
		source = disassembleSD(true, opcode, 5)
		dest = fmt.Sprintf("d%v", register)

	// Dn + <ea> -> <ea>
	case 0x4:
		// byte
		sz = ".b"
		source = fmt.Sprintf("d%v", register)
		dest = disassembleSD(true, opcode, 5)
	case 0x5:
		// word
		sz = ".w"
		source = fmt.Sprintf("d%v", register)
		dest = disassembleSD(true, opcode, 5)
	case 0x6:
		// long
		sz = ".l"
		source = fmt.Sprintf("d%v", register)
		dest = disassembleSD(true, opcode, 5)

	// adda An + <ea> -> <ea>
	case 0x3:
		// word
		sz = ".w"
		source = disassembleSD(true, opcode, 5)
		dest = fmt.Sprintf("a%v", register)
	case 0x7:
		// long
		sz = ".l"
		source = disassembleSD(true, opcode, 5)
		dest = fmt.Sprintf("a%v", register)

	default:
		sz = ".inv"
		source = "inv"
		dest = "inv"
	}

	s := fmt.Sprintf("add%v\t%v,%v", sz, source, dest)
	return s, 2 + len(operand), nil
}
