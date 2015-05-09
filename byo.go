package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/marcopeereboom/byo/bus"
	"github.com/marcopeereboom/byo/cpu"
	"github.com/marcopeereboom/byo/cpu/m68000"
	"github.com/marcopeereboom/byo/memory"
)

func singleCPU(bus *bus.Bus, cpu cpu.CPUer) error {
	bus.Reset(true)
	cpu.Reset()
	err := cpu.Step()
	if err != nil {
		return err
	}
	return nil
}

func parseCPU(name string, bus *bus.Bus) (cpu.CPUer, error) {
	switch name {
	case m68000.M68000:
		ssp := make([]byte, 4)
		pc := make([]byte, 4)
		binary.BigEndian.PutUint32(ssp, 0x1000)
		binary.BigEndian.PutUint32(pc, 0x2000)
		bus.Write(0x0, ssp)
		bus.Write(0x4, pc)
		// move.l d1,a2
		//bus.Write(0x2000, []byte{0x24, 0x41})
		// adda.l d1,a2
		bus.Write(0x2000, []byte{0xd5, 0xc1})
		return m68000.New(bus)
	}

	return nil, fmt.Errorf("invalid CPU type: %v", name)
}

func parseRAM(ramRegions string, bus *bus.Bus) error {
	regions := strings.Split(ramRegions, ",")
	for _, region := range regions {
		// split size@address
		s := strings.Split(region, "@")
		if len(s) != 2 {
			return fmt.Errorf("invalid RAM region: %v", region)
		}
		size, err := strconv.ParseUint(s[0], 0, 64)
		if err != nil {
			return fmt.Errorf("invalid size: %v", region)
		}
		address, err := strconv.ParseUint(s[1], 0, 64)
		if err != nil {
			return fmt.Errorf("invalid address: %v", region)
		}
		ram := memory.NewRAM(size)
		_, err = bus.Attach(address, ram)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// flags
	cpuType := flag.String("cpu", "68000", "CPU type")
	ramRegions := flag.String("ram", "0x8000@0x0000",
		"RAM <size@address>[,size@address]")
	flag.Parse()

	var cpu cpu.CPUer
	bus, err := bus.New()
	if err != nil {
		goto done
	}
	err = parseRAM(*ramRegions, bus)
	if err != nil {
		goto done
	}
	cpu, err = parseCPU(*cpuType, bus)
	if err != nil {
		goto done
	}

	err = singleCPU(bus, cpu)
done:
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
