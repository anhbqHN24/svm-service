package core

import "fmt"

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

var OpCostTable = map[byte]int{
	// Data (Rẻ nhất)
	OP_LOAD_IMM: 2,
	OP_MOV:      2,

	// Arithmetic (Trung bình)
	OP_ADD: 3,
	OP_SUB: 3,
	OP_MUL: 5, // Nhân thường tốn hơn cộng
	OP_DIV: 5, // Chia tốn hơn
	OP_MOD: 5,

	// Logic (Rẻ)
	OP_AND: 3,
	OP_OR:  3,
	OP_XOR: 3,

	// Control Flow (Rất rẻ để khuyến khích logic)
	OP_CMP: 2,
	OP_JMP: 1,
	OP_JEQ: 1,
	OP_JNE: 1,
	OP_JGT: 1,
	OP_JLT: 1,

	// IO & System (Đắt nhất - Heavy Operations)
	OP_PRINT_INT: 10, // Syscall IO
	OP_PRINT_STR: 20, // Syscall IO + Memory Scan loop
	OP_HALT:      0,
}

const MaxComputeCycle = 100

// Module 2: Static Analysis - Tính toán Cost trước khi chạy
func EstimateComputeCost(program []byte) (int, error) {
	totalCost := 0
	// Quét qua toàn bộ binary code
	for i := 0; i < len(program); {
		op := program[i]

		// 1. Cộng cost dựa trên bảng giá
		if cost, ok := OpCostTable[op]; ok {
			totalCost += cost
		} else {
			// Gặp opcode lạ -> Báo lỗi ngay, đỡ tốn công chạy VM
			return 0, fmt.Errorf("invalid opcode detected: 0x%X", op)
		}

		// 2. Nhảy index (vì kiến trúc ta là [OP] [ARG] - 2 bytes)
		// Lưu ý: Static analysis đơn giản này giả định code chạy tuần tự.
		// Với code có loop phức tạp, cost thực tế sẽ cao hơn.
		// Nhưng đây là bước lọc "Pre-flight" rất tốt.
		i += 2
	}
	return totalCost, nil
}
