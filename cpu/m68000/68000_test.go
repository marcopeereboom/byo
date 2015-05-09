package m68000

import (
	"encoding/binary"
	"testing"

	"github.com/marcopeereboom/byo/bus"
	"github.com/marcopeereboom/byo/memory"
)

const (
	size    = 1024 * 1024 // 1M
	pcStart = 0x1000      // start PC
)

func newCpu() (*bus.Bus, *m68k) {
	b, err := bus.New()
	if err != nil {
		panic(err.Error())
	}
	ram := memory.NewRAM(size)
	_, err = b.Attach(0x0, ram)
	if err != nil {
		panic(err)
	}

	// setup ssp and pc
	ssp := make([]byte, 4)
	pc := make([]byte, 4)
	binary.BigEndian.PutUint32(ssp, 0x2000)
	binary.BigEndian.PutUint32(pc, pcStart)
	b.Write(0x0, ssp)
	b.Write(0x4, pc)

	// setup cpu
	c, err := New(b)
	if err != nil {
		panic(err)
	}

	// reset
	b.Reset(true)
	c.Reset()

	return b, c
}

func TestMOVEL(t *testing.T) {
	b, c := newCpu()
	b.Write(pcStart, []byte{0x24, 0x41}) // move.l d1,a2

	// test 0
	c.d[1] = 0x0
	c.a[2] = 0xffffffff
	err := c.Step()
	if err != nil {
		t.Fatal(err)
	}
	if c.pc != 0x1002 {
		t.Fatalf("pc 0x%x != 0x1002", c.pc)
	}

	if c.a[2] != c.d[1] {
		t.Fatalf("move.l 0x%x != 0x3000", c.a[2])
	}
	if c.sr&0x1f != 0x4 {
		t.Fatalf("sr 0x%x != 0x4", c.sr)
	}
	d, _, err := c.disassemble(pcStart)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v -> 0x%08x sr %02x -> %v", d, c.a[2], c.sr, c.ccr())

	// not 0 negative
	c.d[1] = 0xffffffff
	c.a[2] = 0x1
	c.pc = pcStart
	c.sr = 0x0
	err = c.Step()
	if err != nil {
		t.Fatal(err)
	}
	if c.pc != 0x1002 {
		t.Fatalf("pc 0x%x != 0x1002", c.pc)
	}

	if c.a[2] != c.d[1] {
		t.Fatalf("move.l 0x%x != 0x3000", c.a[2])
	}
	if c.sr&0x1f != 0x8 {
		t.Fatalf("sr 0x%x != 0x0", c.sr)
	}
	d, _, err = c.disassemble(pcStart)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v -> 0x%08x sr %02x -> %v", d, c.a[2], c.sr, c.ccr())

	// not 0 not negative
	c.d[1] = 0x1000
	c.a[2] = 0x1
	c.pc = pcStart
	c.sr = 0x0
	err = c.Step()
	if err != nil {
		t.Fatal(err)
	}
	if c.pc != 0x1002 {
		t.Fatalf("pc 0x%x != 0x1002", c.pc)
	}

	if c.a[2] != c.d[1] {
		t.Fatalf("move.l 0x%x != 0x3000", c.a[2])
	}
	if c.sr&0x1f != 0x0 {
		t.Fatalf("sr 0x%x != 0x0", c.sr)
	}
	d, _, err = c.disassemble(pcStart)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v -> 0x%08x sr %02x -> %v", d, c.a[2], c.sr, c.ccr())

	// indirect
	b.Write(pcStart, []byte{0x24, 0x81}) // move.l d1,(a2)
	c.d[1] = 0xaaaa5555
	c.a[2] = 0x4000
	c.pc = pcStart
	c.sr = 0x0
	err = c.Step()
	if err != nil {
		t.Fatal(err)
	}
	if c.pc != 0x1002 {
		t.Fatalf("pc 0x%x != 0x1002", c.pc)
	}

	v := c.read32(0x4000)
	if v != c.d[1] {
		t.Fatalf("move.l 0x%x != 0xaaaa5555", v)
	}
	if c.sr&0x1f != 0x08 {
		t.Fatalf("sr 0x%x != 0x0", c.sr)
	}
	d, _, err = c.disassemble(pcStart)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v -> 0x%08x sr %02x -> %v", d, c.a[2], c.sr, c.ccr())
}

func TestADDD(t *testing.T) {
	// adda.l d1,a2
	b, c := newCpu()
	b.Write(pcStart, []byte{0xd5, 0xc1}) // adda.l d1,a2
	d, _, err := c.disassemble(pcStart)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", d)

	// adda.w d1,a2
	d, _, err = dAdd(c, 0xd4c1, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", d)

	// add.l d1,d2
	d, _, err = dAdd(c, 0xd481, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", d)

	// add.l d1,(a2)
	d, _, err = dAdd(c, 0xd392, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", d)

	// add.w d1,(a2)+
	d, _, err = dAdd(c, 0xd35a, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", d)

	// add.b d6,-(a5)
	d, _, err = dAdd(c, 0xdd25, []byte{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", d)
}

func TestADDAL(t *testing.T) {
	b, c := newCpu()
	b.Write(pcStart, []byte{0xd5, 0xc1}) // adda.l d1,a2

	// test 0
	c.d[1] = 0xffffffff
	c.a[2] = 0x1
	err := c.Step()
	if err != nil {
		t.Fatal(err)
	}
	if c.pc != 0x1002 {
		t.Fatalf("pc 0x%x != 0x1002", c.pc)
	}

	if c.a[2] != 0x0 {
		t.Fatalf("adda.l 0x%x != 0x0", c.a[2])
	}
	if c.sr&0x1f != 0x15 {
		t.Fatalf("sr 0x%x != 0x15", c.sr)
	}
	d, _, err := c.disassemble(pcStart)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v -> 0x%08x sr %02x -> %v", d, c.a[2], c.sr, c.ccr())

	// test overflow
	c.d[1] = 0x7fffffff
	c.a[2] = 0x1
	c.pc = pcStart
	c.sr = 0x0
	err = c.Step()
	if err != nil {
		t.Fatal(err)
	}
	if c.pc != 0x1002 {
		t.Fatalf("pc 0x%x != 0x1002", c.pc)
	}

	if c.a[2] != 0x80000000 {
		t.Fatalf("adda.l 0x%x != 0x80000000", c.a[2])
	}
	if c.sr&0x1f != 0x2 {
		t.Fatalf("sr 0x%x != 0x2", c.sr)
	}
	t.Logf("0x%08x sr %02x -> %v", c.a[2], c.sr, c.ccr())

	// test simple
	c.d[1] = 0x1
	c.a[2] = 0x1
	c.pc = pcStart
	c.sr = 0x0
	err = c.Step()
	if err != nil {
		t.Fatal(err)
	}
	if c.pc != 0x1002 {
		t.Fatalf("pc 0x%x != 0x1002", c.pc)
	}

	if c.a[2] != 0x2 {
		t.Fatalf("adda.l 0x%x != 0x2", c.a[2])
	}
	if c.sr&0x1f != 0x0 {
		t.Fatalf("sr 0x%x != 0x0", c.sr)
	}
	t.Logf("0x%08x sr %02x -> %v", c.a[2], c.sr, c.ccr())
}
