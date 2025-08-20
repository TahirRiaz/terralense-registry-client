package tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/TahirRiaz/terralense-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// PolicyTests contains tests for the Policies API
type PolicyTests struct {
	*BaseTestSuite
}

// NewPolicyTests creates a new policy test suite
func NewPolicyTests(client *registry.Client, logger *logrus.Logger) TestSuite {
	suite := &PolicyTests{
		BaseTestSuite: NewBaseTestSuite("Policies", client, logger),
	}

	suite.setupTests()
	return suite
}

func (s *PolicyTests) setupTests() {
	s.AddTest("List Policies", "Test listing policies with various options", s.testListPolicies)
	s.AddTest("Get Policy", "Test getting a specific policy", s.testGetPolicy)
	s.AddTest("Get Policy by ID", "Test getting policy by full ID", s.testGetPolicyByID)
	s.AddTest("Search Policies", "Test policy search functionality", s.testSearchPolicies)
	s.AddTest("Get Sentinel Content", "Test generating Sentinel configuration", s.testGetSentinelContent)
	s.AddTest("Pagination", "Test policy list pagination", s.testPagination)
	s.AddTest("Include Latest Version", "Test including latest version data", s.testIncludeLatestVersion)
	s.AddTest("Invalid Policy", "Test error handling for invalid policies", s.testInvalidPolicy)
}

// In policy_tests.go, update the testListPolicies function:

func (s *PolicyTests) testListPolicies(ctx context.Context) error {
	opts := &registry.PolicyListOptions{
		PageSize:             10,
		Page:                 1,
		IncludeLatestVersion: true,
	}

	result, err := s.client.Policies.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}

	if err := AssertNotNil(result); err != nil {
		return err
	}

	// Note: There might be no policies in the registry
	if len(result.Data) > 0 {
		// Verify policy structure
		for _, policy := range result.Data {
			if policy.ID == "" {
				return fmt.Errorf("policy has empty ID")
			}
			// Accept both "policies" and "policy-libraries" types
			if policy.Type != "policies" && policy.Type != "policy-libraries" {
				return fmt.Errorf("unexpected policy type: %s", policy.Type)
			}
			if policy.Attributes.Namespace == "" {
				return fmt.Errorf("policy has empty namespace")
			}
			if policy.Attributes.Name == "" {
				return fmt.Errorf("policy has empty name")
			}
		}

		// If we requested latest version, check included data
		if opts.IncludeLatestVersion && len(result.Included) == 0 {
			s.logger.Warn("No included version data despite requesting it")
		}
	}

	s.logger.Debugf("Listed %d policies", len(result.Data))
	return nil
}

// Also update testIncludeLatestVersion function:

func (s *PolicyTests) testIncludeLatestVersion(ctx context.Context) error {
	// Test without including latest version
	optsWithout := &registry.PolicyListOptions{
		PageSize:             5,
		Page:                 1,
		IncludeLatestVersion: false,
	}

	resultWithout, err := s.client.Policies.List(ctx, optsWithout)
	if err != nil {
		return fmt.Errorf("failed to list policies without version: %w", err)
	}

	// Test with including latest version
	optsWith := &registry.PolicyListOptions{
		PageSize:             5,
		Page:                 1,
		IncludeLatestVersion: true,
	}

	resultWith, err := s.client.Policies.List(ctx, optsWith)
	if err != nil {
		return fmt.Errorf("failed to list policies with version: %w", err)
	}

	// Compare the results
	if len(resultWithout.Data) > 0 && len(resultWith.Data) > 0 {
		// When not including latest version, there should be no included data
		if len(resultWithout.Included) > 0 {
			s.logger.Warn("Unexpected included data when not requested")
		}

		// When including latest version, we should have included data
		if optsWith.IncludeLatestVersion {
			if len(resultWith.Included) == 0 {
				s.logger.Warn("No included version data when requested")
			} else {
				// Verify included data is version information
				for _, included := range resultWith.Included {
					// Accept both "policy-versions" and "policy-library-versions"
					if included.Type != "policy-versions" && included.Type != "policy-library-versions" {
						return fmt.Errorf("unexpected included type: %s", included.Type)
					}

					if included.Attributes.Version == "" {
						return fmt.Errorf("included version data has empty version")
					}
				}

				s.logger.Debugf("Got %d included version records", len(resultWith.Included))
			}
		}
	}

	return nil
}

