package dto

import "time"

// StandardResponse wraps all API responses
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	TraceID   string      `json:"traceId,omitempty"`
}

// PaginationMeta includes pagination info
type PaginationMeta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	Total     int `json:"total"`
	TotalPage int `json:"totalPage"`
}

// AppError represents application errors with context
type AppError struct {
	Code    string
	Message string
	Status  int
	Details map[string]interface{}
}

// Common error codes
const (
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeRateLimit       = "RATE_LIMIT_EXCEEDED"
	ErrCodeUnavailable     = "SERVICE_UNAVAILABLE"
	ErrCodeDatabaseError   = "DATABASE_ERROR"
)

// NewAppError creates an application error
func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds contextual details to an error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}
