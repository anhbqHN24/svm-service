package runtime

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"svm_whiteboard/helper"
	"svm_whiteboard/internal/core"
	"sync"
)

type SVMLedger struct {
	Accounts map[core.Pubkey]*core.Account // Account: Address public key - pointer to mem zone of account
	mu       sync.RWMutex                  // Thread safe
}

func NewSVMLedger() *SVMLedger {
	svm := &SVMLedger{
		Accounts: make(map[core.Pubkey]*core.Account),
	}
	return svm
}

func (svm *SVMLedger) GetAccount(key core.Pubkey) (*core.Account, bool) {
	svm.mu.RLock()
	defer svm.mu.RUnlock()
	acc, ok := svm.Accounts[key]
	return acc, ok
}

func (svm *SVMLedger) SetAccount(acc *core.Account) {
	svm.mu.Lock()
	defer svm.mu.Unlock()
	svm.Accounts[acc.Key] = acc
}

// BootstrapSystem: Tạo các account khởi thủy (Genesis)
func (svm *SVMLedger) InitGenesisAccount() {
	// Tạo Program Account (Executable = True)
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	AccAddress := core.Pubkey(helper.EncodePubKeyToString(pubKey))
	progAcc := &core.Account{
		Key:        AccAddress,
		Owner:      core.BPFLoaderID, // Program belong to Loader
		Executable: true,
		Data: []byte{
			core.OP_ADD, 0,
			core.OP_PRINT, 0,
			core.OP_LOAD_2, 50,
			core.OP_ADD, 0,
			core.OP_PRINT, 0,
			core.OP_HALT, 0,
		},
	}
	svm.SetAccount(progAcc)
	fmt.Printf("--- GENESIS: Program Account %s created ---\n", AccAddress)
}
