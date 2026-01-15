package programs

import (
	"encoding/binary"
	"fmt"
	"svm_whiteboard/internal/core" // Import core types
)

// ProgramCounter: Logic tăng biến đếm
func ProgramCounter(programID core.Pubkey, accounts []*core.Account, input []byte) error {
	// 1. Lấy Data Account (account lưu state)
	if len(accounts) == 0 {
		return fmt.Errorf("missing data account")
	}
	dataAcc := accounts[0]

	// 2. Security Check (Quan trọng nhất): Chỉ Owner mới được sửa Data
	if dataAcc.Owner != programID {
		return fmt.Errorf("violation: Program %s is NOT the owner of %s", programID, dataAcc.Key)
	}

	// 3. Deserialize State
	var counter uint32
	if len(dataAcc.Data) >= 4 {
		counter = binary.LittleEndian.Uint32(dataAcc.Data)
	}

	// 4. Execute Logic
	counter++
	fmt.Printf(">>> [PROGRAM LOGIC] Counter incremented to: %d\n", counter)

	// 5. Serialize State back to Account
	newData := make([]byte, 4)
	binary.LittleEndian.PutUint32(newData, counter)
	dataAcc.Data = newData

	return nil
}
