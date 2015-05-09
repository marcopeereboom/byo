package memory

import (
	"encoding/binary"
	"testing"
)

func TestRAM(t *testing.T) {
	r := NewRAM(8 * 1024)
	r.Write(0x400, []byte{0x55, 0xaa})
	x := r.Read(0x400, 2)
	if binary.LittleEndian.Uint16(x) != 0xaa55 {
		t.Fatalf("invalid uint16 %x", x)
	}
}

func TestROM(t *testing.T) {
	r, err := NewROM(8*1024, "test/rom.bin")
	if err != nil {
		t.Fatal(err)
	}
	//r.Write(0x400, []byte{0x55, 0xaa})
	x := r.Read(0x0000, 2)
	if binary.LittleEndian.Uint16(x) != 0x504d {
		t.Fatalf("invalid uint16 %x", x)
	}
}
