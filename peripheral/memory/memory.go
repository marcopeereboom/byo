package memory

import (
	"errors"
	"os"
)

var (
//_ peripheral.Peripheraler = (*Memory)(nil) // ensure interface is satisfied
)

type Mode int

const (
	RAM Mode = iota
	ROM
	RAMBacked
)

var (
	ErrROMOverflow = errors.New("ROM overflow")
)

type Memory struct {
	backing []byte
	rommem  []byte
	rp      []byte // read pointer
	wp      []byte // write pointer
	mode    Mode
}

// NewRAMBackedROM returns a ROM that has backing memory.  If it is written to
// the backing memory will contain the written value.  However when the same
// location is read back it returns the ROM value.  The ROM can be however
// switched out to reflect the underlying RAM.
func NewRAMBackedROM(size int, filename string) (*Memory, error) {
	m, err := NewROM(size, filename)
	if err != nil {
		return nil, err
	}
	m.backing = make([]byte, size)
	m.wp = m.backing
	return m, nil
}

func (m *Memory) EnableBacking() {
	if m.mode != RAMBacked {
		panic("unsupported mode")
	}
	m.rp = m.backing
}

func (m *Memory) EnableROM() {
	if m.mode != RAMBacked {
		panic("unsupported mode")
	}
	m.rp = m.rommem
}

func NewROM(size int, filename string) (*Memory, error) {
	// ensure we can fit ROM in specified size
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Size() > int64(size) {
		return nil, ErrROMOverflow
	}

	// create memory structure
	m := Memory{
		mode:   ROM,
		rommem: make([]byte, size),
	}
	m.rp = m.rommem

	// obtain ROM image
	_, err = f.Read(m.rommem)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func NewRAM(size int) *Memory {
	m := Memory{
		mode:    RAM,
		backing: make([]byte, size),
	}
	m.wp = m.backing
	m.rp = m.backing
	return &m
}

func (m *Memory) Reset(powerOn bool) {
	if powerOn {
		m.backing = make([]byte, len(m.backing))

		if m.mode == RAMBacked {
			m.EnableROM()
		}
	}
}

func (m *Memory) Interrupt() {
}

func (m *Memory) Step() error {
	return nil // not implemented and not a failure
}

func (m *Memory) Read(address, size int) []byte {
	return m.rp[address : address+size]
}

func (m *Memory) Write(address int, data []byte) {
	copy(m.wp[address:], data)
}
