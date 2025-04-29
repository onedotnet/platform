// Package errors provides standardized error types and handling for the platform
package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with context
type AppError struct {
	// HTTP status code to return
	StatusCode int `json:"-"`

	// Error code (unique identifier for this error)
	Code string `json:"code"`

	// User-facing error message
	Message string `json:"message"`

	// Detailed error message (not exposed to the client)
	Detail string `json:"-"`

	// Original error if wrapping one
	Err error `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap supports the errors.Is and errors.As functions
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithError adds an underlying error to the AppError
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// WithDetail adds detailed information to the AppError
func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

// Common error constructors

// BadRequest creates a 400 Bad Request error
func BadRequest(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    message,
	}
}

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
	}
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) *AppError {
	if message == "" {
		message = "You don't have permission to access this resource"
	}
	return &AppError{
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    message,
	}
}

// NotFound creates a 404 Not Found error
func NotFound(message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    message,
	}
}

// Conflict creates a 409 Conflict error
func Conflict(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    message,
	}
}

// ValidationError creates a 422 Unprocessable Entity error
func ValidationError(message string) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return &AppError{
		StatusCode: http.StatusUnprocessableEntity,
		Code:       "VALIDATION_FAILED",
		Message:    message,
	}
}

// InternalServer creates a 500 Internal Server Error
func InternalServer(message string) *AppError {
	if message == "" {
		message = "An internal server error occurred"
	}
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    message,
	}
}

// ServiceUnavailable creates a 503 Service Unavailable error
func ServiceUnavailable(message string) *AppError {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return &AppError{
		StatusCode: http.StatusServiceUnavailable,
		Code:       "SERVICE_UNAVAILABLE",
		Message:    message,
	}
}

// FromError converts a standard error to an AppError
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	// Default to internal server error
	return InternalServer("").WithError(err)
}
