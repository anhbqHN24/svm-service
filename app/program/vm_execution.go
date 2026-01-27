package program

import (
	"encoding/binary"
	"fmt"
	"svm_whiteboard/app/model"
)

type VMError struct {
	PC      int
	Message string
}

func (e *VMError) Error() string {
	return fmt.Sprintf("Runtime Error at PC %d: %s", e.PC, e.Message)
}

type VM struct {
	Registers [4]int  // Scratchpad registers (R0, R1, R2, R3)
	RegTypes  [4]byte // Type tracking for registers

	Stack [1024]int // Stack Memory: place to store local variables

	Memory  [1024]byte // Heap Memory: for dynamic string allocation
	HeapPtr int
	Flag    int
	Program []byte
	PC      int
	Output  []string
}

func (vm *VM) GetStackValue(i int) any {
	val := vm.Stack[i]
	// Heuristic to determine if val is string pointer or int
	if val >= 0 && val < vm.HeapPtr {
		// Try read as string
		str, err := vm.ReadString(val)
		if err == nil {
			return str
		}
	}
	return val
}

func NewVM(code []byte) *VM {
	return &VM{
		Program: code,
		PC:      0,
		HeapPtr: 0,
		Output:  []string{},
		// Stack & Registers auto-initialized to 0
	}
}

// Get Register 0 (Return value)
func (vm *VM) GetRegister1() int {
	return vm.Registers[0]
}

func (vm *VM) MallocString(content string) (int, error) {
	bytes := []byte(content)
	if vm.HeapPtr+len(bytes)+1 > len(vm.Memory) {
		return -1, fmt.Errorf("Out of Memory (Heap Overflow)")
	}
	startAddr := vm.HeapPtr
	copy(vm.Memory[startAddr:], bytes)
	vm.Memory[startAddr+len(bytes)] = 0 // Null terminator
	vm.HeapPtr += len(bytes) + 1
	return startAddr, nil
}

func (vm *VM) ReadMem(addr int) (byte, error) {
	if addr < 0 || addr >= len(vm.Memory) {
		return 0, fmt.Errorf("Segfault %d", addr)
	}
	return vm.Memory[addr], nil
}

