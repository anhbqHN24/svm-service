package compiler

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"svm_whiteboard/app/model"
)

type CompilerState struct {
	SymbolTable map[string]uint16
	NextStack   uint16
}

func NewCompilerState() *CompilerState {
	return &CompilerState{
		SymbolTable: map[string]uint16{
			"param_1": 0, // Reserved: always at Stack[0]
			"param_2": 1, // Reserved: always at Stack[1]
		},
		NextStack: 2, // User variables start from Stack[2] onwards
	}
}

// Helper: Get stack address (or allocate new) for a variable
func (cs *CompilerState) GetOrAllocVar(name string) (uint16, error) {
	if addr, ok := cs.SymbolTable[name]; ok {
		return addr, nil
	}
	if cs.NextStack >= 1024 {
		return 0, fmt.Errorf("stack overflow: too many variables")
	}
	addr := cs.NextStack
	cs.SymbolTable[name] = addr
	cs.NextStack++
	return addr, nil
}

// Helper: generate bytecode for memory related command (LOAD/STORE)
func emitMemOp(op byte, reg byte, addr uint16) []byte {
	buf := make([]byte, 4)
	buf[0] = op
	buf[1] = reg
	binary.BigEndian.PutUint16(buf[2:], addr)
	return buf
}

// Helper: auto generate Load (IMM, STR, or MEM) based on input
func (cs *CompilerState) emitLoad(reg byte, rawSrc string) ([]byte, error) {
	var bytecode []byte

	// CASE 1: String Literal ("hello")
	if strings.HasPrefix(rawSrc, "\"") && strings.HasSuffix(rawSrc, "\"") {
		content := rawSrc[1 : len(rawSrc)-1] // Remove quotes
		strBytes := []byte(content)
		// Opcode: [LOAD_STR] [Reg] [Len] [Bytes...]
		bytecode = append(bytecode, model.OP_LOAD_STR, reg, byte(len(strBytes)))
		bytecode = append(bytecode, strBytes...)
		return bytecode, nil
	}

	// CASE 2: Number Literal (10)
	if val, err := strconv.Atoi(rawSrc); err == nil {
		if val > 255 || val < 0 {
			// Currently demo only supports byte (0-255).
			return nil, fmt.Errorf("value %d out of range (0-255)", val)
		}
		// Opcode: [LOAD_IMM] [Reg] [Value]
		bytecode = append(bytecode, model.OP_LOAD_IMM, reg, byte(val))
		return bytecode, nil
	}

	// CASE 3: Variable (x, param_1)
	srcAddr, ok := cs.SymbolTable[rawSrc]
	if !ok {
		return nil, fmt.Errorf("undefined variable '%s'", rawSrc)
	}
	// Opcode: [LOAD] [Reg] [AddrHi] [AddrLo]
	return emitMemOp(model.OP_LOAD, reg, srcAddr), nil
}

