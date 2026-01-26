package model

import "fmt"

// VM Instructions (Opcodes)
const (
	// --- Data Movement ---
	OP_LOAD_IMM = 0x01 // [Reg] [Value] -> Load value into Reg
	OP_MOV      = 0x02 // [Dest] [Src]  -> Copy Reg value to Reg
	OP_LOAD_STR = 0x05 // [Reg] [StrID] -> Load string from Heap to Reg
	OP_LOAD     = 0x06 // [Reg] [Addr] -> Load from Stack to Reg
	OP_STORE    = 0x07 // [Reg] [Addr] -> Store from Reg to Stack

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
	OP_LOAD:     5, // Memory access cost
	OP_STORE:    5, // Memory access cost
	OP_LOAD_STR: 5, // String operation cost

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

const MaxComputeCycle = 2000

// Module 2: Static Analysis - Calculate Cost before running
func EstimateComputeCost(program []byte) (int, error) {
	totalCost := 0
	pc := 0
	lenProg := len(program)
	// Scan the binary code
	for pc < lenProg {
		op := program[pc]
		if op == 0x00 {
			pc++
			continue
		} // Skip padding

		if cost, ok := OpCostTable[op]; ok {
			totalCost += cost
		} else {
			return 0, fmt.Errorf("invalid opcode detected: 0x%X at PC %d", op, pc)
		}

		// Move PC according to instruction format
		switch op {
		case OP_LOAD_STR:
			// [OP] [Dest] [Len] [Bytes...]
			if pc+3 > lenProg {
				return 0, fmt.Errorf("EOF at PC %d", pc)
			}
			strLen := int(program[pc+2])
			pc += 3 + strLen

		case OP_LOAD, OP_STORE:
			// [OP] [Reg] [AddrHi] [AddrLo] -> 4 bytes
			pc += 4

		case OP_HALT:
			return totalCost, nil

		default:
			// Standard 3 bytes: [OP] [ARG1] [ARG2]
			pc += 3
		}
	}
	return totalCost, nil
}
