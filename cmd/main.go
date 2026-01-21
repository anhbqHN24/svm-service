package main

import (
	"fmt"
	"log"
	"net/http"
	"svm_whiteboard/app"
	"svm_whiteboard/runtime"
)

func main() {
	// 1. Init Runtime (Engine)
	SVMMemoryManager := runtime.NewSVMMemoryManager()
	SVMMemoryManager.InitGenesisAccount()

	// 2. Init Server (API Layer)
	server := &app.Server{SVMMemoryManager: SVMMemoryManager}

	// 3. Define Routes
	http.HandleFunc("GET /read/{address}", server.HandleGetAccount)
	http.HandleFunc("POST /write", server.HandleWriteAccount)
	http.HandleFunc("POST /execute", server.HandleInteract)

	// 4. Start
	port := ":9924"
	fmt.Printf("Solana Modular SVM running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