// ==========================================
// CORE EXECUTION LOOP
// ==========================================
func (vm *VM) Run() ([]string, error) {
	vm.Output = append(vm.Output, "--- VM STARTED ---")

	for vm.PC < len(vm.Program) {
		op := vm.Program[vm.PC]

		// 1. VARIABLE LENGTH INSTRUCTION (LOAD_STR)
		if op == model.OP_LOAD_STR {
			if vm.PC+2 >= len(vm.Program) {
				return vm.Output, &VMError{vm.PC, "Unexpected EOF"}
			}
			dest := vm.Program[vm.PC+1]
			strLen := int(vm.Program[vm.PC+2])

			if vm.PC+3+strLen > len(vm.Program) {
				return vm.Output, &VMError{vm.PC, "EOF reading string"}
			}
			content := string(vm.Program[vm.PC+3 : vm.PC+3+strLen])

			// Alloc & Store Ptr to Register
			ptr, err := vm.MallocString(content)
			if err != nil {
				return vm.Output, &VMError{vm.PC, err.Error()}
			}

			vm.Registers[dest] = ptr
			vm.RegTypes[dest] = model.TYPE_STR
			vm.Output = append(vm.Output, fmt.Sprintf(">> LOAD_STR R%d = \"%s\"", dest, content))

			// Debug log (Optional)
			// vm.Output = append(vm.Output, fmt.Sprintf(">> LOAD_STR R%d = \"%s\"", dest, content))

			vm.PC += 3 + strLen
			continue
		}

		// 2. STACK OPS (4 Bytes)
		if op == model.OP_LOAD || op == model.OP_STORE {
			if vm.PC+3 >= len(vm.Program) {
				return vm.Output, &VMError{vm.PC, "Unexpected EOF"}
			}
			reg := vm.Program[vm.PC+1]
			addrIdx := binary.BigEndian.Uint16(vm.Program[vm.PC+2 : vm.PC+4])

			if int(addrIdx) >= len(vm.Stack) {
				return vm.Output, &VMError{vm.PC, "Stack Overflow"}
			}

			if op == model.OP_LOAD {
				// Stack -> Register
				vm.Registers[reg] = vm.Stack[addrIdx]
				// Giả sử type từ Stack luôn valid (Simple model: coi như INT hoặc PTR đều là số)
				// Để chính xác hơn cần StackTypes[] nhưng ta tạm bỏ qua cho gọn
				vm.RegTypes[reg] = model.TYPE_INT
				// Nếu giá trị > 0 và nằm trong Heap range thì có thể là STR, nhưng ta cứ để INT
				// Runtime check lúc PRINT sẽ lo.
			} else {
				// Register -> Stack
				vm.Stack[addrIdx] = vm.Registers[reg]
			}
			vm.PC += 4
			continue
		}

		// 3. STANDARD OPS (3 Bytes)
		if vm.PC+2 >= len(vm.Program) {
			if op == model.OP_HALT {
				vm.Output = append(vm.Output, ">> HALT encountered")
				break
			}
			return vm.Output, &VMError{vm.PC, "Unexpected End"}
		}
		p1, p2 := int(vm.Program[vm.PC+1]), int(vm.Program[vm.PC+2])

		getReg := func(i int) int { return vm.Registers[i] }

		switch op {
		case model.OP_LOAD_IMM:
			vm.Registers[p1] = p2
			vm.RegTypes[p1] = model.TYPE_INT
			vm.Output = append(vm.Output, fmt.Sprintf(">> LOAD_IMM R%d = %d", p1, p2))

		case model.OP_MOV:
			vm.Registers[p1] = getReg(p2)
			vm.RegTypes[p1] = vm.RegTypes[p2]
			vm.Output = append(vm.Output, fmt.Sprintf(">> MOV R%d = R%d (%d)", p1, p2, getReg(p2)))

		// --- ARITHMETIC ---

		case model.OP_ADD:
			vm.Registers[p1] += getReg(p2)
			vm.RegTypes[p1] = model.TYPE_INT
			vm.Output = append(vm.Output, fmt.Sprintf(">> ADD R%d += R%d (%d)", p1, p2, getReg(p2)))
		case model.OP_SUB:
			vm.Registers[p1] -= getReg(p2)
			vm.RegTypes[p1] = model.TYPE_INT
			vm.Output = append(vm.Output, fmt.Sprintf(">> SUB R%d -= R%d (%d)", p1, p2, getReg(p2)))
		case model.OP_MUL:
			vm.Registers[p1] *= getReg(p2)
			vm.RegTypes[p1] = model.TYPE_INT
			vm.Output = append(vm.Output, fmt.Sprintf(">> MUL R%d *= R%d (%d)", p1, p2, getReg(p2)))
		case model.OP_DIV:
			if v := getReg(p2); v == 0 {
				return vm.Output, &VMError{vm.PC, "Div By Zero"}
			} else {
				vm.Registers[p1] /= v
				vm.RegTypes[p1] = model.TYPE_INT
				vm.Output = append(vm.Output, fmt.Sprintf(">> DIV R%d /= R%d (%d)", p1, p2, v))
			}
		case model.OP_MOD:
			if v := getReg(p2); v == 0 {
				return vm.Output, &VMError{vm.PC, "Mod By Zero"}
			} else {
				vm.Registers[p1] %= v
				vm.RegTypes[p1] = model.TYPE_INT
				vm.Output = append(vm.Output, fmt.Sprintf(">> MOD R%d %%= R%d (%d)", p1, p2, v))
			}

		// --- COMPARISON & JUMP ---

		case model.OP_CMP:
			v1, v2 := getReg(p1), getReg(p2)
			if v1 == v2 {
				vm.Flag = 0
			} else if v1 > v2 {
				vm.Flag = 1
			} else {
				vm.Flag = -1
			}
			vm.Output = append(vm.Output, fmt.Sprintf(">> CMP R%d (%d) vs R%d (%d) => Flag=%d", p1, v1, p2, v2, vm.Flag))

		case model.OP_JMP:
			vm.PC = p2
			vm.Output = append(vm.Output, fmt.Sprintf(">> JMP to %d", p2))
			continue
		case model.OP_JEQ:
			if vm.Flag == 0 {
				vm.PC = p2
				vm.Output = append(vm.Output, fmt.Sprintf(">> JEQ to %d", p2))
				continue
			}
		case model.OP_JNE:
			if vm.Flag != 0 {
				vm.PC = p2
				vm.Output = append(vm.Output, fmt.Sprintf(">> JNE to %d", p2))
				continue
			}
		case model.OP_JGT:
			if vm.Flag == 1 {
				vm.PC = p2
				vm.Output = append(vm.Output, fmt.Sprintf(">> JGT to %d", p2))
				continue
			}
		case model.OP_JLT:
			if vm.Flag == -1 {
				vm.PC = p2
				vm.Output = append(vm.Output, fmt.Sprintf(">> JLT to %d", p2))
				continue
			}

		case model.OP_CONCAT:
			str1, _ := vm.GetStringVal(p1)
			str2, _ := vm.GetStringVal(p2)
			newPtr, err := vm.MallocString(str1 + str2)
			if err != nil {
				return vm.Output, &VMError{vm.PC, err.Error()}
			}
			vm.Registers[p1] = newPtr
			vm.RegTypes[p1] = model.TYPE_STR
			vm.Output = append(vm.Output, fmt.Sprintf(">> CONCAT R%d = \"%s\" + \"%s\"", p1, str1, str2))

		// --- IO & SYSTEM ---

		case model.OP_PRINT_INT:
			// Smart Print: Check type thật sự
			// Note: Sau khi LOAD từ Stack, type có thể bị mất (default INT).
			// Nếu muốn smart print hoàn hảo, cần lưu type vào StackTypes.
			// Nhưng ở đây ta check tạm: Nếu opcode là PRINT_INT nhưng user muốn in chuỗi?
			// Ta cho phép user dùng PRINT (Compiler map về OP_PRINT_INT).
			val := vm.Registers[p1]
			// Heuristic: Nếu val là địa chỉ hợp lệ trong heap và trỏ đến string -> In string?
			// Nhưng rủi ro. Ta cứ in INT. Nếu muốn in string dùng hàm GetStringVal logic.
			// Ở đây ta dùng logic đơn giản:
			vm.Output = append(vm.Output, fmt.Sprintf(">> OUT: %d", val))

		case model.OP_PRINT_STR:
			// Đây là opcode dành riêng in string
			str, err := vm.GetStringVal(p1)
			if err != nil {
				return vm.Output, &VMError{vm.PC, err.Error()}
			}
			vm.Output = append(vm.Output, fmt.Sprintf(">> OUT: %s", str))

		case model.OP_HALT:
			vm.Output = append(vm.Output, ">> HALT encountered")
			return vm.Output, nil
		}
		vm.PC += 3
	}
	return vm.Output, nil
}

