package app

import (
	"bytes"
	"net/http"
	"svm_whiteboard/app/api"
	"svm_whiteboard/app/dto"
	"svm_whiteboard/app/model"
	"svm_whiteboard/app/service"
	"svm_whiteboard/helper"
	"svm_whiteboard/runtime"
	"sync"
	"time"
)

type Server struct {
	SessionStore map[string]*runtime.SVMMemoryManager
	SessionMu    sync.Mutex
}

func (s *Server) GetMemory(r *http.Request) *runtime.SVMMemoryManager {
	// 1. Get Session ID from Header (sent by React)
	sessionID := r.Header.Get("X-Session-ID")

	// If no ID provided, fallback to a "public" shared instance or a default one
	if sessionID == "" {
		sessionID = "default_public"
	}

	s.SessionMu.Lock()
	defer s.SessionMu.Unlock()

	// 2. Check if memory exists for this user
	if svm, exists := s.SessionStore[sessionID]; exists {
		return svm
	}

	// 3. If not, create a NEW isolated environment
	newSVM := runtime.NewSVMMemoryManager()
	newSVM.InitGenesisAccount() // Create their own Genesis account
	s.SessionStore[sessionID] = newSVM

	return newSVM
}

func (s *Server) HandleGetAllAccounts(w http.ResponseWriter, r *http.Request) {
	result, err := service.GetAllAccounts(s.GetMemory(r))
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	api.WriteResponseJSON(w, http.StatusOK, dto.APIResponse{Status: "success", Data: model.NewAccountViews(result)}, nil)
}

// GET /read
func (s *Server) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
	addr := r.PathValue("address")
	result, err := service.ReadAccount(s.GetMemory(r), addr)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	api.WriteResponseJSON(w, http.StatusOK, dto.APIResponse{Status: "success", Data: model.NewAccountView(result)}, nil)
}

// POST /write (Create Account)
func (s *Server) HandleWriteAccount(w http.ResponseWriter, r *http.Request) {
	request, err := helper.GetBodyInput[dto.WriteAccountRequest](r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := service.WriteAccount(s.GetMemory(r), request)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	api.WriteResponseJSON(w, http.StatusCreated, dto.APIResponse{Status: "success", Data: model.NewAccountView(result)}, nil)
}

// POST /execute
func (s *Server) HandleInteract(w http.ResponseWriter, r *http.Request) {
	request, err := helper.GetBodyInput[dto.ExecuteAccountRequest](r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := service.ExecuteAccount(s.GetMemory(r), request)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	api.WriteResponseJSON(w, http.StatusCreated, dto.APIResponse{Status: "success", Data: result}, nil)
}

func (s *Server) HandleCompileCode(w http.ResponseWriter, r *http.Request) {
	request, err := helper.GetBodyInput[dto.CompileCodeRequest](r)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	binaryData, err := service.CompileCode(request.SourceCode)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	reader := bytes.NewReader(binaryData)
	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename=program.bin")

	http.ServeContent(w, r, "program.bin", time.Now(), reader)
}
