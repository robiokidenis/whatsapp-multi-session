package models

import "fmt"

// Custom error types for proper HTTP status code handling

// NotFoundError represents a 404 error
type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

// UnauthorizedError represents a 401 error
type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

// BadRequestError represents a 400 error
type BadRequestError struct {
	Message string
}

func (e BadRequestError) Error() string {
	return e.Message
}

// ServiceUnavailableError represents a 503 error
type ServiceUnavailableError struct {
	Message string
}

func (e ServiceUnavailableError) Error() string {
	return e.Message
}

// Helper functions to create errors

func NewNotFoundError(format string, args ...interface{}) error {
	return NotFoundError{Message: fmt.Sprintf(format, args...)}
}

func NewUnauthorizedError(format string, args ...interface{}) error {
	return UnauthorizedError{Message: fmt.Sprintf(format, args...)}
}

func NewBadRequestError(format string, args ...interface{}) error {
	return BadRequestError{Message: fmt.Sprintf(format, args...)}
}

func NewServiceUnavailableError(format string, args ...interface{}) error {
	return ServiceUnavailableError{Message: fmt.Sprintf(format, args...)}
}