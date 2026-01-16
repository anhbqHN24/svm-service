package program

import (
	"fmt"
	"svm_whiteboard/internal/core"
)

type VM struct {
	R1, R2  int      // Registers (Compute Units)
	Flag    int      // Status Register (for comparisons)
	Program []byte   // The Smart Contract Bytecode
	PC      int      // Program Counter (Instruction Pointer)
	Output  []string // Logs (Sysvar: Log)
}

func NewVM(code []byte, p1, p2 int) *VM {
	return &VM{
		R1:      p1,
		R2:      p2,
		Flag:    0,
		Program: code,
		PC:      0,
		Output:  []string{},
	}
}

// Run executes the bytecode cycle by cycle
func (vm *VM) RunComplier() []string {
	// Safety limit to prevent infinite loops (Solana "Compute Budget")
	cycles := 0
	maxCycles := 100

	for vm.PC < len(vm.Program) && cycles < maxCycles {
		op := vm.Program[vm.PC]
		cycles++

		// Fetch Operand (Our architecture is fixed: 2 Bytes [OP, ARG])
		var operand int
		if vm.PC+1 < len(vm.Program) {
			operand = int(vm.Program[vm.PC+1])
		}

		switch op {
		case core.OP_LOAD_1:
			vm.R1 = operand
			vm.PC += 2
		case core.OP_LOAD_2:
			vm.R2 = operand
			vm.PC += 2
		case core.OP_ADD:
			vm.R1 = vm.R1 + vm.R2
			vm.Output = append(vm.Output, fmt.Sprintf("ADD: New R1=%d", vm.R1))
			vm.PC += 2
		case core.OP_SUB:
			vm.R1 = vm.R1 - vm.R2
			vm.Output = append(vm.Output, fmt.Sprintf("SUB: New R1=%d", vm.R1))
			vm.PC += 2
		case core.OP_PRINT:
			vm.Output = append(vm.Output, fmt.Sprintf(">> PRINT R1: %d", vm.R1))
			vm.PC += 2

		// --- Control Flow ---
		case core.OP_CMP:
			// Simulates CPU Compare: result is stored in Flag
			vm.Flag = vm.R1 - vm.R2
			vm.Output = append(vm.Output, fmt.Sprintf("CMP: %d vs %d (Flag=%d)", vm.R1, vm.R2, vm.Flag))
			vm.PC += 2

		case core.OP_JEQ: // Jump if Equal
			if vm.Flag == 0 {
				vm.Output = append(vm.Output, fmt.Sprintf("JEQ: Jumping to %d", operand))
				vm.PC = operand
			} else {
				vm.PC += 2
			}

		case core.OP_JGT: // Jump if Greater
			if vm.Flag > 0 {
				vm.Output = append(vm.Output, fmt.Sprintf("JGT: Jumping to %d", operand))
				vm.PC = operand
			} else {
				vm.PC += 2
			}

		case core.OP_JMP: // Always Jump
			vm.Output = append(vm.Output, fmt.Sprintf("JMP: Jumping to %d", operand))
			vm.PC = operand

		case core.OP_HALT:
			vm.Output = append(vm.Output, "HALT")
			return vm.Output

		case core.OP_NOOP:
			vm.PC += 1 // Skip padding

		default:
			vm.PC++
		}
	}

	if cycles >= maxCycles {
		vm.Output = append(vm.Output, "ERR: Compute Budget Exceeded")
	}

	return vm.Output
}
