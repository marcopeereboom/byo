package peripheral

// Peripheraler is the interpace that all devices must comply to.
// Reset asserts the reset pin.  The argument indicates if this is a power on event.
type Peripheraler interface {
	Reset(bool) // reset peripheral
	Interrupt() // interrupt peripheral
	Step()      // execute next state

	Read8()  // read from peripheral
	Write8() // write to peripheral
}
