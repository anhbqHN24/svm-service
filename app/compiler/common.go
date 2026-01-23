package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"svm_whiteboard/app/model"
)

// Map alias registers/params to internal Register Index
var regMap = map[string]byte{
	"R0": 0, "R1": 1, "R2": 2, "R3": 3,
	"param_1": 0, // param_1 is loaded into R0
	"param_2": 1, // param_2 is loaded into R1
}

func Compile(source string) ([]byte, error) {
	var bytecode []byte
	lines := strings.Split(source, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		parts := strings.Split(line, ",")
		// Trim spaces for each part
		for j := range parts {
			parts[j] = strings.TrimSpace(parts[j])
		}

		opName := strings.ToUpper(parts[0])

		switch opName {
		case "SET":
			// FORMAT: SET, Dest, Src (Value or Reg)
			if len(parts) != 3 {
				return nil, fmt.Errorf("line %d: SET requires 2 arguments", i+1)
			}
			dest, ok := regMap[parts[1]]
			if !ok {
				return nil, fmt.Errorf("line %d: invalid destination register %s", i+1, parts[1])
			}

			// Check if Src is a Number (LOAD_IMM) or Register (MOV)
			if val, err := strconv.Atoi(parts[2]); err == nil {
				// Case 1: Src is Number -> LOAD_IMM
				bytecode = append(bytecode, model.OP_LOAD_IMM, dest, byte(val))
			} else if src, ok := regMap[parts[2]]; ok {
				// Case 2: Src is Register/Param -> MOV
				bytecode = append(bytecode, model.OP_MOV, dest, src)
			} else {
				return nil, fmt.Errorf("line %d: invalid source %s", i+1, parts[2])
			}

		case "ADD", "SUB", "MUL", "DIV", "MOD":
			// FORMAT: ADD, Dest, Src
			if len(parts) != 3 {
				return nil, fmt.Errorf("line %d: %s requires 2 arguments", i+1, opName)
			}
			dest, ok1 := regMap[parts[1]]
			src, ok2 := regMap[parts[2]]
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("line %d: invalid registers", i+1)
			}

			var opcode byte
			switch opName {
			case "ADD":
				opcode = model.OP_ADD
			case "SUB":
				opcode = model.OP_SUB
			case "MUL":
				opcode = model.OP_MUL
			case "DIV":
				opcode = model.OP_DIV
			case "MOD":
				opcode = model.OP_MOD
			}
			bytecode = append(bytecode, opcode, dest, src)

		case "PRINT":
			// FORMAT: PRINT, Reg
			if len(parts) != 2 {
				return nil, fmt.Errorf("line %d: PRINT requires 1 argument", i+1)
			}
			reg, ok := regMap[parts[1]]
			if !ok {
				return nil, fmt.Errorf("line %d: invalid register", i+1)
			}
			bytecode = append(bytecode, model.OP_PRINT_INT, reg, 0)

		case "HALT":
			bytecode = append(bytecode, model.OP_HALT, 0, 0)

		default:
			return nil, fmt.Errorf("line %d: unknown instruction %s", i+1, opName)
		}
	}
	// Auto append HALT if missing (optional safety)
	bytecode = append(bytecode, model.OP_HALT, 0, 0)

	return bytecode, nil
}
