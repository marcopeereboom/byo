package m68000

var (
	opcodes = map[uint16]instruction{
		0x2441: {
			// move.l d1,a2
			disassemble:      dMove,
			fetchOperand:     fetchOperandNop,
			fetchSource:      fetchDn,
			source:           1,
			fetchDestination: fetchNop,
			storeDestination: storeAn,
			destination:      2,
			execute:          movel,
		},
		0x2481: {
			// move.l d1,(a2)
			disassemble:      dMove,
			fetchOperand:     fetchOperandNop,
			fetchSource:      fetchDn,
			source:           1,
			fetchDestination: fetchNop,
			storeDestination: storeAnIndirect,
			destination:      2,
			execute:          movel,
		},
		0xd5c1: {
			// adda.l d1,a2
			disassemble:      dAdd,
			fetchOperand:     fetchOperandNop,
			fetchSource:      fetchDn,
			source:           1,
			fetchDestination: fetchAn,
			storeDestination: storeAn,
			destination:      2,
			execute:          addal,
		},
	}
)
