package program

import (
	"fmt"
	"svm_whiteboard/internal/core"
)

type VMError struct {
	PC      int
	Message string
}

func (e *VMError) Error() string {
	return fmt.Sprintf("Runtime Error at PC %d: %s", e.PC, e.Message)
}

// ==========================================
// 3. VM ARCHITECTURE
// ==========================================
type VM struct {
	Registers [4]int     // R0, R1, R2, R3
	Memory    [1024]byte // Heap Memory (For Strings)
	HeapPtr   int        // Con trỏ quản lý bộ nhớ
	Flag      int        // Trạng thái so sánh (-1: <, 0: =, 1: >)
	Program   []byte
	PC        int
	Output    []string
}

func NewVM(code []byte) *VM {
	return &VM{
		Program: code,
		PC:      0,
		HeapPtr: 0, // Bộ nhớ bắt đầu từ 0
		Output:  []string{},
	}
}

// Cấp phát bộ nhớ cho String và trả về địa chỉ (Pointer)
func (vm *VM) MallocString(content string) (int, error) {
	bytes := []byte(content)
	// Kiểm tra tràn bộ nhớ
	if vm.HeapPtr+len(bytes)+1 > len(vm.Memory) {
		return -1, fmt.Errorf("Out of Memory (Heap Overflow)")
	}

	startAddr := vm.HeapPtr
	copy(vm.Memory[startAddr:], bytes)

	// Thêm Null Terminator (0) để đánh dấu hết chuỗi
	vm.Memory[startAddr+len(bytes)] = 0

	// Cập nhật con trỏ Heap
	vm.HeapPtr += len(bytes) + 1

	return startAddr, nil
}

// Đọc an toàn từ Memory
func (vm *VM) ReadMem(addr int) (byte, error) {
	if addr < 0 || addr >= len(vm.Memory) {
		return 0, fmt.Errorf("Segmentation Fault (Access Addr %d)", addr)
	}
	return vm.Memory[addr], nil
}

// ==========================================
// 4. CORE EXECUTION LOOP
// ==========================================
func (vm *VM) Run() ([]string, error) {
	vm.Output = append(vm.Output, "--- VM STARTED ---")

	for vm.PC < len(vm.Program) {
		// 1. Safety Check: Instruction Bounds
		if vm.PC+2 >= len(vm.Program) {
			if vm.Program[vm.PC] == core.OP_HALT {
				break
			} // Halt lệnh cuối ok
			return vm.Output, &VMError{vm.PC, "Unexpected End of Program"}
		}

		op := vm.Program[vm.PC]
		p1 := int(vm.Program[vm.PC+1])
		p2 := int(vm.Program[vm.PC+2])

		// Helper lấy tên Register log cho đẹp
		regVal := func(idx int) int {
			if idx >= 0 && idx < 4 {
				return vm.Registers[idx]
			}
			return 0
		}

		switch op {
		// --- DATA ---
		case core.OP_LOAD_IMM:
			if p1 < 0 || p1 > 3 {
				return vm.Output, &VMError{vm.PC, "Invalid Register Index"}
			}
			vm.Registers[p1] = p2

		case core.OP_MOV:
			vm.Registers[p1] = vm.Registers[p2]

		// --- ARITHMETIC (Có check lỗi) ---
		case core.OP_ADD:
			vm.Registers[p1] += vm.Registers[p2]
		case core.OP_SUB:
			vm.Registers[p1] -= vm.Registers[p2]
		case core.OP_MUL:
			vm.Registers[p1] *= vm.Registers[p2]
		case core.OP_DIV:
			if regVal(p2) == 0 {
				return vm.Output, &VMError{vm.PC, "Division By Zero"}
			}
			vm.Registers[p1] /= vm.Registers[p2]
		case core.OP_MOD:
			if regVal(p2) == 0 {
				return vm.Output, &VMError{vm.PC, "Modulo By Zero"}
			}
			vm.Registers[p1] %= vm.Registers[p2]

		// --- LOGIC ---
		case core.OP_AND:
			vm.Registers[p1] &= vm.Registers[p2]
		case core.OP_OR:
			vm.Registers[p1] |= vm.Registers[p2]
		case core.OP_XOR:
			vm.Registers[p1] ^= vm.Registers[p2]

		// --- CONTROL FLOW ---
		case core.OP_CMP:
			v1, v2 := vm.Registers[p1], vm.Registers[p2]
			if v1 == v2 {
				vm.Flag = 0
			} else if v1 > v2 {
				vm.Flag = 1
			} else {
				vm.Flag = -1
			}

		case core.OP_JMP:
			vm.PC = p2
			continue
		case core.OP_JEQ:
			if vm.Flag == 0 {
				vm.PC = p2
				continue
			}
		case core.OP_JNE:
			if vm.Flag != 0 {
				vm.PC = p2
				continue
			}
		case core.OP_JGT:
			if vm.Flag == 1 {
				vm.PC = p2
				continue
			}
		case core.OP_JLT:
			if vm.Flag == -1 {
				vm.PC = p2
				continue
			}

		// --- IO (String & Int) ---
		case core.OP_PRINT_INT:
			vm.Output = append(vm.Output, fmt.Sprintf(">> INT: %d", vm.Registers[p1]))

		case core.OP_PRINT_STR:
			// Lấy địa chỉ từ Register -> Đọc Heap
			addr := vm.Registers[p1]
			var strBuf []byte
			for {
				b, err := vm.ReadMem(addr)
				if err != nil {
					return vm.Output, &VMError{vm.PC, err.Error()}
				}
				if b == 0 {
					break
				} // Null terminator
				strBuf = append(strBuf, b)
				addr++
				if len(strBuf) > 128 { // Safety break để tránh in quá dài
					break
				}
			}
			vm.Output = append(vm.Output, fmt.Sprintf(">> STRING: %s", string(strBuf)))

		case core.OP_HALT:
			return vm.Output, nil

		default:
			return vm.Output, &VMError{vm.PC, fmt.Sprintf("Illegal Opcode 0x%X", op)}
		}

		vm.PC += 3
	}
	return vm.Output, nil
}

// ==========================================
// 5. INPUT PARSING (API LAYER)
// ==========================================

func LoadStrictParams(vm *VM, p1 interface{}, p2 interface{}) error {
	inputs := []interface{}{p1, p2}

	// Loop qua 2 input để nạp vào R0 và R1
	for i, val := range inputs {
		if val == nil {
			// Nếu null thì mặc định là 0
			vm.Registers[i] = 0
			continue
		}

		switch v := val.(type) {
		case float64: // JSON Number
			vm.Registers[i] = int(v)
		case string: // JSON String -> Heap
			addr, err := vm.MallocString(v)
			if err != nil {
				return err
			}
			vm.Registers[i] = addr
		default:
			return fmt.Errorf("param_%d invalid type (must be string or int)", i+1)
		}
	}
	return nil
}
