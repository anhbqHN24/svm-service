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
	"time"
)

type Server struct {
	SVMMemoryManager *runtime.SVMMemoryManager
}

func (s *Server) HandleGetAllAccounts(w http.ResponseWriter, r *http.Request) {
	result, err := service.GetAllAccounts(s.SVMMemoryManager)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	api.WriteResponseJSON(w, http.StatusOK, dto.APIResponse{Status: "success", Data: model.NewAccountViews(result)}, nil)
}

// GET /read
func (s *Server) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
	addr := r.PathValue("address")
	result, err := service.ReadAccount(s.SVMMemoryManager, addr)
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

	result, err := service.WriteAccount(s.SVMMemoryManager, request)
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

	result, err := service.ExecuteAccount(s.SVMMemoryManager, request)
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
