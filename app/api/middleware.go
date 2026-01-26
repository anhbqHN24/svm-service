package api

import (
	"net/http"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origin for simplicity; adjust as needed
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Allow specific methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")

		// Allow specific headers (important because you send Content-Type: application/json)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle Preflight Request (browser sends OPTIONS before POST/PUT)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue processing the main logic
		next.ServeHTTP(w, r)
	})
}
