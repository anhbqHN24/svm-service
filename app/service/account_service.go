package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"svm_whiteboard/app/compiler"
	"svm_whiteboard/app/dto"
	"svm_whiteboard/app/model"
	"svm_whiteboard/app/program"
	"svm_whiteboard/helper"
	"svm_whiteboard/runtime"

	"github.com/near/borsh-go"
)

func GetAllAccounts(svm *runtime.SVMMemoryManager) ([]*model.Account, error) {
	accounts, ok := svm.GetAllAccounts()
	if !ok {
		return nil, errors.New("No Accounts Found")
	}
	return accounts, nil
}

func ReadAccount(svm *runtime.SVMMemoryManager, addr string) (*model.Account, error) {
	acc, ok := svm.GetAccount(model.Address(addr))
	if !ok {
		return nil, errors.New("Invalid Account Address")
	}
	return acc, nil
}

func WriteAccount(svm *runtime.SVMMemoryManager, request dto.WriteAccountRequest) (*model.Account, error) {
	progAcc, ok := svm.GetAccount(request.Owner)
	// In Solana, the owner field must be an executable program Account
	if !ok || !progAcc.Executable {
		return nil, errors.New("Invalid Owner Address")
	}

	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	newAccountAddress := model.Address(helper.EncodePubKeyToString(pubKey))

	var dataOnChain []byte

	if request.Executable {
		// [NEW LOGIC] Check inputs type: Raw Bytes (Array) or Source Code (String)
		switch v := request.Data.(type) {
		case string:
			// Existing logic for JSON array input
			bytes, err := hex.DecodeString(v)
			if err != nil {
				return nil, err
			}
			dataOnChain = bytes
		default:
			return nil, errors.New("invalid data format for executable account")
		}
	} else {
		// Logic for Data Account (Serialize via Borsh)
		dataOnChain, err = borsh.Serialize(request.Data)
		if err != nil {
			return nil, errors.New(err.Error())
		}
	}

	dataAcc := &model.Account{
		Lamports:   50, // Not yet implemented -> set fixed rent fee for now
		Owner:      model.ToAddressPtr(request.Owner),
		Executable: request.Executable,
		Data:       dataOnChain,
	}

	svm.SetAccount(newAccountAddress, dataAcc)
	return dataAcc, nil
}

func CompileCode(sourceCode string) ([]byte, error) {
	compiledBytes, err := compiler.Compile(sourceCode)
	if err != nil {
		return nil, errors.New("Compilation Error: " + err.Error())
	}
	return compiledBytes, nil
}

func ExecuteAccount(svm *runtime.SVMMemoryManager, request dto.ExecuteAccountRequest) (*dto.ExecuteAccountResponse, error) {
	progAcc, ok := svm.GetAccount(request.ProgAddr)
	if !ok || !progAcc.Executable {
		return nil, errors.New("Invalid Program Account Address")
	}

	// Estimate computation cost of program binary
	estimatedCost, err := model.EstimateComputeCost(progAcc.Data)
	if err != nil {
		return nil, errors.New("Bad Bytecode: " + err.Error())
	}

	if estimatedCost > model.MaxComputeCycle {
		return nil, errors.New("Bytecode Compute Budget Exceeded")
	}

	var (
		vm   *program.VM
		logs []string
	)
	// --- STEP 2: Retrieve Data Account & OWNERSHIP CHECK (Module 1) ---
	if request.DataAddr != "" {
		dataAcc, ok := svm.GetAccount(request.DataAddr)
		if !ok {
			return nil, errors.New("Invalid Data Account Address")
		}

		// Check Ownership: Is the Program the Owner of this Data Account?
		if dataAcc.Owner != &request.ProgAddr {
			return nil, errors.New("Invalid Account Owner")
		}

		// --- STEP 3: ATOMIC LOCKING (Lock 2 Accounts) ---
		// To prevent Deadlocks (e.g., A locks X waiting for Y, B locks Y waiting for X),
		// we always lock in Address order (lowest to highest).

		// Lock sorting logic
		firstLock, secondLock := &progAcc.Mu, &dataAcc.Mu
		if request.ProgAddr > request.DataAddr {
			firstLock, secondLock = &dataAcc.Mu, &progAcc.Mu
		}

		firstLock.Lock()  // Lock Program Account
		secondLock.Lock() // Lock Data Account

		// Use defer to ensure unlocking when function ends (even on panic)
		defer secondLock.Unlock()
		defer firstLock.Unlock()

		// --- STEP 4: EXECUTE VM ---
		// Simulation: Load old data from Data Account into Register 2
		currentVal := int(0)
		if len(dataAcc.Data) >= 4 {
			currentVal = int(binary.LittleEndian.Uint32(dataAcc.Data))
		}

		vm, logs, err = handleVMExecution(progAcc, request.Params.Param1, currentVal)
		if err != nil {
			return nil, err
		}

		// --- STEP 5: UPDATE STATE ---
		// Get result from R1 after execution to update Data Account
		newVal := uint32(vm.GetRegister1())

		// Overwrite new data into Data Account (Currently Safe Locked)
		newStateBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(newStateBytes, newVal)
		dataAcc.Data = newStateBytes
		svm.SetAccount(request.DataAddr, dataAcc)
	} else {
		_, logs, err = handleVMExecution(progAcc, request.Params.Param1, request.Params.Param2)
		if err != nil {
			return nil, err
		}
	}
	return &dto.ExecuteAccountResponse{
		ProgAddr:    request.ProgAddr,
		DataAddr:    request.DataAddr,
		ComputeCost: estimatedCost,
		Logs:        logs,
	}, nil
}

func handleVMExecution(progAcc *model.Account, param1 any, param2 any) (*program.VM, []string, error) {
	vm := program.NewVM(progAcc.Data)
	// Load exactly 2 params from struct into R0, R1
	if err := program.LoadStrictParams(vm, param1, param2); err != nil {
		return nil, nil, err
	}
	// 4. Run VM program execution & return logs
	logs, err := vm.Run()
	if err != nil {
		return nil, nil, err
	}
	return vm, logs, nil
}
