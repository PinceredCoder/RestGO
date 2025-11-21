package errors

import (
	"encoding/json"
	"net/http"
)

type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND"
	ErrorTypeBadRequest   ErrorType = "BAD_REQUEST"
	ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
)

type APIError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewValidationError(message string, details any) *APIError {
	return &APIError{
		Type:    ErrorTypeValidation,
		Message: message,
		Details: details,
	}
}

func NewNotFoundError(message string) *APIError {
	return &APIError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

func NewBadRequestError(message string) *APIError {
	return &APIError{
		Type:    ErrorTypeBadRequest,
		Message: message,
	}
}

func NewInternalError(message string) *APIError {
	return &APIError{
		Type:    ErrorTypeInternal,
		Message: message,
	}
}

func RespondWithError(w http.ResponseWriter, statusCode int, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(err)
}

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
