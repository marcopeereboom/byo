package bus

import "errors"

var (
	ErrRegionNotFound = errors.New("region not found")
)

// Buser is the required interface for all peripherals.
type Buser interface {
	Read(uint64, uint64) []byte
	Write(uint64, []byte)

	//BusError()
	Reset(bool)
	//Interrupt()

	Length() uint64
}

type buser struct {
	Buser

	start uint64
	end   uint64
}

// Bus is the glue for all peripherals.
type Bus struct {
	Buser
	peripherals []*buser
}

// New returns a new Bus sturcture.  Note that the caller is responsible for
// concurrency, this is by design for performance reasons.
func New() (*Bus, error) {
	return &Bus{}, nil
}

// Attach a peripheral on provided Bus.  On success it retuns a peripheral ID
// that may be used later for lookup.
func (b *Bus) Attach(address uint64, peripheral Buser) (int, error) {
	p := &buser{
		start: address,
		end:   address + peripheral.Length(),
		Buser: peripheral,
	}
	b.peripherals = append(b.peripherals, p)
	return len(b.peripherals) - 1, nil
}

// Reset sends reset to all peripherals.
func (b *Bus) Reset(powerOn bool) {
	for _, v := range b.peripherals {
		v.Reset(powerOn)
	}
}

// Lookup translates address to peripheral ID.
func (b *Bus) Lookup(address uint64) (int, error) {
	for k, v := range b.peripherals {
		if v.start <= address && v.end >= address {
			return k, nil
		}
	}
	return -1, ErrRegionNotFound
}

// Read from peripheral at provided address.  Address is always looked up in
// the peripheral list.  It is therefore recommended to use ReadID.
func (b *Bus) Read(address uint64, length uint64) []byte {
	id, err := b.Lookup(address)
	if err != nil {
		// assert bus error
		panic("bus error")
	}
	return b.ReadID(id, address, length)
}

// Write to peripheral at provided address.  Address is always looked up in
// the peripheral list.  It is therefore recommended to use WriteID.
func (b *Bus) Write(address uint64, data []byte) {
	id, err := b.Lookup(address)
	if err != nil {
		// assert bus error
		panic("bus error")
	}
	b.WriteID(id, address, data)
}

// ReadID read from peripheral id at provided address.
func (b *Bus) ReadID(id int, address uint64, length uint64) []byte {
	p := b.peripherals[id]
	return p.Read(address-p.start, length)
}

// WriteID write to peripheral id at provided address.
func (b *Bus) WriteID(id int, address uint64, data []byte) {
	p := b.peripherals[id]
	p.Write(address-p.start, data)
}
