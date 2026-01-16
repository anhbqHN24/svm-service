package core

// These are the "Syscalls" your VM understands.
const (
	OP_NOOP   = 0x00
	OP_LOAD_1 = 0x01 // Load Operand into R1
	OP_LOAD_2 = 0x02 // Load Operand into R2
	OP_ADD    = 0x03 // R1 = R1 + R2
	OP_SUB    = 0x04 // R1 = R1 - R2
	OP_PRINT  = 0x06 // Print R1

	// Control Flow (The "Turing Complete" part)
	OP_CMP = 0x07 // Flag = R1 - R2
	OP_JEQ = 0x08 // Jump to Address (Operand) if Flag == 0
	OP_JGT = 0x09 // Jump to Address (Operand) if Flag > 0
	OP_JMP = 0x0A // Unconditional Jump to Address (Operand)

	OP_HALT = 0xFF // Stop execution
)
