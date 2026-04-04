package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorBody struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"invalid request"`
}

// ErrorResponse используется в swagger-документации.
type ErrorResponse struct {
	Error errorBody `json:"error"`
}

type errorResponse = ErrorResponse

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, errorResponse{
		Error: errorBody{Code: code, Message: message},
	})
}