// ==========================================
// INPUT PARSING
// ==========================================

func LoadStrictParams(vm *VM, p1 interface{}, p2 interface{}) error {
	inputs := []interface{}{p1, p2}
	// [IMPORTANT] Params được load vào STACK[0] và STACK[1]
	for i, val := range inputs {
		if val == nil {
			vm.Stack[i] = 0
			continue
		}
		switch v := val.(type) {
		case float64:
			vm.Stack[i] = int(v)
		case int:
			vm.Stack[i] = v
		case string:
			addr, err := vm.MallocString(v)
			if err != nil {
				return err
			}
			vm.Stack[i] = addr
		default:
			return fmt.Errorf("param_%d invalid type", i+1)
		}
	}
	return nil
}

func (vm *VM) ReadString(addr int) (string, error) {
	if addr < 0 || addr >= len(vm.Memory) {
		return "", fmt.Errorf("Invalid Mem Access")
	}
	var strBuf []byte
	curr := addr
	for {
		if curr >= len(vm.Memory) {
			break
		}
		b := vm.Memory[curr]
		if b == 0 {
			break
		}
		strBuf = append(strBuf, b)
		curr++
		if len(strBuf) > 256 {
			break
		} // Safety
	}
	return string(strBuf), nil
}

func (vm *VM) GetStringVal(regIdx int) (string, error) {
	val := vm.Registers[regIdx]
	// Nếu gọi hàm này, ta assume user muốn đọc nó như string
	return vm.ReadString(val)
}
