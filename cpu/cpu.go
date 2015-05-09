package cpu

import "errors"

type CPUer interface {
	Reset()      // reset cpu
	Interrupt()  // interrupt cpu
	Step() error // execute next instruction
}

var (
	ErrInvalidOpcode = errors.New("invalid opcode")
)
