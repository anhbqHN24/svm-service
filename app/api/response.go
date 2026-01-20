package api

import (
	"encoding/json"
	"net/http"
	"svm_whiteboard/app/dto"
)

// Helper to send JSON success
func WriteResponseJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Thêm custom headers nếu cần
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// 3. Helper để gửi lỗi (Error Response)
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	env := dto.APIResponse{Status: "error", Message: message, Data: nil} // Standard format: {"error": "..."}

	// Có thể log lỗi ở đây nếu là lỗi 500
	// log.Println(message)

	WriteResponseJSON(w, status, env, nil)
}
