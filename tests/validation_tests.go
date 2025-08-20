package tests

import (
	"context"
	"fmt"

	"terralense-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// ValidationTests contains tests for input validation
type ValidationTests struct {
	*BaseTestSuite
}

// NewValidationTests creates a new validation test suite
func NewValidationTests(client *registry.Client, logger *logrus.Logger) TestSuite {
	suite := &ValidationTests{
		BaseTestSuite: NewBaseTestSuite("Validation", client, logger),
	}

	suite.setupTests()
	return suite
}

func (s *ValidationTests) setupTests() {
	s.AddTest("Module Parameters", "Test module parameter validation", s.testModuleParameters)
	s.AddTest("Provider Parameters", "Test provider parameter validation", s.testProviderParameters)
	s.AddTest("Policy Parameters", "Test policy parameter validation", s.testPolicyParameters)
	s.AddTest("Version Validation", "Test version string validation", s.testVersionValidation)
	s.AddTest("Pagination Limits", "Test pagination parameter limits", s.testPaginationLimits)
	s.AddTest("Module ID Format", "Test module ID parsing", s.testModuleIDFormat)
	s.AddTest("Policy ID Format", "Test policy ID parsing", s.testPolicyIDFormat)
	s.AddTest("Provider URI Format", "Test provider URI parsing", s.testProviderURIFormat)
}

func (s *ValidationTests) testModuleParameters(ctx context.Context) error {
	// Test invalid namespace
	_, err := s.client.Modules.Get(ctx, "", "name", "provider", "1.0.0")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty namespace, got: %v", err)
	}

	// Test invalid name
	_, err = s.client.Modules.Get(ctx, "namespace", "", "provider", "1.0.0")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty name, got: %v", err)
	}

	// Test invalid provider
	_, err = s.client.Modules.Get(ctx, "namespace", "name", "", "1.0.0")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty provider, got: %v", err)
	}

	// Test invalid version format
	_, err = s.client.Modules.Get(ctx, "namespace", "name", "provider", "invalid-version")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for invalid version, got: %v", err)
	}

	// Test with special characters in namespace
	_, err = s.client.Modules.Get(ctx, "name@space", "name", "provider", "1.0.0")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for invalid namespace characters, got: %v", err)
	}

	// Test with uppercase in provider (should be lowercase)
	_, err = s.client.Modules.Get(ctx, "namespace", "name", "AWS", "1.0.0")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for uppercase provider, got: %v", err)
	}

	s.logger.Debug("Module parameter validation working correctly")
	return nil
}

func (s *ValidationTests) testProviderParameters(ctx context.Context) error {
	// Test empty namespace
	_, err := s.client.Providers.Get(ctx, "", "aws")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty namespace, got: %v", err)
	}

	// Test empty name
	_, err = s.client.Providers.Get(ctx, "hashicorp", "")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty name, got: %v", err)
	}

	// Test invalid characters
	_, err = s.client.Providers.Get(ctx, "hash!corp", "aws")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for invalid namespace characters, got: %v", err)
	}

	// Test uppercase in provider name (should be lowercase)
	_, err = s.client.Providers.Get(ctx, "hashicorp", "AWS")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for uppercase provider name, got: %v", err)
	}

	s.logger.Debug("Provider parameter validation working correctly")
	return nil
}

func (s *ValidationTests) testPolicyParameters(ctx context.Context) error {
	// Test empty namespace
	_, err := s.client.Policies.Get(ctx, "", "policy", "1.0.0")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty namespace, got: %v", err)
	}

	// Test empty name
	_, err = s.client.Policies.Get(ctx, "namespace", "", "1.0.0")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty name, got: %v", err)
	}

	// Test empty version
	_, err = s.client.Policies.Get(ctx, "namespace", "policy", "")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty version, got: %v", err)
	}

	// Test invalid version format
	_, err = s.client.Policies.Get(ctx, "namespace", "policy", "not-a-version")
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for invalid version format, got: %v", err)
	}

	s.logger.Debug("Policy parameter validation working correctly")
	return nil
}

