package core

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
	Owner      *Address   `json:"owner"`      // what got the right to interact with account
	Executable bool       `json:"executable"` // True = Program, False = Data Account
	_          [7]byte    // Padding (Explicit or Compiler-added for alignment)
	Mu         sync.Mutex `json:"-"` // NEW: Fine-grained Lock. Dùng để khoá riêng account này khi đang có transaction ghi vào.
}

// Helper to print data neatly at JSON response
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
