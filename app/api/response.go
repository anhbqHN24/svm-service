package api

import (
	"encoding/json"
	"maps"
	"net/http"
	"svm_whiteboard/app/dto"
)

// Helper to write JSON response
func WriteResponseJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Add custom headers
	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// Helper to write standard error response
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	env := dto.APIResponse{
		Status:  "error",
		Message: message,
		Data:    nil,
	}

	// Note: You can add 500 error logging here if needed
	WriteResponseJSON(w, status, env, nil)
}
