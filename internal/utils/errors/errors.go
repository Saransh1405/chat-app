package errors

import (
	"fmt"
	"net/http"
)

// ErrorDetail represents a field-specific error
type ErrorDetail struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
	Value string `json:"value,omitempty"`
}

// ErrorResponse represents the standardized error response format
type ErrorResponse struct {
	Error struct {
		Code      string       `json:"code"`
		Message   string       `json:"message"`
		Details   []ErrorDetail `json:"details,omitempty"`
		TraceID   string       `json:"traceId,omitempty"`
		Timestamp string       `json:"timestamp"`
	} `json:"error"`
}

// Standard error codes
const (
	ErrCodeValidationError   = "VALIDATION_ERROR"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeForbidden         = "FORBIDDEN"
	ErrCodeNotFound          = "RESOURCE_NOT_FOUND"
	ErrCodeConflict          = "RESOURCE_CONFLICT"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeDatabaseError     = "DATABASE_ERROR"
	ErrCodeWebSocketError    = "WEBSOCKET_ERROR"
)

// AppError represents an application error
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Details    []ErrorDetail
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details ...ErrorDetail) *AppError {
	e.Details = details
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// Predefined error constructors
func NewValidationError(message string, details ...ErrorDetail) *AppError {
	return NewAppError(ErrCodeValidationError, message, http.StatusBadRequest).WithDetails(details...)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError(ErrCodeForbidden, message, http.StatusForbidden)
}

func NewNotFoundError(message string) *AppError {
	return NewAppError(ErrCodeNotFound, message, http.StatusNotFound)
}

func NewConflictError(message string) *AppError {
	return NewAppError(ErrCodeConflict, message, http.StatusConflict)
}

func NewInternalError(message string) *AppError {
	return NewAppError(ErrCodeInternalError, message, http.StatusInternalServerError)
}

func NewDatabaseError(message string, err error) *AppError {
	return NewAppError(ErrCodeDatabaseError, message, http.StatusInternalServerError).WithError(err)
}

