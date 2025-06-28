package errors

import (
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeNotFound indicates a resource was not found
	ErrorTypeNotFound ErrorType = "NOT_FOUND"
	// ErrorTypeValidation indicates a validation error
	ErrorTypeValidation ErrorType = "VALIDATION"
	// ErrorTypeInternal indicates an internal error
	ErrorTypeInternal ErrorType = "INTERNAL"
	// ErrorTypeConflict indicates a conflict error
	ErrorTypeConflict ErrorType = "CONFLICT"
	// ErrorTypePermission indicates a permission error
	ErrorTypePermission ErrorType = "PERMISSION"
)

// Error represents a custom error with additional context
type Error struct {
	Type    ErrorType
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Err
}

// NewError creates a new error
func NewError(errType ErrorType, message string, err error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	e.Context[key] = value
	return e
}

// Common error constructors

// NotFound creates a not found error
func NotFound(resource string, id string) *Error {
	return NewError(ErrorTypeNotFound, fmt.Sprintf("%s not found: %s", resource, id), nil).
		WithContext("resource", resource).
		WithContext("id", id)
}

// ValidationError creates a validation error
func ValidationError(field string, message string) *Error {
	return NewError(ErrorTypeValidation, fmt.Sprintf("validation failed for %s: %s", field, message), nil).
		WithContext("field", field)
}

// InternalError creates an internal error
func InternalError(message string, err error) *Error {
	return NewError(ErrorTypeInternal, message, err)
}

// ConflictError creates a conflict error
func ConflictError(resource string, message string) *Error {
	return NewError(ErrorTypeConflict, fmt.Sprintf("conflict in %s: %s", resource, message), nil).
		WithContext("resource", resource)
}

// PermissionError creates a permission error
func PermissionError(action string, resource string) *Error {
	return NewError(ErrorTypePermission, fmt.Sprintf("permission denied for %s on %s", action, resource), nil).
		WithContext("action", action).
		WithContext("resource", resource)
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeNotFound
	}
	return false
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeValidation
	}
	return false
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeInternal
	}
	return false
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypeConflict
	}
	return false
}

// IsPermission checks if an error is a permission error
func IsPermission(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrorTypePermission
	}
	return false
}