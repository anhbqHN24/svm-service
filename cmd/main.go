package main

import (
	"fmt"
	"log"
	"net/http"
	"svm_whiteboard/api"
	"svm_whiteboard/internal/runtime"
)

func main() {
	// 1. Init Runtime (Engine)
	svmLedger := runtime.NewSVMLedger()
	svmLedger.InitGenesisAccount()

	// 2. Init Server (API Layer)
	server := &api.Server{SVMLedger: svmLedger}

	// 3. Define Routes
	http.HandleFunc("GET /read/{address}", server.HandleGetAccount)
	http.HandleFunc("/write", server.HandleCreateData)
	http.HandleFunc("/execute", server.HandleInteract)

	// 4. Start
	port := ":9924"
	fmt.Printf("Solana Modular SVM running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