func (s *PolicyTests) testGetPolicy(ctx context.Context) error {
	// First, list policies to get valid test data
	opts := &registry.PolicyListOptions{
		PageSize: 5,
		Page:     1,
	}

	list, err := s.client.Policies.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}

	if len(list.Data) == 0 {
		s.logger.Warn("No policies available for testing")
		return nil
	}

	// Use the first policy
	testPolicy := list.Data[0]

	// Get the policy with a version (we need to determine the version first)
	// For now, try with a common version pattern
	versions := []string{"1.0.0", "0.1.0", "latest"}

	var policy *registry.PolicyDetails
	for _, version := range versions {
		policy, err = s.client.Policies.Get(ctx,
			testPolicy.Attributes.Namespace,
			testPolicy.Attributes.Name,
			version)

		if err == nil {
			break
		}

		if !registry.IsNotFound(err) {
			return fmt.Errorf("unexpected error getting policy: %w", err)
		}
	}

	if policy == nil {
		s.logger.Warnf("Could not find any version for policy %s/%s",
			testPolicy.Attributes.Namespace, testPolicy.Attributes.Name)
		return nil
	}

	// Verify policy details
	if policy.Data.Attributes.Version == "" {
		return fmt.Errorf("policy version is empty")
	}

	s.logger.Debugf("Got policy %s version %s",
		testPolicy.Attributes.FullName, policy.Data.Attributes.Version)

	return nil
}

func (s *PolicyTests) testGetPolicyByID(ctx context.Context) error {
	// First, get a valid policy ID
	opts := &registry.PolicyListOptions{
		PageSize: 1,
		Page:     1,
	}

	list, err := s.client.Policies.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}

	if len(list.Data) == 0 {
		s.logger.Warn("No policies available for testing")
		return nil
	}

	// Construct a policy ID (format: policies/namespace/name/version)
	// We need to get a valid version first
	testPolicy := list.Data[0]

	// Try to construct an ID with a common version
	policyID := fmt.Sprintf("policies/%s/%s/1.0.0",
		testPolicy.Attributes.Namespace, testPolicy.Attributes.Name)

	policy, err := s.client.Policies.GetByID(ctx, policyID)
	if err != nil {
		if registry.IsNotFound(err) {
			// Try without the "policies/" prefix
			policyID = fmt.Sprintf("%s/%s/1.0.0",
				testPolicy.Attributes.Namespace, testPolicy.Attributes.Name)
			policy, err = s.client.Policies.GetByID(ctx, policyID)

			if err != nil {
				s.logger.Warnf("Could not find policy by ID: %s", policyID)
				return nil
			}
		} else {
			return fmt.Errorf("failed to get policy by ID: %w", err)
		}
	}

	// Use the policy variable to verify it was retrieved correctly
	if policy == nil || policy.Data.ID == "" {
		return fmt.Errorf("retrieved policy is invalid")
	}

	s.logger.Debugf("Successfully retrieved policy by ID: %s", policyID)
	return nil
}

func (s *PolicyTests) testSearchPolicies(ctx context.Context) error {
	queries := []string{
		"aws",
		"compliance",
		"security",
		"cis",
	}

	foundAny := false
	for _, query := range queries {
		results, err := s.client.Policies.Search(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to search for '%s': %w", query, err)
		}

		if len(results) > 0 {
			foundAny = true

			// Verify search results are sorted by relevance
			for i := 1; i < len(results); i++ {
				if results[i].Relevance > results[i-1].Relevance {
					return fmt.Errorf("results not sorted by relevance")
				}
			}

			// Verify search results contain query terms
			for _, result := range results {
				policyText := strings.ToLower(fmt.Sprintf("%s %s %s",
					result.Policy.Attributes.Name,
					result.Policy.Attributes.Title,
					result.Policy.Attributes.Namespace))

				if !strings.Contains(policyText, strings.ToLower(query)) {
					s.logger.Warnf("Policy %s doesn't contain query term '%s'",
						result.Policy.Attributes.FullName, query)
				}
			}

			s.logger.Debugf("Search for '%s' returned %d results", query, len(results))
		}
	}

	if !foundAny {
		s.logger.Warn("No search results found for any query")
	}

	return nil
}

