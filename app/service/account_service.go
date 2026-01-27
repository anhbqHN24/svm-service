package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"svm_whiteboard/app/compiler"
	"svm_whiteboard/app/dto"
	"svm_whiteboard/app/model"
	"svm_whiteboard/app/program"
	"svm_whiteboard/helper"
	"svm_whiteboard/runtime"
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
		dataOnChain, err = helper.SerializePrimitive(request.Data)
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

	paramsToLoad := []dto.ExecuteParam{request.Params.Param1, request.Params.Param2}

	// --- STEP 2: Load Params into VM Stack/Heap ---
	for i, p := range paramsToLoad {
		switch p.Type {
		case helper.PARAM_TYPE_PRIMITIVE: // Primitive (int hoặc string)
			if valStr, ok := p.Value.(string); ok {
				// Nếu là string, nạp vào Heap và đưa con trỏ vào Stack
				p.Value = valStr
			} else if valNum, ok := p.Value.(float64); ok {
				// JSON unmarshal thường coi số là float64
				p.Value = int(valNum)
			} else {
				return nil, fmt.Errorf("param_%d: unsupported primitive type", i+1)
			}

		case helper.PARAM_TYPE_ADDRESS: // Address (Data Account)
			addrStr, ok := p.Value.(string)
			if !ok {
				return nil, fmt.Errorf("param_%d: address must be a string", i+1)
			}
			// read data account from SVM
			dataAcc, ok := svm.GetAccount(model.Address(addrStr))
			if !ok {
				return nil, fmt.Errorf("data account %s not found", addrStr)
			}
			// set account field for later use
			p.Account = dataAcc

			// Check Ownership: Is the Program the Owner of this Data Account?
			if dataAcc.Owner != &request.ProgAddr {
				return nil, errors.New("Invalid Account Ownership for address: " + string(p.Account.Key))
			}

			// --- STEP 3: ATOMIC LOCKING (Lock 2 Accounts) ---
			// To prevent Deadlocks (e.g., A locks X waiting for Y, B locks Y waiting for X),
			// we always lock in Address order (lowest to highest).

			// Lock sorting logic
			firstLock, secondLock := &progAcc.Mu, &dataAcc.Mu
			if request.ProgAddr > model.Address(addrStr) {
				firstLock, secondLock = &dataAcc.Mu, &progAcc.Mu
			}

			firstLock.Lock()  // Lock Program Account
			secondLock.Lock() // Lock Data Account

			// Use defer to ensure unlocking when function ends (even on panic)
			defer secondLock.Unlock()
			defer firstLock.Unlock()

			dataContent, err := helper.DeserializePrimitive(dataAcc.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to deserialize data account %s: %v", addrStr, err)
			}
			p.Value = dataContent

		default:
			return nil, fmt.Errorf("invalid param type: %d", p.Type)
		}
	}

	vm, logs, err = handleVMExecution(progAcc, request.Params.Param1.Value, request.Params.Param2.Value)
	if err != nil {
		return nil, err
	}

	if request.Params.Param1.Type == helper.PARAM_TYPE_ADDRESS {
		request.Params.Param1.Account.Data, err = helper.SerializePrimitive(vm.GetStackValue(0))
		if err != nil {
			return nil, fmt.Errorf("failed to serialize param_1 back to account: %v", err)
		}
		svm.SetAccount(request.Params.Param1.Account.Key, request.Params.Param1.Account)
	}
	if request.Params.Param2.Type == helper.PARAM_TYPE_ADDRESS {
		request.Params.Param2.Account.Data, err = helper.SerializePrimitive(vm.GetStackValue(1))
		if err != nil {
			return nil, fmt.Errorf("failed to serialize param_2 back to account: %v", err)
		}
		svm.SetAccount(request.Params.Param2.Account.Key, request.Params.Param2.Account)
	}

	// --- STEP 4: RETURN RESPONSE ---

	return &dto.ExecuteAccountResponse{
		ProgAddr:    request.ProgAddr,
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
