package model

import "fmt"

// VM Instructions (Opcodes)
const (
	// --- Data Movement ---
	OP_LOAD_IMM = 0x01 // [Reg] [Value] -> Load value into Reg
	OP_MOV      = 0x02 // [Dest] [Src]  -> Copy Reg value to Reg

	// --- Arithmetic ---
	OP_ADD = 0x10 // [Dest] [Src]
	OP_SUB = 0x11 // [Dest] [Src]
	OP_MUL = 0x12 // [Dest] [Src]
	OP_DIV = 0x13 // [Dest] [Src] -> Checks for zero division
	OP_MOD = 0x14 // [Dest] [Src] -> Checks for zero division

	// --- Logic ---
	OP_AND = 0x20 // [Dest] [Src]
	OP_OR  = 0x21 // [Dest] [Src]
	OP_XOR = 0x22 // [Dest] [Src]

	// --- Control Flow ---
	OP_CMP = 0x30 // [RegA] [RegB] -> Set Flag (-1, 0, 1)
	OP_JMP = 0x31 // [0] [Addr]    -> Jump always
	OP_JEQ = 0x32 // [0] [Addr]    -> Jump if Flag == 0 (Equal)
	OP_JNE = 0x33 // [0] [Addr]    -> Jump if Flag != 0
	OP_JGT = 0x34 // [0] [Addr]    -> Jump if Flag == 1 (Greater)
	OP_JLT = 0x35 // [0] [Addr]    -> Jump if Flag == -1 (Less)

	// --- String Operations ---
	OP_CONCAT = 0x40 // [DestReg] [SrcReg] -> Concatenate strings from SrcReg to DestReg

	// --- IO & System ---
	OP_PRINT_INT = 0xEE // [Reg] [0] -> Print integer
	OP_PRINT_STR = 0xF0 // [Reg] [0] -> Print string from memory pointer
	OP_HALT      = 0xFF // Stop execution
)

var OpCostTable = map[byte]int{
	// Data (Cheap)
	OP_LOAD_IMM: 2,
	OP_MOV:      2,

	// Arithmetic (Medium)
	OP_ADD: 3,
	OP_SUB: 3,
	OP_MUL: 5,
	OP_DIV: 5,
	OP_MOD: 5,

	// Logic (Cheap)
	OP_AND: 3,
	OP_OR:  3,
	OP_XOR: 3,

	// Control Flow (Very cheap)
	OP_CMP: 2,
	OP_JMP: 1,
	OP_JEQ: 1,
	OP_JNE: 1,
	OP_JGT: 1,
	OP_JLT: 1,

	// String Operations (Expensive)
	OP_CONCAT: 15, // high cost due to alloc memory operations

	// IO & System (Expensive)
	OP_PRINT_INT: 10,
	OP_PRINT_STR: 20,
	OP_HALT:      0,
}

// --- TYPE SYSTEM ---
const (
	TYPE_INT = 0 // Register contain Integer value
	TYPE_STR = 1 // Register contain String value
)

const MaxComputeCycle = 100

// Module 2: Static Analysis - Calculate Cost before running
func EstimateComputeCost(program []byte) (int, error) {
	totalCost := 0
	// Scan the binary code
	for i := 0; i < len(program); {
		if i+3 > len(program) {
			break
		}
		op := program[i]

		if op == 0x00 {
			i += 3 // Bỏ qua block này (Padding)
			continue
		}

		// 1. Add cost from table
		if cost, ok := OpCostTable[op]; ok {
			totalCost += cost
		} else {
			// Return error if opcode is unknown
			return 0, fmt.Errorf("invalid opcode detected: 0x%X", op)
		}

		// 2. Skip to next instruction
		// NOTE: Changed to 3 because instruction format is [OP] [ARG1] [ARG2]
		i += 3
	}
	return totalCost, nil
}
