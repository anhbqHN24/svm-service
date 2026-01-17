package core

// These are the "Syscalls" your VM understands.
const (
	// --- Data Movement ---
	OP_LOAD_IMM = 0x01 // [Reg] [Value] -> Nạp số vào Reg
	OP_MOV      = 0x02 // [Dest] [Src]  -> Copy giá trị Reg -> Reg

	// --- Arithmetic ---
	OP_ADD = 0x10 // [Dest] [Src]
	OP_SUB = 0x11 // [Dest] [Src]
	OP_MUL = 0x12 // [Dest] [Src]
	OP_DIV = 0x13 // [Dest] [Src] -> Checks zero division
	OP_MOD = 0x14 // [Dest] [Src] -> Checks zero division

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

	// --- IO & System ---
	OP_PRINT_INT = 0xEE // [Reg] [0] -> In số
	OP_PRINT_STR = 0xF0 // [Reg] [0] -> In chuỗi từ địa chỉ (Pointer) trong Reg
	OP_HALT      = 0xFF // Dừng
)
