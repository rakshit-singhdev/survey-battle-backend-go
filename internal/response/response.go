package response

import (
	"encoding/json"
	"net/http"
)

type envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type paginatedEnvelope struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Metadata any    `json:"metadata"`
	Data     any    `json:"data"`
}

// Success mirrors ApiResponse.success: {success: true, message, data}.
func Success(w http.ResponseWriter, data any, message string, statusCode int) {
	writeJSON(w, statusCode, envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Paginate mirrors ApiResponse.paginate: {success: true, message, metadata, data}.
func Paginate(w http.ResponseWriter, data any, metadata any, message string, statusCode int) {
	writeJSON(w, statusCode, paginatedEnvelope{
		Success:  true,
		Message:  message,
		Metadata: metadata,
		Data:     data,
	})
}

// Error mirrors ApiResponse.error: {success: false, message, code?, error}.
// `code` is the machine-readable reason; clients should branch on it rather
// than `message`, which is written for humans and can change.
func Error(w http.ResponseWriter, message string, statusCode int, code string, err any) {
	if err == nil {
		err = true
	}

	writeJSON(w, statusCode, envelope{
		Success: false,
		Message: message,
		Code:    code,
		Error:   err,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}
