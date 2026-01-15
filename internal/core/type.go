package core

import "encoding/hex"

// Constants
const (
	BPFLoaderID = "BPFLoader_1"
)

type Pubkey string

type Account struct {
	Key        Pubkey `json:"key"`
	Lamports   uint64 `json:"lamports"`
	Data       []byte `json:"data"`
	Owner      Pubkey `json:"owner"`      // what got the right to interact with account
	Executable bool   `json:"executable"` // True = Program, False = Data Account
}

// LogicFunction: Chữ ký hàm cho mọi Smart Contract
type LogicFunction func(programID Pubkey, accounts []*Account, input []byte) error

// Helper để in data đẹp hơn trong JSON response
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
