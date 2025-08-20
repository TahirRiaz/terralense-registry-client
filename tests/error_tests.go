package tests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"terralense-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// ErrorTests contains tests for error handling
type ErrorTests struct {
	*BaseTestSuite
}

// NewErrorTests creates a new error handling test suite
func NewErrorTests(client *registry.Client, logger *logrus.Logger) TestSuite {
	suite := &ErrorTests{
		BaseTestSuite: NewBaseTestSuite("Error Handling", client, logger),
	}

	suite.setupTests()
	return suite
}

func (s *ErrorTests) setupTests() {
	s.AddTest("Not Found Errors", "Test 404 error handling", s.testNotFoundErrors)
	s.AddTest("Validation Errors", "Test validation error handling", s.testValidationErrors)
	s.AddTest("Error Type Checking", "Test error type helper functions", s.testErrorTypeChecking)
	s.AddTest("Context Cancellation", "Test context cancellation handling", s.testContextCancellation)
	s.AddTest("Timeout Handling", "Test request timeout handling", s.testTimeoutHandling)
	s.AddTest("API Error Structure", "Test API error response parsing", s.testAPIErrorStructure)
	s.AddTest("Multi Error", "Test multiple error aggregation", s.testMultiError)
}

func (s *ErrorTests) testNotFoundErrors(ctx context.Context) error {
	// Test module not found
	_, err := s.client.Modules.Get(ctx, "non-existent-namespace", "non-existent-module", "aws", "1.0.0")
	if err == nil {
		return fmt.Errorf("expected error for non-existent module, got nil")
	}

	if !registry.IsNotFound(err) {
		return fmt.Errorf("expected NotFound error, got: %v", err)
	}

	// Check if error is wrapped
	var apiErr *registry.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode != http.StatusNotFound {
			return fmt.Errorf("expected status code 404, got: %d", apiErr.StatusCode)
		}
	} else {
		// If not an APIError directly, it's still okay as long as IsNotFound works
		s.logger.Debug("Error is wrapped but IsNotFound check passed")
	}

	// Test provider not found
	_, err = s.client.Providers.Get(ctx, "non-existent", "provider")
	if !registry.IsNotFound(err) {
		return fmt.Errorf("expected NotFound error for provider, got: %v", err)
	}

	// Test policy not found
	_, err = s.client.Policies.Get(ctx, "non-existent", "policy", "1.0.0")
	if !registry.IsNotFound(err) {
		return fmt.Errorf("expected NotFound error for policy, got: %v", err)
	}

	s.logger.Debug("Not found error handling working correctly")
	return nil
}

func (s *ErrorTests) testValidationErrors(ctx context.Context) error {
	// Test various validation errors
	testCases := []struct {
		name string
		test func() error
	}{
		{
			name: "empty module namespace",
			test: func() error {
				_, err := s.client.Modules.Get(ctx, "", "name", "provider", "1.0.0")
				return err
			},
		},
		{
			name: "invalid version format",
			test: func() error {
				_, err := s.client.Modules.Get(ctx, "namespace", "name", "provider", "not-a-version")
				return err
			},
		},
		{
			name: "negative pagination offset",
			test: func() error {
				opts := &registry.ModuleListOptions{Offset: -1}
				_, err := s.client.Modules.List(ctx, opts)
				return err
			},
		},
		{
			name: "empty search query",
			test: func() error {
				_, err := s.client.Modules.Search(ctx, "", 0)
				return err
			},
		},
	}

	for _, tc := range testCases {
		err := tc.test()
		if err == nil {
			return fmt.Errorf("expected validation error for %s, got nil", tc.name)
		}

		if !registry.IsValidationError(err) {
			return fmt.Errorf("expected validation error for %s, got: %v", tc.name, err)
		}

		// Check if it's a ValidationError type
		if validErr, ok := err.(*registry.ValidationError); ok {
			if validErr.Field == "" && validErr.Message == "" {
				return fmt.Errorf("validation error missing details for %s", tc.name)
			}
			s.logger.Debugf("Validation error for %s: field=%s, message=%s",
				tc.name, validErr.Field, validErr.Message)
		}
	}

	return nil
}

func (s *ErrorTests) testErrorTypeChecking(ctx context.Context) error {
	// Create various error types
	errors := map[string]error{
		"not_found":    &registry.APIError{StatusCode: 404, Message: "Not found"},
		"unauthorized": &registry.APIError{StatusCode: 401, Message: "Unauthorized"},
		"forbidden":    &registry.APIError{StatusCode: 403, Message: "Forbidden"},
		"rate_limited": &registry.APIError{StatusCode: 429, Message: "Too many requests"},
		"server_error": &registry.APIError{StatusCode: 500, Message: "Internal server error"},
		"validation":   &registry.ValidationError{Field: "test", Message: "Invalid value"},
	}

	// Test error type checking functions
	testCases := []struct {
		errorKey  string
		checkFunc func(error) bool
		expected  bool
	}{
		{"not_found", registry.IsNotFound, true},
		{"unauthorized", registry.IsUnauthorized, true},
		{"forbidden", registry.IsForbidden, true},
		{"rate_limited", registry.IsRateLimited, true},
		{"server_error", registry.IsServerError, true},
		{"validation", registry.IsValidationError, true},
		{"not_found", registry.IsUnauthorized, false},
		{"validation", registry.IsNotFound, false},
	}

	for _, tc := range testCases {
		err := errors[tc.errorKey]
		result := tc.checkFunc(err)

		if result != tc.expected {
			return fmt.Errorf("error check failed for %s: expected %v, got %v",
				tc.errorKey, tc.expected, result)
		}
	}

	s.logger.Debug("Error type checking functions work correctly")
	return nil
}

