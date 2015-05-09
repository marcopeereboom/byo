package m68000

// instruction describes a motorola 68000 instruction in a way that it can be
// executed and disassembled.
type instruction struct {
	// disassmbly
	disassemble func(*m68k, uint16, []byte) (string, int, error)

	// execution
	operandSize  uint32
	fetchOperand func(*m68k, uint32, uint32) []byte

	fetchSource func(*m68k, uint32, []byte) uint32
	source      uint32

	fetchDestination func(*m68k, uint32, []byte) uint32
	storeDestination func(*m68k, uint32, uint32, []byte)
	destination      uint32

	execute func(*m68k, uint32, uint32, []byte) uint32
}

var (
	operandNop = []byte{} // to prevent GC
)

func (c *m68k) evalNL(d uint32) {
	// N
	if d&0x8000000 == 0 {
		c.sr &^= negative
	} else {
		c.sr |= negative
	}
}

func (c *m68k) evalZL(d uint32) {
	// Z
	if d == 0 {
		c.sr |= zero
	} else {
		c.sr &^= zero
	}
}

func (c *m68k) evalNZL(d uint32) {
	c.evalNL(d)
	c.evalZL(d)
}

func fetchOperandNop(c *m68k, address, size uint32) []byte {
	return operandNop
}

func fetchOperand(c *m68k, address, size uint32) []byte {
	return c.bus.Read(uint64(address), uint64(size))
}

func fetchNop(c *m68k, reg uint32, operand []byte) uint32 {
	return 0
}

func fetchDn(c *m68k, reg uint32, operand []byte) uint32 {
	return c.d[reg]
}

func fetchAn(c *m68k, reg uint32, operand []byte) uint32 {
	return c.a[reg]
}

func storeAn(c *m68k, reg, intermediate uint32, operand []byte) {
	c.a[reg] = intermediate
}

func storeAnIndirect(c *m68k, reg, intermediate uint32, operand []byte) {
	c.write32(c.a[reg], intermediate)
}

func movel(c *m68k, src uint32, dest uint32, operand []byte) uint32 {
	// set flags per page 3-18
	c.evalNZL(src)
	c.sr &^= overflow
	c.sr &^= carry

	return src
}

func addal(c *m68k, src uint32, dest uint32, operand []byte) uint32 {
	inter := src + dest

	// set flags per page 3-18
	v := src&dest&^inter | ^dest&inter
	if v&0x80000000 == 0 {
		c.sr &^= overflow
	} else {
		c.sr |= overflow
	}

	ca := src&dest | ^inter&dest | src&^inter
	if ca&0x80000000 == 0 {
		c.sr &^= carry
		c.sr &^= extend
	} else {
		c.sr |= carry
		c.sr |= extend
	}

	c.evalNZL(inter)

	return inter
}
