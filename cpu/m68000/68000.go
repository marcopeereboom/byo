package m68000

import (
	"encoding/binary"

	"github.com/marcopeereboom/byo/bus"
	"github.com/marcopeereboom/byo/cpu"
)

const (
	M68000 = "68000"

	carry    = 1 << 0
	overflow = 1 << 1
	zero     = 1 << 2
	negative = 1 << 3
	extend   = 1 << 4
)

var (
	_ cpu.CPUer = (*m68k)(nil) // ensure interface is satisfied

)

// m68k represents the Motorola 68000 CPU.  Note that this is a big endian CPU,
// however register content will be stored in native format.  Exception are the
// bit only adressable registers.
type m68k struct {
	// data registers
	d []uint32

	// address registers
	a []uint32

	// other registers
	pc uint32
	sr uint16 // user instructions may not touch upper 8 bits

	// bus
	bus *bus.Bus
}

// New returns a new m68k instance.
func New(bus *bus.Bus) (*m68k, error) {
	cpu := m68k{
		bus: bus,
		a:   make([]uint32, 8),
		d:   make([]uint32, 8),
	}
	return &cpu, nil
}

// Interrupt asserts the CPU's interrupt.  This is part of the CPUer interface.
func (c *m68k) Interrupt() {
}

// Reset asserts the CPU's reset.  This is part of the CPUer interface.  The
// 68000 CPU sets the SSP to the vector found in $0-$3 and the PC to the vector
// found in $4-$7.  These locations are usually shadowed by ROM.
func (c *m68k) Reset() {
	c.a[7] = c.read32(0)
	c.pc = c.read32(4)
}

// Step executes the next instruction on the CPU.  This is part of the CPUer
// interface.
func (c *m68k) Step() error {
	opcode := c.read16(c.pc)
	i, found := opcodes[opcode]
	if !found {
		return cpu.ErrInvalidOpcode
	}

	operand := i.fetchOperand(c, c.pc+2, i.operandSize)
	source := i.fetchSource(c, i.source, operand)
	destination := i.fetchDestination(c, i.destination, operand)
	intermediate := i.execute(c, source, destination, operand)
	i.storeDestination(c, i.destination, intermediate, operand)

	c.pc += 2 + uint32(len(operand))

	return nil
}

// write32 is a helper function to convert memory bytes into host endianess.
func (c *m68k) write32(address uint32, value uint32) {
	v := make([]byte, 4)
	binary.BigEndian.PutUint32(v, value)
	c.bus.Write(uint64(address), v)
}

// read32 is a helper function to convert memory bytes into host endianess.
func (c *m68k) read32(address uint32) uint32 {
	return binary.BigEndian.Uint32(c.bus.Read(uint64(address), 4))
}

// read16 is a helper function to convert memory bytes into host endianess.
func (c *m68k) read16(address uint32) uint16 {
	return binary.BigEndian.Uint16(c.bus.Read(uint64(address), 2))
}