func (s *ValidationTests) testVersionValidation(ctx context.Context) error {
	validVersions := []string{
		"1.0.0",
		"0.1.0",
		"10.20.30",
		"v1.0.0",
		"v2.5.0",
		"1.0.0-alpha",
		"1.0.0-beta.1",
		"1.0.0-rc.1",
		"latest", // Special case
		"",       // Empty means latest
	}

	invalidVersions := []string{
		"1",
		"1.0",
		"1.0.0.0",
		"abc",
		"1.a.0",
		"1.0.a",
		"v1.0.0.0",
		"1.0.0-",
	}

	// Test valid versions
	for _, version := range validVersions {
		err := registry.ValidateProviderVersion(version)
		if err != nil {
			return fmt.Errorf("expected valid version '%s' to pass validation: %v", version, err)
		}
	}

	// Test invalid versions
	for _, version := range invalidVersions {
		err := registry.ValidateProviderVersion(version)
		if err == nil {
			return fmt.Errorf("expected invalid version '%s' to fail validation", version)
		}
	}

	s.logger.Debug("Version validation working correctly")
	return nil
}

func (s *ValidationTests) testPaginationLimits(ctx context.Context) error {
	// Test negative offset
	opts := &registry.ModuleListOptions{
		Offset: -1,
		Limit:  10,
	}

	_, err := s.client.Modules.List(ctx, opts)
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for negative offset, got: %v", err)
	}

	// Test negative limit
	opts = &registry.ModuleListOptions{
		Offset: 0,
		Limit:  -1,
	}

	_, err = s.client.Modules.List(ctx, opts)
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for negative limit, got: %v", err)
	}

	// Test limit over maximum
	opts = &registry.ModuleListOptions{
		Offset: 0,
		Limit:  200, // Max is typically 100
	}

	_, err = s.client.Modules.List(ctx, opts)
	if err == nil || !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for limit over maximum, got: %v", err)
	}

	// Test valid pagination
	opts = &registry.ModuleListOptions{
		Offset: 0,
		Limit:  50,
	}

	_, err = s.client.Modules.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("valid pagination parameters failed: %v", err)
	}

	s.logger.Debug("Pagination limit validation working correctly")
	return nil
}

func (s *ValidationTests) testModuleIDFormat(ctx context.Context) error {
	testCases := []struct {
		moduleID    string
		expectError bool
		errorType   string
	}{
		{"namespace/name/provider/1.0.0", false, ""},
		{"namespace/name/provider/v1.0.0", false, ""},
		{"namespace/name/provider", true, "validation"},
		{"namespace/name", true, "validation"},
		{"namespace", true, "validation"},
		{"", true, "validation"},
		{"namespace//provider/1.0.0", true, "validation"},
		{"namespace/name/provider/invalid-version", true, "validation"},
		{"name@space/name/provider/1.0.0", true, "validation"},
	}

	for _, tc := range testCases {
		namespace, name, provider, version, err := registry.ParseModuleID(tc.moduleID)

		if tc.expectError {
			if err == nil {
				return fmt.Errorf("expected error for module ID '%s', got nil", tc.moduleID)
			}
			s.logger.Debugf("Module ID '%s' correctly rejected: %v", tc.moduleID, err)
		} else {
			if err != nil {
				return fmt.Errorf("unexpected error for valid module ID '%s': %v", tc.moduleID, err)
			}

			// Verify parsed components
			if namespace == "" || name == "" || provider == "" || version == "" {
				return fmt.Errorf("parsed empty component from module ID '%s'", tc.moduleID)
			}

			s.logger.Debugf("Module ID '%s' parsed: %s/%s/%s@%s",
				tc.moduleID, namespace, name, provider, version)
		}
	}

	return nil
}

