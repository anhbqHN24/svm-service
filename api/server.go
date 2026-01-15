package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"svm_whiteboard/helper"
	"svm_whiteboard/internal/core"
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
func (s *Server) HandleCreateData(w http.ResponseWriter, r *http.Request) {
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

	vm := core.NewVM(progAcc.Data, req.Params.Param1, req.Params.Param2)
	logs := vm.Run()

	response := map[string]interface{}{
		"Account":      req.Address,
		"logs":         logs,
		"final_result": vm.R1, // Trả về giá trị cuối cùng của R1
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(APIResponse{Status: "success", Message: "Tx Finalized", Data: response})
}
