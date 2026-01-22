package program

import (
	"fmt"
	"svm_whiteboard/app/model"
)

// VMError includes the Program Counter (PC) for debugging
type VMError struct {
	PC      int
	Message string
}

func (e *VMError) Error() string {
	return fmt.Sprintf("Runtime Error at PC %d: %s", e.PC, e.Message)
}

type VM struct {
	Registers [4]int     // R0, R1, R2, R3
	RegTypes  [4]byte    // Type of data in each register (int or string pointer)
	Memory    [1024]byte // Heap Memory (For Strings)
	HeapPtr   int        // Pointer to next free memory address
	Flag      int        // Comparison flag (-1: <, 0: =, 1: >)
	Program   []byte
	PC        int
	Output    []string
}

func NewVM(code []byte) *VM {
	return &VM{
		Program: code,
		PC:      0,
		HeapPtr: 0,
		Output:  []string{},
	}
}

// Helper: Get value of R1 (usually the return value)
func (vm *VM) GetRegister1() int {
	return vm.Registers[0]
}

// Helper: Allocate string to heap and return its address
// Used to pre-load data into memory before Run()
func (vm *VM) MallocString(content string) (int, error) {
	bytes := []byte(content)
	// Check for heap overflow
	if vm.HeapPtr+len(bytes)+1 > len(vm.Memory) {
		return -1, fmt.Errorf("Out of Memory (Heap Overflow)")
	}

	startAddr := vm.HeapPtr
	copy(vm.Memory[startAddr:], bytes)

	// Add Null Terminator (0)
	vm.Memory[startAddr+len(bytes)] = 0

	// Update Heap Pointer
	vm.HeapPtr += len(bytes) + 1

	return startAddr, nil
}

// Helper: Safely read a byte from Memory
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
		// Safety Check: Instruction Bounds
		// Each instruction is 3 bytes: [OP] [ARG1] [ARG2]
		if vm.PC+2 >= len(vm.Program) {
			// Allow HALT if it is the last byte
			if vm.Program[vm.PC] == model.OP_HALT {
				break
			}
			return vm.Output, &VMError{vm.PC, "Unexpected End of Program"}
		}

		op := vm.Program[vm.PC]
		p1 := int(vm.Program[vm.PC+1]) // Register Index or Value
		p2 := int(vm.Program[vm.PC+2]) // Register Index or Value

		// Helper to safely access registers
		regVal := func(idx int) int {
			if idx >= 0 && idx < 4 {
				return vm.Registers[idx]
			}
			return 0
		}

		switch op {
		// --- DATA ---
		case model.OP_LOAD_IMM:
			if p1 < 0 || p1 > 3 {
				return vm.Output, &VMError{vm.PC, "Invalid Register Index"}
			}
			vm.Registers[p1] = p2
			vm.RegTypes[p1] = model.TYPE_INT // [UPDATE] Load số -> Type là INT

		case model.OP_MOV:
			if p1 < 0 || p1 > 3 {
				return vm.Output, &VMError{vm.PC, "Invalid Dest Register"}
			}
			vm.Registers[p1] = regVal(p2)
			vm.RegTypes[p1] = vm.RegTypes[p2]

		// --- ARITHMETIC (With error checks) ---
		case model.OP_ADD:
			vm.Registers[p1] += regVal(p2)
			vm.RegTypes[p1] = model.TYPE_INT

		case model.OP_SUB:
			vm.Registers[p1] -= regVal(p2)
			vm.RegTypes[p1] = model.TYPE_INT

		case model.OP_MUL:
			vm.Registers[p1] *= regVal(p2)
			vm.RegTypes[p1] = model.TYPE_INT

		case model.OP_DIV:
			if regVal(p2) == 0 {
				return vm.Output, &VMError{vm.PC, "Division By Zero"}
			}
			vm.Registers[p1] /= regVal(p2)
			vm.RegTypes[p1] = model.TYPE_INT

		case model.OP_MOD:
			if regVal(p2) == 0 {
				return vm.Output, &VMError{vm.PC, "Modulo By Zero"}
			}
			vm.Registers[p1] %= regVal(p2)

		// --- LOGIC ---
		case model.OP_AND:
			vm.Registers[p1] &= regVal(p2)
		case model.OP_OR:
			vm.Registers[p1] |= regVal(p2)
		case model.OP_XOR:
			vm.Registers[p1] ^= regVal(p2)

		// --- CONTROL FLOW ---
		case model.OP_CMP:
			v1, v2 := regVal(p1), regVal(p2)
			if v1 == v2 {
				vm.Flag = 0
			} else if v1 > v2 {
				vm.Flag = 1
			} else {
				vm.Flag = -1
			}

		case model.OP_JMP:
			vm.PC = p2
			continue // Skip PC increment
		case model.OP_JEQ:
			if vm.Flag == 0 {
				vm.PC = p2
				continue
			}
		case model.OP_JNE:
			if vm.Flag != 0 {
				vm.PC = p2
				continue
			}
		case model.OP_JGT:
			if vm.Flag == 1 {
				vm.PC = p2
				continue
			}
		case model.OP_JLT:
			if vm.Flag == -1 {
				vm.PC = p2
				continue
			}

		// --- STRING OPERATIONS ---
		case model.OP_CONCAT:
			str1, err := vm.GetStringVal(p1)
			if err != nil {
				return vm.Output, &VMError{vm.PC, err.Error()}
			}

			// 2. Lấy String từ P2 (Tự động nhận biết Int/Ptr)
			str2, err := vm.GetStringVal(p2)
			if err != nil {
				return vm.Output, &VMError{vm.PC, err.Error()}
			}

			// 3. Ghép chuỗi
			newStr := str1 + str2

			// 4. Cấp phát vùng nhớ mới
			newPtr, err := vm.MallocString(newStr)
			if err != nil {
				return vm.Output, &VMError{vm.PC, err.Error()}
			}

			// 5. Lưu kết quả vào P1
			vm.Registers[p1] = newPtr
			vm.RegTypes[p1] = model.TYPE_STR

		// --- IO (String & Int) ---
		case model.OP_PRINT_INT:
			vm.Output = append(vm.Output, fmt.Sprintf(">> INT: %d", regVal(p1)))

		case model.OP_PRINT_STR:
			if vm.RegTypes[p1] != model.TYPE_STR {
				return vm.Output, &VMError{vm.PC, "Invalid String Pointer"}
			}
			// Get address from p1 -> Read from Heap
			addr := regVal(p1)
			var strBuf []byte
			currAddr := addr

			// Read memory until Null Terminator (0)
			for {
				b, err := vm.ReadMem(currAddr)
				if err != nil {
					return vm.Output, &VMError{vm.PC, err.Error()}
				}
				if b == 0 {
					break // Null terminator
				}
				strBuf = append(strBuf, b)
				currAddr++

				// Safety break to prevent long output or loops
				if len(strBuf) > 128 {
					strBuf = append(strBuf, []byte("...[TRUNCATED]")...)
					break
				}
			}
			vm.Output = append(vm.Output, fmt.Sprintf(">> STRING: %s", string(strBuf)))

		case model.OP_HALT:
			return vm.Output, nil

		// Skip NOOP or Padding
		case 0x00:
			// Do nothing

		default:
			return vm.Output, &VMError{vm.PC, fmt.Sprintf("Illegal Opcode 0x%X", op)}
		}

		// Move to next instruction (3 bytes)
		vm.PC += 3
	}
	return vm.Output, nil
}

