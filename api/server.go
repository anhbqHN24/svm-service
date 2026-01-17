package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"svm_whiteboard/helper"
	"svm_whiteboard/internal/core"
	"svm_whiteboard/internal/program"
	"svm_whiteboard/internal/runtime"

	"github.com/labyla/borsh-go"
)

type Server struct {
	SVMLedger *runtime.SVMLedger
}

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// GET /read
func (s *Server) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	addr := r.PathValue("address")
	acc, ok := s.SVMLedger.GetAccount(core.Pubkey(addr))
	if !ok {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Not found"})
		return
	}
	json.NewEncoder(w).Encode(APIResponse{Status: "success", Data: core.NewAccountView(acc)})
}

// POST /write (Tạo Account)
func (s *Server) HandleCreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OwnerID    string `json:"owner_id"`
		Data       any    `json:"data"`
		Executable bool   `json:"executable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	progAcc, ok := s.SVMLedger.GetAccount(core.Pubkey(req.OwnerID))
	// in Solana, owner field must be executable program Account
	if !ok || !progAcc.Executable {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid Owner Address"})
		return
	}

	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	newDataID := core.Pubkey(helper.EncodePubKeyToString(pubKey))

	dataOnChain, err := borsh.Serialize(req.Data)
	if err != nil {
		log.Fatal(err)
	}

	dataAcc := &core.Account{
		Key:        newDataID,
		Lamports:   50, //fixed rent fee amount at the moment for the sake of whiteboard purpose
		Owner:      core.Pubkey(req.OwnerID),
		Executable: req.Executable,
		Data:       dataOnChain,
	}
	s.SVMLedger.SetAccount(dataAcc)
	json.NewEncoder(w).Encode(APIResponse{Status: "success", Data: map[string]string{"data_address": string(newDataID)}})
}

// POST /execute
func (s *Server) HandleInteract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Address string `json:"address"`
		Params  struct {
			Param1 int `json:"param_1"`
			Param2 int `json:"param_2"`
		} `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	progAcc, ok := s.SVMLedger.GetAccount(core.Pubkey(req.Address))
	if !ok || !progAcc.Executable {
		json.NewEncoder(w).Encode(APIResponse{Status: "error", Message: "Invalid Account Address"})
		return
	}

	vm := program.NewVM(progAcc.Data)

	// Load đúng 2 params từ struct vào R0, R1
	if err := program.LoadStrictParams(vm, req.Params.Param1, req.Params.Param2); err != nil {
		http.Error(w, "Input Error: "+err.Error(), 400)
		return
	}

	// 4. Run VM program execution & Return logs
	logs, err := vm.Run()

	response := map[string]interface{}{
		"Account":        req.Address,
		"logs":           logs,
		"final_register": vm.Registers, // return the final result
	}
	if err != nil {
		http.Error(w, "Error: "+err.Error(), 400)
		return
	}
	json.NewEncoder(w).Encode(APIResponse{Status: "success", Message: "Tx Finalized", Data: response})
}
