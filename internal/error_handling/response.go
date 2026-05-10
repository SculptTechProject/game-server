package error_handling

import (
	"encoding/json"
	"log"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	resp := APIResponse{
		Success: true,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, appErr *AppError) {
	resp := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    appErr.Code,
			Message: appErr.Message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPStatus)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

func WriteErrorStatus(w http.ResponseWriter, status int, code Code, message string) {
	WriteError(w, &AppError{Code: code, Message: message, HTTPStatus: status})
}