// ==========================================
// 5. INPUT PARSING (API LAYER)
// ==========================================

func LoadStrictParams(vm *VM, p1 interface{}, p2 interface{}) error {
	inputs := []interface{}{p1, p2}

	for i, val := range inputs {
		if val == nil {
			vm.Registers[i] = 0
			vm.RegTypes[i] = model.TYPE_INT // Default
			continue
		}

		switch v := val.(type) {
		case float64: // JSON Number
			vm.Registers[i] = int(v)
			vm.RegTypes[i] = model.TYPE_INT // Flag as INT
		case string: // JSON String
			addr, err := vm.MallocString(v)
			if err != nil {
				return err
			}
			vm.Registers[i] = addr
			vm.RegTypes[i] = model.TYPE_STR // Flag as STRING
		case int:
			vm.Registers[i] = v
			vm.RegTypes[i] = model.TYPE_INT // Flag as INT
		default:
			return fmt.Errorf("param_%d invalid type", i+1)
		}
	}
	return nil
}

// Read string from memory heap given starting address
func (vm *VM) ReadString(addr int) (string, error) {
	if addr < 0 || addr >= len(vm.Memory) {
		return "", fmt.Errorf("segfault: invalid memory access at %d", addr)
	}
	var strBuf []byte
	currAddr := addr
	for {
		if currAddr >= len(vm.Memory) {
			break
		}
		b := vm.Memory[currAddr]
		if b == 0 {
			break
		} // Null terminator
		strBuf = append(strBuf, b)
		currAddr++
		if len(strBuf) > 128 {
			break
		} // Limit độ dài
	}
	return string(strBuf), nil
}

// Helper: Lấy giá trị chuỗi của Register bất kỳ (Tự parse Int/String)
func (vm *VM) GetStringVal(regIdx int) (string, error) {
	if regIdx < 0 || regIdx > 3 {
		return "", fmt.Errorf("invalid register index")
	}

	val := vm.Registers[regIdx]
	typeVal := vm.RegTypes[regIdx]

	if typeVal == model.TYPE_STR {
		// Nếu là Pointer -> Đọc từ Heap
		return vm.ReadString(val)
	}
	// Nếu là Int -> Convert sang string
	return fmt.Sprintf("%d", val), nil
}