// Helper: split up the command (e.g: SET x "hello world")
func parseLine(line string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false

	for _, r := range line {
		if r == '"' {
			inQuotes = !inQuotes
			current.WriteRune(r)
		} else if r == ' ' || r == ',' { // support comma as well
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func Compile(source string) ([]byte, error) {
	var bytecode []byte
	state := NewCompilerState()

	// --- VM Registers ---
	const (
		R0 = 0 // Accumulator
		R1 = 1 // Operand
	)

	lines := strings.Split(source, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		parts := parseLine(line)
		if len(parts) == 0 {
			continue
		}

		opName := strings.ToUpper(parts[0])

		switch opName {
		// --- set the value (SET) ---
		case "SET":
			// SET [Dest] [Src]
			// e.g: SET x 10, SET x "hello", SET x y
			if len(parts) < 3 {
				return nil, fmt.Errorf("line %d: SET requires dest and src", i+1)
			}
			destVar, srcRaw := parts[1], parts[2]

			// 1. Load Source into R0 (auto detect number/string/variable)
			loadCmd, err := state.emitLoad(R0, srcRaw)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", i+1, err)
			}
			bytecode = append(bytecode, loadCmd...)

			// 2. Store R0 vào Dest Variable
			destAddr, err := state.GetOrAllocVar(destVar)
			if err != nil {
				return nil, err
			}
			bytecode = append(bytecode, emitMemOp(model.OP_STORE, R0, destAddr)...)

		// --- CALCULATE (2 Operands: Dest = Dest OP Src) ---
		// e.g: ADD x 10, SUB x y, CONCAT msg " world"
		case "ADD", "SUB", "MUL", "DIV", "MOD", "CONCAT", "AND", "OR", "XOR":
			if len(parts) != 3 {
				return nil, fmt.Errorf("line %d: %s requires 2 operands", i+1, opName)
			}
			destVar, srcRaw := parts[1], parts[2]

			// 1. Load Dest (Var) into R0
			destAddr, err := state.GetOrAllocVar(destVar)
			if err != nil {
				return nil, err
			} // Auto alloc if not exists (like behavior SET)
			bytecode = append(bytecode, emitMemOp(model.OP_LOAD, R0, destAddr)...)

			// 2. Load Src (Var/Num/Str) into R1
			loadCmd, err := state.emitLoad(R1, srcRaw)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", i+1, err)
			}
			bytecode = append(bytecode, loadCmd...)

			// 3. Execute Opcode
			var op byte
			switch opName {
			case "ADD":
				op = model.OP_ADD
			case "SUB":
				op = model.OP_SUB
			case "MUL":
				op = model.OP_MUL
			case "DIV":
				op = model.OP_DIV
			case "MOD":
				op = model.OP_MOD
			case "CONCAT":
				op = model.OP_CONCAT
			case "AND":
				op = model.OP_AND
			case "OR":
				op = model.OP_OR
			case "XOR":
				op = model.OP_XOR
			}
			bytecode = append(bytecode, op, R0, R1)

			// 4. Store Result R0 -> Dest
			bytecode = append(bytecode, emitMemOp(model.OP_STORE, R0, destAddr)...)

		// --- COMPARE (CMP) ---
		// e.g: CMP x 10, CMP 5 10 (compare literals)
		case "CMP":
			if len(parts) != 3 {
				return nil, fmt.Errorf("line %d: CMP requires 2 operands", i+1)
			}

			// Load Op1 -> R0 (canbe number or variable)
			cmd1, err1 := state.emitLoad(R0, parts[1])
			if err1 != nil {
				return nil, fmt.Errorf("line %d: op1 error: %v", i+1, err1)
			}
			bytecode = append(bytecode, cmd1...)

			// Load Op2 -> R1 (can be number or variable)
			cmd2, err2 := state.emitLoad(R1, parts[2])
			if err2 != nil {
				return nil, fmt.Errorf("line %d: op2 error: %v", i+1, err2)
			}
			bytecode = append(bytecode, cmd2...)

			bytecode = append(bytecode, model.OP_CMP, R0, R1)

		// --- PRINT ---
		case "PRINT_INT", "PRINT_STR":
			if len(parts) != 2 {
				return nil, fmt.Errorf("line %d: PRINT requires 1 operand", i+1)
			}

			// Load Op -> R0 (supports printing variables or direct numbers: PRINT 100)
			loadCmd, err := state.emitLoad(R0, parts[1])
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", i+1, err)
			}
			bytecode = append(bytecode, loadCmd...)

			op := model.OP_PRINT_INT
			if opName == "PRINT_STR" {
				op = model.OP_PRINT_STR
			}
			bytecode = append(bytecode, uint8(op), R0, 0)

		// --- CONTROL FLOW ---
		case "JMP", "JEQ", "JNE", "JGT", "JLT":
			val, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("line %d: jump target must be number", i+1)
			}

			op := getJumpOp(opName)
			bytecode = append(bytecode, op, 0, byte(val))

		case "HALT":
			bytecode = append(bytecode, model.OP_HALT, 0, 0)

		default:
			return nil, fmt.Errorf("line %d: unknown instruction '%s'", i+1, opName)
		}
	}

	bytecode = append(bytecode, model.OP_HALT, 0, 0)
	return bytecode, nil
}

func getJumpOp(name string) byte {
	switch name {
	case "JMP":
		return model.OP_JMP
	case "JEQ":
		return model.OP_JEQ
	case "JNE":
		return model.OP_JNE
	case "JGT":
		return model.OP_JGT
	case "JLT":
		return model.OP_JLT
	}
	return 0
}
