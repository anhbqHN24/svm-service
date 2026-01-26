package main

import (
	"fmt"
	"log"
	"net/http"
	"svm_whiteboard/app"
	"svm_whiteboard/app/api"
	"svm_whiteboard/runtime"
)

func main() {
	// 1. Init Runtime (Engine)
	SVMMemoryManager := runtime.NewSVMMemoryManager()
	SVMMemoryManager.InitGenesisAccount()

	// 2. Init Server (API Layer)
	server := &app.Server{SVMMemoryManager: SVMMemoryManager}

	mux := http.NewServeMux()

	// 3. Define Routes
	mux.HandleFunc("GET /all", server.HandleGetAllAccounts)
	mux.HandleFunc("GET /read/{address}", server.HandleGetAccount)
	mux.HandleFunc("POST /write", server.HandleWriteAccount)
	mux.HandleFunc("POST /execute", server.HandleInteract)
	mux.HandleFunc("POST /compile", server.HandleCompileCode)

	// 4. Start
	port := ":9924"
	fmt.Printf("Solana Modular SVM running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, api.CORSMiddleware(mux)))
}