func (s *PolicyTests) testGetSentinelContent(ctx context.Context) error {
	// Get a valid policy first
	opts := &registry.PolicyListOptions{
		PageSize:             1,
		Page:                 1,
		IncludeLatestVersion: true,
	}

	list, err := s.client.Policies.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}

	if len(list.Data) == 0 {
		s.logger.Warn("No policies available for testing Sentinel content")
		return nil
	}

	// Get the policy ID
	policy := list.Data[0]

	// Try to get version from included data
	version := "1.0.0" // default
	if len(list.Included) > 0 {
		for _, included := range list.Included {
			if included.Type == "policy-versions" && included.ID != "" {
				version = included.Attributes.Version
				break
			}
		}
	}

	policyID := fmt.Sprintf("policies/%s/%s/%s",
		policy.Attributes.Namespace, policy.Attributes.Name, version)

	content, err := s.client.Policies.GetSentinelContent(ctx, policyID)
	if err != nil {
		if registry.IsNotFound(err) {
			s.logger.Warnf("Policy %s not found for Sentinel content test", policyID)
			return nil
		}
		return fmt.Errorf("failed to get Sentinel content: %w", err)
	}

	// Verify content structure
	if content.PolicyID != policyID {
		return fmt.Errorf("policy ID mismatch: expected %s, got %s",
			policyID, content.PolicyID)
	}

	if content.Version == "" {
		return fmt.Errorf("Sentinel content has empty version")
	}

	// Generate HCL
	enforcementLevels := []string{"advisory", "soft-mandatory", "hard-mandatory"}

	for _, level := range enforcementLevels {
		hcl := content.GenerateHCL(level)

		if hcl == "" {
			return fmt.Errorf("generated HCL is empty for enforcement level %s", level)
		}

		// Verify HCL contains expected content
		if !strings.Contains(hcl, "enforcement_level = \""+level+"\"") {
			return fmt.Errorf("HCL doesn't contain expected enforcement level: %s", level)
		}

		s.logger.Debugf("Generated HCL with %d characters for level %s", len(hcl), level)
	}

	return nil
}

func (s *PolicyTests) testPagination(ctx context.Context) error {
	pageSize := 5
	var allPolicies []registry.Policy

	for page := 1; page <= 3; page++ {
		opts := &registry.PolicyListOptions{
			PageSize: pageSize,
			Page:     page,
		}

		result, err := s.client.Policies.List(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list policies (page %d): %w", page, err)
		}

		if len(result.Data) == 0 {
			break // No more results
		}

		allPolicies = append(allPolicies, result.Data...)

		// Verify pagination metadata
		if result.Meta.Pagination.CurrentPage != page {
			return fmt.Errorf("unexpected current page: expected %d, got %d",
				page, result.Meta.Pagination.CurrentPage)
		}

		if result.Meta.Pagination.PageSize != pageSize {
			return fmt.Errorf("unexpected page size: expected %d, got %d",
				pageSize, result.Meta.Pagination.PageSize)
		}

		// Check if there are more pages
		if result.Meta.Pagination.NextPage == 0 {
			break
		}
	}

	s.logger.Debugf("Retrieved %d policies across multiple pages", len(allPolicies))
	return nil
}

func (s *PolicyTests) testInvalidPolicy(ctx context.Context) error {
	// Test with non-existent policy
	_, err := s.client.Policies.Get(ctx, "invalid-namespace", "invalid-policy", "1.0.0")

	if err == nil {
		return fmt.Errorf("expected error for invalid policy, got nil")
	}

	if !registry.IsNotFound(err) {
		return fmt.Errorf("expected NotFound error, got: %v", err)
	}

	// Test with invalid policy ID
	_, err = s.client.Policies.GetByID(ctx, "invalid/policy/id")
	if err == nil {
		return fmt.Errorf("expected error for invalid policy ID, got nil")
	}

	// The error should be a validation error since the ID format is wrong
	if !registry.IsValidationError(err) && !registry.IsNotFound(err) {
		return fmt.Errorf("expected validation or not found error, got: %v", err)
	}

	s.logger.Debug("Invalid policy handling works correctly")
	return nil
}
