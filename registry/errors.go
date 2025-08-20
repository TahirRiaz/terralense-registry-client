package registry

import (
	"errors"
	"fmt"
	"net/http"
)

// Common errors
var (
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when access is forbidden
	ErrForbidden = errors.New("forbidden")

	// ErrRateLimited is returned when rate limit is exceeded
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")

	// ErrTimeout is returned when a request times out
	ErrTimeout = errors.New("request timeout")

	// ErrServerError is returned for server-side errors
	ErrServerError = errors.New("server error")
)

// APIError represents an error returned by the Terraform Registry API
type APIError struct {
	StatusCode int         `json:"-"`
	Message    string      `json:"message"`
	Code       string      `json:"code,omitempty"`
	Details    interface{} `json:"details,omitempty"`
	Headers    http.Header `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error (status %d, code %s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// Is implements error matching
func (e *APIError) Is(target error) bool {
	switch e.StatusCode {
	case http.StatusNotFound:
		return target == ErrNotFound
	case http.StatusUnauthorized:
		return target == ErrUnauthorized
	case http.StatusForbidden:
		return target == ErrForbidden
	case http.StatusTooManyRequests:
		return target == ErrRateLimited
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return target == ErrServerError
	}
	return false
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return ErrServerError
	}
	return nil
}

// RequestError represents an error that occurred while making a request
type RequestError struct {
	Method string
	URL    string
	Err    error
}

// Error implements the error interface
func (e *RequestError) Error() string {
	return fmt.Sprintf("request error (%s %s): %v", e.Method, e.URL, e.Err)
}

// Unwrap returns the underlying error
func (e *RequestError) Unwrap() error {
	return e.Err
}

// ResponseError represents an error that occurred while processing a response
type ResponseError struct {
	StatusCode int
	Err        error
}

// Error implements the error interface
func (e *ResponseError) Error() string {
	return fmt.Sprintf("response error (status %d): %v", e.StatusCode, e.Err)
}

// Unwrap returns the underlying error
func (e *ResponseError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// Is implements error matching
func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// MultiError represents multiple errors
type MultiError struct {
	Errors []error
}

// Error implements the error interface
func (e *MultiError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors occurred (%d errors)", len(e.Errors))
}

// Add adds an error to the multi-error
func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// HasErrors returns true if there are any errors
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// ErrorOrNil returns nil if there are no errors, otherwise returns the multi-error
func (e *MultiError) ErrorOrNil() error {
	if !e.HasErrors() {
		return nil
	}
	if len(e.Errors) == 1 {
		return e.Errors[0]
	}
	return e
}

// IsNotFound returns true if the error is a 404 Not Found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsRateLimited returns true if the error is a 429 Too Many Requests error
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden returns true if the error is a 403 Forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsServerError returns true if the error is a 5xx server error
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerError)
}

// IsTimeout returns true if the error is a timeout error
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsValidationError returns true if the error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}