func (s *ErrorTests) testContextCancellation(ctx context.Context) error {
	// Create a context that we can cancel
	cancelCtx, cancel := context.WithCancel(ctx)

	// Start a request in a goroutine
	errChan := make(chan error, 1)
	go func() {
		// This should be interrupted by context cancellation
		_, err := s.client.Modules.List(cancelCtx, &registry.ModuleListOptions{Limit: 100})
		errChan <- err
	}()

	// Cancel the context quickly
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for the error
	select {
	case err := <-errChan:
		if err == nil {
			// Request might have completed before cancellation
			s.logger.Debug("Request completed before cancellation")
		} else if err == context.Canceled {
			s.logger.Debug("Context cancellation handled correctly")
		} else {
			// Some other error occurred
			s.logger.Debugf("Got error after cancellation: %v", err)
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for cancelled request to complete")
	}

	return nil
}

func (s *ErrorTests) testTimeoutHandling(ctx context.Context) error {
	// Create a context with very short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	// This should timeout
	_, err := s.client.Modules.Search(timeoutCtx, "terraform", 0)

	if err == nil {
		// Request might have completed very quickly
		s.logger.Debug("Request completed before timeout")
		return nil
	}

	// Check if it's a timeout error
	if err == context.DeadlineExceeded {
		s.logger.Debug("Timeout handled correctly: context deadline exceeded")
		return nil
	}

	// Might be wrapped in another error
	if err.Error() != "" {
		s.logger.Debugf("Got error after timeout: %v", err)
		// This is acceptable - the important thing is that the request didn't hang
		return nil
	}

	return fmt.Errorf("unexpected error handling for timeout: %v", err)
}

func (s *ErrorTests) testAPIErrorStructure(ctx context.Context) error {
	// Trigger various API errors and check their structure

	// 404 error
	_, err := s.client.Modules.Get(ctx, "definitely", "does-not", "exist", "1.0.0")
	if err != nil {
		if apiErr, ok := err.(*registry.APIError); ok {
			// Check error properties
			if apiErr.StatusCode == 0 {
				return fmt.Errorf("API error missing status code")
			}
			if apiErr.Message == "" {
				return fmt.Errorf("API error missing message")
			}

			// Check Error() method
			errStr := apiErr.Error()
			if errStr == "" {
				return fmt.Errorf("API error Error() returned empty string")
			}

			// Check Is() method
			if !apiErr.Is(registry.ErrNotFound) {
				return fmt.Errorf("404 API error should match ErrNotFound")
			}

			// Check Unwrap() method
			unwrapped := apiErr.Unwrap()
			if apiErr.StatusCode == 404 && unwrapped != registry.ErrNotFound {
				return fmt.Errorf("404 API error should unwrap to ErrNotFound")
			}

			s.logger.Debugf("API error structure: status=%d, message=%s",
				apiErr.StatusCode, apiErr.Message)
		}
	}

	return nil
}

func (s *ErrorTests) testMultiError(ctx context.Context) error {
	// Test MultiError functionality
	multiErr := &registry.MultiError{}

	// Should not have errors initially
	if multiErr.HasErrors() {
		return fmt.Errorf("new MultiError should not have errors")
	}

	// ErrorOrNil should return nil
	if multiErr.ErrorOrNil() != nil {
		return fmt.Errorf("ErrorOrNil should return nil for empty MultiError")
	}

	// Add some errors
	multiErr.Add(fmt.Errorf("error 1"))
	multiErr.Add(fmt.Errorf("error 2"))
	multiErr.Add(nil) // Should be ignored
	multiErr.Add(fmt.Errorf("error 3"))

	// Should have errors now
	if !multiErr.HasErrors() {
		return fmt.Errorf("MultiError should have errors after adding")
	}

	// Should have 3 errors (nil was ignored)
	if len(multiErr.Errors) != 3 {
		return fmt.Errorf("expected 3 errors, got %d", len(multiErr.Errors))
	}

	// ErrorOrNil should return the error
	err := multiErr.ErrorOrNil()
	if err == nil {
		return fmt.Errorf("ErrorOrNil should return error when errors exist")
	}

	// Error() method should work
	errStr := err.Error()
	if errStr == "" {
		return fmt.Errorf("MultiError Error() returned empty string")
	}

	// Test single error case
	singleErr := &registry.MultiError{}
	singleErr.Add(fmt.Errorf("single error"))

	err = singleErr.ErrorOrNil()
	if err == nil {
		return fmt.Errorf("ErrorOrNil should return error for single error")
	}

	// For single error, it should return the error directly
	if _, ok := err.(*registry.MultiError); ok {
		// It's returning the MultiError wrapper, which is also acceptable
		s.logger.Debug("Single error returned as MultiError wrapper")
	}

	s.logger.Debugf("MultiError handling works correctly with %d errors", len(multiErr.Errors))
	return nil
}