func (s *ValidationTests) testPolicyIDFormat(ctx context.Context) error {
	testCases := []struct {
		policyID    string
		expectError bool
	}{
		{"namespace/name/1.0.0", false},
		{"policies/namespace/name/1.0.0", false},
		{"namespace/name/v1.0.0", false},
		{"namespace/name", true},
		{"namespace", true},
		{"", true},
		{"namespace//1.0.0", true},
		{"namespace/name/invalid-version", true},
	}

	for _, tc := range testCases {
		namespace, name, version, err := registry.ParsePolicyID(tc.policyID)

		if tc.expectError {
			if err == nil {
				return fmt.Errorf("expected error for policy ID '%s', got nil", tc.policyID)
			}
			s.logger.Debugf("Policy ID '%s' correctly rejected: %v", tc.policyID, err)
		} else {
			if err != nil {
				return fmt.Errorf("unexpected error for valid policy ID '%s': %v", tc.policyID, err)
			}

			// Verify parsed components
			if namespace == "" || name == "" || version == "" {
				return fmt.Errorf("parsed empty component from policy ID '%s'", tc.policyID)
			}

			s.logger.Debugf("Policy ID '%s' parsed: %s/%s@%s",
				tc.policyID, namespace, name, version)
		}
	}

	return nil
}

func (s *ValidationTests) testProviderURIFormat(ctx context.Context) error {
	testCases := []struct {
		uri         string
		expectError bool
		expected    struct {
			namespace string
			name      string
			version   string
		}
	}{
		{
			uri:         "hashicorp/aws",
			expectError: false,
			expected:    struct{ namespace, name, version string }{"hashicorp", "aws", ""},
		},
		{
			uri:         "hashicorp/aws/4.0.0",
			expectError: false,
			expected:    struct{ namespace, name, version string }{"hashicorp", "aws", "4.0.0"},
		},
		{
			uri:         "registry://hashicorp/aws",
			expectError: false,
			expected:    struct{ namespace, name, version string }{"hashicorp", "aws", ""},
		},
		{
			uri:         "providers/hashicorp/aws/4.0.0",
			expectError: false,
			expected:    struct{ namespace, name, version string }{"hashicorp", "aws", "4.0.0"},
		},
		{
			uri:         "hashicorp",
			expectError: true,
			expected:    struct{ namespace, name, version string }{},
		},
		{
			uri:         "",
			expectError: true,
			expected:    struct{ namespace, name, version string }{},
		},
		{
			uri:         "hash!corp/aws",
			expectError: true,
			expected:    struct{ namespace, name, version string }{},
		},
		{
			uri:         "hashicorp/AWS", // Uppercase provider
			expectError: true,
			expected:    struct{ namespace, name, version string }{},
		},
	}

	for _, tc := range testCases {
		namespace, name, version, err := registry.ExtractProviderInfo(tc.uri)

		if tc.expectError {
			if err == nil {
				return fmt.Errorf("expected error for provider URI '%s', got nil", tc.uri)
			}
			s.logger.Debugf("Provider URI '%s' correctly rejected: %v", tc.uri, err)
		} else {
			if err != nil {
				return fmt.Errorf("unexpected error for valid provider URI '%s': %v", tc.uri, err)
			}

			// Verify parsed components
			if namespace != tc.expected.namespace {
				return fmt.Errorf("namespace mismatch for URI '%s': expected '%s', got '%s'",
					tc.uri, tc.expected.namespace, namespace)
			}
			if name != tc.expected.name {
				return fmt.Errorf("name mismatch for URI '%s': expected '%s', got '%s'",
					tc.uri, tc.expected.name, name)
			}
			if version != tc.expected.version {
				return fmt.Errorf("version mismatch for URI '%s': expected '%s', got '%s'",
					tc.uri, tc.expected.version, version)
			}

			s.logger.Debugf("Provider URI '%s' parsed: %s/%s@%s",
				tc.uri, namespace, name, version)
		}
	}

	return nil
}
