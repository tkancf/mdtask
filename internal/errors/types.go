package errors

import (
	"fmt"
)

// ErrorType represents the type of error
type ErrorType int

const (
	// ErrNotFound indicates a resource was not found
	ErrNotFound ErrorType = iota
	// ErrInvalidInput indicates invalid user input
	ErrInvalidInput
	// ErrInternal indicates an internal error
	ErrInternal
	// ErrDuplicate indicates a duplicate resource
	ErrDuplicate
	// ErrPermission indicates a permission error
	ErrPermission
	// ErrValidation indicates a validation error
	ErrValidation
	// ErrConflict indicates a conflict error
	ErrConflict
)

// AppError represents an application error with type information
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the error is of a specific type
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// NotFound creates a not found error
func NotFound(resource string, id string) error {
	return &AppError{
		Type:    ErrNotFound,
		Message: fmt.Sprintf("%s not found: %s", resource, id),
	}
}

// InvalidInput creates an invalid input error
func InvalidInput(field string, value interface{}) error {
	return &AppError{
		Type:    ErrInvalidInput,
		Message: fmt.Sprintf("invalid %s: %v", field, value),
	}
}

// InternalError creates an internal error
func InternalError(message string, err error) error {
	return &AppError{
		Type:    ErrInternal,
		Message: message,
		Err:     err,
	}
}

// Duplicate creates a duplicate error
func Duplicate(resource string, id string) error {
	return &AppError{
		Type:    ErrDuplicate,
		Message: fmt.Sprintf("%s already exists: %s", resource, id),
	}
}

// Permission creates a permission error
func Permission(action string, resource string) error {
	return &AppError{
		Type:    ErrPermission,
		Message: fmt.Sprintf("permission denied: cannot %s %s", action, resource),
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Type == ErrNotFound
}

// IsInvalidInput checks if an error is an invalid input error
func IsInvalidInput(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Type == ErrInvalidInput
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Type == ErrInternal
}

// IsDuplicate checks if an error is a duplicate error
func IsDuplicate(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Type == ErrDuplicate
}

// IsPermission checks if an error is a permission error
func IsPermission(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Type == ErrPermission
}

// ValidationError creates a validation error
func ValidationError(field string, message string) error {
	return &AppError{
		Type:    ErrValidation,
		Message: fmt.Sprintf("validation error for %s: %s", field, message),
	}
}

// ConflictError creates a conflict error
func ConflictError(resource string, message string) error {
	return &AppError{
		Type:    ErrConflict,
		Message: fmt.Sprintf("conflict with %s: %s", resource, message),
	}
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Type == ErrValidation
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	e, ok := err.(*AppError)
	return ok && e.Type == ErrConflict
}