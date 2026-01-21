package runtime

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"svm_whiteboard/app/model"
	"svm_whiteboard/helper"
	"sync"
)

type SVMMemoryManager struct {
	ProgramCache sync.Map                         // sync.Map is optimized for stable keys and concurrent reads.
	Accounts     map[model.Address]*model.Account // Account: Address - pointer to allocated memory zone of account
	DataMu       sync.RWMutex                     // Lock to protect Memory map structure
}

func NewSVMMemoryManager() *SVMMemoryManager {
	svm := &SVMMemoryManager{
		Accounts: make(map[model.Address]*model.Account),
	}
	return svm
}

func (svm *SVMMemoryManager) GetAccount(addr model.Address) (*model.Account, bool) {
	// Priority Check: Look in Program Cache first (Lock-free path)
	if val, ok := svm.ProgramCache.Load(addr); ok {
		return val.(*model.Account), true
	}

	// 1. Manager Lock (Level 1): just using to FIND pointer
	svm.DataMu.RLock()
	defer svm.DataMu.RUnlock()
	acc, exists := svm.Accounts[addr]
	return acc, exists
}

func (svm *SVMMemoryManager) SetAccount(addr model.Address, acc *model.Account) {
	// 1. Account is program: using "Write-Once, Read-Many (WORM)" strategy, allow parallel reading without mutex lock
	if acc.Executable {
		svm.ProgramCache.Store(addr, acc)
		return
	}
	svm.DataMu.Lock()         // Lock the whole Memory Map (Write Lock)
	defer svm.DataMu.Unlock() // Unlock right after it done

	svm.Accounts[addr] = acc
}

// BootstrapSystem: Create new Genesis Account (parent of all other account, similar to Wallet Account)
func (svm *SVMMemoryManager) InitGenesisAccount() {
	// Create Program Account (Executable = True)
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	AccAddress := model.Address(helper.EncodePubKeyToString(pubKey))
	progAcc := &model.Account{
		Owner:      nil,
		Executable: true,
		Data: []byte{ // Simple Addition of 2 param
			model.OP_ADD, 0, 1, // ADD R0, R1
			model.OP_PRINT_INT, 0, 0, // PRINT_INT R0
			model.OP_HALT, 0, 0, // HALT
		},
	}
	svm.SetAccount(AccAddress, progAcc)
	fmt.Printf("--- GENESIS: Program Account %s created ---\n", AccAddress)
}
