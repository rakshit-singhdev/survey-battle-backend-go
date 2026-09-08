package apperror

import "fmt"

// AppError mirrors utils/errorHandler/app.error.ts: an operational error that
// carries the HTTP status and an optional machine-readable code a client can
// branch on instead of the human-readable message.
type AppError struct {
	Message    string
	StatusCode int
	Code       string
}

func New(message string, statusCode int) *AppError {
	return &AppError{Message: message, StatusCode: statusCode}
}

func NewWithCode(message string, statusCode int, code string) *AppError {
	return &AppError{Message: message, StatusCode: statusCode, Code: code}
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s (status %d)", e.Message, e.StatusCode)
}
