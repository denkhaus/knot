package sqlite

import (
	"errors"
	"fmt"
)

// RepositoryError represents a repository operation error
type RepositoryError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// ErrorType represents different types of repository errors
type ErrorType int

const (
	ErrorTypeNotFound ErrorType = iota
	ErrorTypeConstraintViolation
	ErrorTypeCircularDependency
	ErrorTypeMaxDepthExceeded
	ErrorTypeMaxTasksExceeded
	ErrorTypeConnectionError
	ErrorTypeTransactionError
	ErrorTypeMigrationError
	ErrorTypeValidationError
)

// Error returns the error message
func (e *RepositoryError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *RepositoryError) Unwrap() error {
	return e.Cause
}

// Is checks if the error is of a specific type
func (e *RepositoryError) Is(target error) bool {
	if re, ok := target.(*RepositoryError); ok {
		return e.Type == re.Type
	}
	return false
}

// Helper functions for creating specific errors

// NewNotFoundError creates a new not found error
func NewNotFoundError(entity string, id string) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s with ID %s not found", entity, id),
	}
}

// NewConstraintViolationError creates a new constraint violation error
func NewConstraintViolationError(constraint string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeConstraintViolation,
		Message: fmt.Sprintf("constraint violation: %s", constraint),
		Cause:   cause,
	}
}

// NewCircularDependencyError creates a new circular dependency error
func NewCircularDependencyError(message string) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeCircularDependency,
		Message: message,
	}
}

// NewMaxDepthExceededError creates a new max depth exceeded error
func NewMaxDepthExceededError(maxDepth int) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeMaxDepthExceeded,
		Message: fmt.Sprintf("maximum task depth of %d exceeded", maxDepth),
	}
}

// NewMaxTasksExceededError creates a new max tasks exceeded error
func NewMaxTasksExceededError(maxTasks int) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeMaxTasksExceeded,
		Message: fmt.Sprintf("maximum number of tasks (%d) exceeded", maxTasks),
	}
}

// NewConnectionError creates a new connection error
func NewConnectionError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeConnectionError,
		Message: message,
		Cause:   cause,
	}
}

// NewTransactionError creates a new transaction error
func NewTransactionError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeTransactionError,
		Message: message,
		Cause:   cause,
	}
}

// NewMigrationError creates a new migration error
func NewMigrationError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeMigrationError,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, cause error) *RepositoryError {
	return &RepositoryError{
		Type:    ErrorTypeValidationError,
		Message: message,
		Cause:   cause,
	}
}

// Error type check helpers

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeNotFound
}

// IsConstraintViolationError checks if an error is a constraint violation error
func IsConstraintViolationError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeConstraintViolation
}

// IsCircularDependencyError checks if an error is a circular dependency error
func IsCircularDependencyError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeCircularDependency
}

// IsMaxDepthExceededError checks if an error is a max depth exceeded error
func IsMaxDepthExceededError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeMaxDepthExceeded
}

// IsMaxTasksExceededError checks if an error is a max tasks exceeded error
func IsMaxTasksExceededError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeMaxTasksExceeded
}

// IsConnectionError checks if an error is a connection error
func IsConnectionError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeConnectionError
}

// IsTransactionError checks if an error is a transaction error
func IsTransactionError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeTransactionError
}

// IsMigrationError checks if an error is a migration error
func IsMigrationError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeMigrationError
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var re *RepositoryError
	return errors.As(err, &re) && re.Type == ErrorTypeValidationError
}
