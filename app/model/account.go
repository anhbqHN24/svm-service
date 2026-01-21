package model

import (
	"encoding/hex"
	"sync"
)

type Address string

func ToAddressPtr(a Address) *Address {
	return &a
}

type Account struct {
	Lamports   uint64     `json:"lamports"`
	Data       []byte     `json:"data"`
	Owner      *Address   `json:"owner"`      // The program that owns this account
	Executable bool       `json:"executable"` // True = Program, False = Data Account
	Mu         sync.Mutex `json:"-"`          // Mutex to lock this account during write operations
}

// Helper to format account data for JSON responses
type AccountView struct {
	*Account
	DataHex string `json:"data_hex"`
}

func NewAccountView(acc *Account) AccountView {
	return AccountView{
		Account: acc,
		DataHex: hex.EncodeToString(acc.Data),
	}
}
