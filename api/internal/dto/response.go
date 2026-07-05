package dto

import (
	"maps"
	"time"
)

// StandardResponse wraps all API responses
type StandardResponse struct {
	Success   bool      `json:"success"`
	Data      any       `json:"data,omitempty"`
	Error     string    `json:"error,omitempty"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"traceId,omitempty"`
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
	Details map[string]any
}

// Common error codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeRateLimit     = "RATE_LIMIT_EXCEEDED"
	ErrCodeUnavailable   = "SERVICE_UNAVAILABLE"
	ErrCodeDatabaseError = "DATABASE_ERROR"
)

// NewAppError creates an application error
func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Details: make(map[string]any),
	}
}

// WithDetails adds contextual details to an error
func (e *AppError) WithDetails(details map[string]any) *AppError {
	maps.Copy(e.Details, details)
	return e
}

// CreateProjectRequest represents the input for creating a new project
type CreateProjectRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=200"`
	Description string  `json:"description"`
	StatusID    int32   `json:"statusId" validate:"required,min=1"`
	OwnerID     *string `json:"ownerId,omitempty"`
	DueAt       *string `json:"dueAt,omitempty"`
}
