package handlers

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, data interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, status int) error {

	type ErrorResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	return writeJSON(w, ErrorResponse{Success: false, Message: message}, status)
}
