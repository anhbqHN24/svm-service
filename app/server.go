package app

import (
	"encoding/binary"
	"net/http"
	"svm_whiteboard/app/api"
	"svm_whiteboard/app/dto"
	"svm_whiteboard/app/model"
	"svm_whiteboard/app/program"
	"svm_whiteboard/app/service"
	"svm_whiteboard/helper"
	"svm_whiteboard/runtime"
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

	progAcc, ok := s.SVMMemoryManager.GetAccount(request.ProgAddr)
	if !ok || !progAcc.Executable {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid Program Account Address")
		return
	}

	// Estimate computation cost of program binary
	estimatedCost, err := model.EstimateComputeCost(progAcc.Data)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Bad Bytecode: "+err.Error())
		return
	}

	if estimatedCost > model.MaxComputeCycle {
		api.ErrorResponse(w, http.StatusBadRequest, "Bytecode Compute Budget Exceeded")
		return
	}

	var (
		vm   *program.VM
		logs []string
	)
	// --- STEP 2: Retrieve Data Account & OWNERSHIP CHECK (Module 1) ---
	if request.DataAddr != "" {
		dataAcc, ok := s.SVMMemoryManager.GetAccount(request.DataAddr)
		if !ok {
			api.ErrorResponse(w, http.StatusBadRequest, "Invalid Data Account Address")
			return
		}

		// Check Ownership: Is the Program the Owner of this Data Account?
		if dataAcc.Owner != &request.ProgAddr {
			api.ErrorResponse(w, http.StatusBadRequest, "Invalid Account Owner")
			return
		}

		// --- STEP 3: ATOMIC LOCKING (Lock 2 Accounts) ---
		// To prevent Deadlocks (e.g., A locks X waiting for Y, B locks Y waiting for X),
		// we always lock in Address order (lowest to highest).

		// Lock sorting logic
		firstLock, secondLock := &progAcc.Mu, &dataAcc.Mu
		if request.ProgAddr > request.DataAddr {
			firstLock, secondLock = &dataAcc.Mu, &progAcc.Mu
		}

		firstLock.Lock()  // Lock Program Account
		secondLock.Lock() // Lock Data Account

		// Use defer to ensure unlocking when function ends (even on panic)
		defer secondLock.Unlock()
		defer firstLock.Unlock()

		// --- STEP 4: EXECUTE VM ---
		// Simulation: Load old data from Data Account into Register 2
		currentVal := int(0)
		if len(dataAcc.Data) >= 4 {
			currentVal = int(binary.LittleEndian.Uint32(dataAcc.Data))
		}

		vm, logs, err = handleVMExecution(progAcc, request.Params.Param1, currentVal)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// --- STEP 5: UPDATE STATE ---
		// Get result from R1 after execution to update Data Account
		newVal := uint32(vm.GetRegister1())

		// Overwrite new data into Data Account (Currently Safe Locked)
		newStateBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(newStateBytes, newVal)
		dataAcc.Data = newStateBytes
		s.SVMMemoryManager.SetAccount(request.DataAddr, dataAcc)
	} else {
		_, logs, err = handleVMExecution(progAcc, request.Params.Param1, request.Params.Param2)
		if err != nil {
			api.ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	api.WriteResponseJSON(w, http.StatusCreated, dto.APIResponse{
		Status: "success",
		Data: dto.ExecuteAccountResponse{
			ProgAddr:    request.ProgAddr,
			DataAddr:    request.DataAddr,
			ComputeCost: estimatedCost,
			Logs:        logs,
		},
	}, nil)
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
