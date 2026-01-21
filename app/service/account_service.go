package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"svm_whiteboard/app/dto"
	"svm_whiteboard/app/model"
	"svm_whiteboard/app/program"
	"svm_whiteboard/helper"
	"svm_whiteboard/runtime"

	"github.com/near/borsh-go"
)

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

	// Serialize Account data using borsh format
	dataOnChain, err := borsh.Serialize(request.Data)
	if err != nil {
		return nil, errors.New(err.Error())
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
