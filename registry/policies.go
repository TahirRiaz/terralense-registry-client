package registry

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// PoliciesService handles communication with the policy related
// methods of the Terraform Registry API.
type PoliciesService struct {
	client *Client
}

// PolicyListOptions specifies optional parameters to the List method
type PolicyListOptions struct {
	// PageSize specifies the number of items per page (max 100)
	PageSize int `url:"page[size],omitempty"`

	// Page specifies the page number for pagination
	Page int `url:"page[number],omitempty"`

	// IncludeLatestVersion includes the latest version information
	IncludeLatestVersion bool
}

// Validate validates the policy list options
func (o *PolicyListOptions) Validate() error {
	if o == nil {
		return nil
	}

	if o.PageSize < 0 || o.PageSize > 100 {
		return &ValidationError{
			Field:   "PageSize",
			Value:   o.PageSize,
			Message: "page size must be between 0 and 100",
		}
	}

	if o.Page < 0 {
		return &ValidationError{
			Field:   "Page",
			Value:   o.Page,
			Message: "page cannot be negative",
		}
	}

	return nil
}

// List returns a list of policies
func (s *PoliciesService) List(ctx context.Context, opts *PolicyListOptions) (*PolicyList, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	values := url.Values{}

	if opts != nil {
		if opts.PageSize > 0 {
			values.Add("page[size]", fmt.Sprintf("%d", opts.PageSize))
		} else {
			values.Add("page[size]", "50") // Default page size
		}

		if opts.Page > 0 {
			values.Add("page[number]", fmt.Sprintf("%d", opts.Page))
		}

		if opts.IncludeLatestVersion {
			values.Add("include", "latest-version")
		}
	} else {
		values.Add("page[size]", "50")
		values.Add("include", "latest-version")
	}

	path := fmt.Sprintf("policies?%s", values.Encode())

	var result PolicyList
	if err := s.client.get(ctx, path, "v2", &result); err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	return &result, nil
}

// Get returns details about a specific policy version
func (s *PoliciesService) Get(ctx context.Context, namespace, name, version string) (*PolicyDetails, error) {
	if err := validatePolicyParams(namespace, name, version); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("policies/%s/%s/%s?include=policies,policy-modules,policy-library",
		url.PathEscape(namespace), url.PathEscape(name), url.PathEscape(version))

	var result PolicyDetails
	if err := s.client.get(ctx, path, "v2", &result); err != nil {
		return nil, fmt.Errorf("failed to get policy %s/%s/%s: %w", namespace, name, version, err)
	}

	return &result, nil
}

// GetByID returns details about a policy using its full ID
func (s *PoliciesService) GetByID(ctx context.Context, policyID string) (*PolicyDetails, error) {
	if policyID == "" {
		return nil, &ValidationError{
			Field:   "policyID",
			Value:   policyID,
			Message: "policy ID cannot be empty",
		}
	}

	// Extract namespace, name, and version from ID
	namespace, name, version, err := ParsePolicyID(policyID)
	if err != nil {
		return nil, &ValidationError{
			Field:   "policyID",
			Value:   policyID,
			Message: err.Error(),
		}
	}

	return s.Get(ctx, namespace, name, version)
}

// Search searches for policies based on a query string
func (s *PoliciesService) Search(ctx context.Context, query string) ([]PolicySearchResult, error) {
	if query == "" {
		return nil, &ValidationError{
			Field:   "query",
			Value:   query,
			Message: "search query cannot be empty",
		}
	}

	// Get all policies (pagination handled internally)
	allPolicies := []Policy{}
	page := 1
	maxPages := 100 // Prevent infinite loops

	for pageCount := 0; pageCount < maxPages; pageCount++ {
		opts := &PolicyListOptions{
			PageSize:             100,
			Page:                 page,
			IncludeLatestVersion: true,
		}

		result, err := s.List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to search policies: %w", err)
		}

		allPolicies = append(allPolicies, result.Data...)

		// Check if there are more pages
		if result.Meta.Pagination.NextPage == 0 {
			break
		}

		page = result.Meta.Pagination.NextPage
	}

	// Filter and rank policies based on query
	var searchResults []PolicySearchResult
	queryLower := strings.ToLower(query)
	queryParts := strings.Fields(queryLower)

	for _, policy := range allPolicies {
		// Calculate match score
		matchScore := calculatePolicyMatchScore(policy, queryLower, queryParts)

		if matchScore > 0 {
			searchResult := PolicySearchResult{
				Policy:    policy,
				Relevance: matchScore,
			}
			searchResults = append(searchResults, searchResult)
		}
	}

	// Sort by relevance
	sort.Slice(searchResults, func(i, j int) bool {
		return searchResults[i].Relevance > searchResults[j].Relevance
	})

	return searchResults, nil
}

// calculatePolicyMatchScore calculates the relevance score for a policy
func calculatePolicyMatchScore(policy Policy, queryLower string, queryParts []string) float64 {
	relevance := 0.0

	nameLower := strings.ToLower(policy.Attributes.Name)
	titleLower := strings.ToLower(policy.Attributes.Title)
	namespaceLower := strings.ToLower(policy.Attributes.Namespace)

	// Exact name match (highest weight)
	if nameLower == queryLower {
		relevance += 10.0
	} else if strings.Contains(nameLower, queryLower) {
		relevance += 5.0
	} else {
		// Check if all query parts are in the name
		allPartsInName := true
		for _, part := range queryParts {
			if !strings.Contains(nameLower, part) {
				allPartsInName = false
				break
			}
		}
		if allPartsInName {
			relevance += 3.0
		}
	}

	// Title match
	if strings.Contains(titleLower, queryLower) {
		relevance += 3.0
	} else {
		// Check if all query parts are in the title
		allPartsInTitle := true
		for _, part := range queryParts {
			if !strings.Contains(titleLower, part) {
				allPartsInTitle = false
				break
			}
		}
		if allPartsInTitle {
			relevance += 1.5
		}
	}

	// Namespace match
	if strings.Contains(namespaceLower, queryLower) {
		relevance += 2.0
	}

	// Verification status
	if policy.Attributes.Verified {
		relevance += 2.0
	}

	// Download count (normalized)
	if policy.Attributes.Downloads > 10000 {
		relevance += 3.0
	} else if policy.Attributes.Downloads > 1000 {
		relevance += 2.0
	} else if policy.Attributes.Downloads > 100 {
		relevance += 1.0
	}

	return relevance
}

// PolicySearchResult represents a search result with relevance information
type PolicySearchResult struct {
	Policy
	Relevance float64 // Calculated relevance score
}

// GetSentinelContent generates Sentinel policy content for a policy
func (s *PoliciesService) GetSentinelContent(ctx context.Context, policyID string) (*SentinelPolicyContent, error) {
	details, err := s.GetByID(ctx, policyID)
	if err != nil {
		return nil, err
	}

	content := &SentinelPolicyContent{
		PolicyID:    policyID,
		Description: details.Data.Attributes.Description,
		Version:     details.Data.Attributes.Version,
		Modules:     []SentinelModule{},
		Policies:    []SentinelPolicy{},
	}

	// Extract modules and policies from included data
	for _, included := range details.Included {
		switch included.Type {
		case "policy-modules":
			if included.Attributes.Name == "" || included.Attributes.Shasum == "" {
				s.client.logger.Warnf("Skipping policy module with missing data: %+v", included)
				continue
			}

			module := SentinelModule{
				Name: included.Attributes.Name,
				Source: fmt.Sprintf("https://registry.terraform.io/v2%s/policy-module/%s.sentinel?checksum=sha256:%s",
					policyID, included.Attributes.Name, included.Attributes.Shasum),
			}
			content.Modules = append(content.Modules, module)

		case "policies":
			if included.Attributes.Name == "" || included.Attributes.Shasum == "" {
				s.client.logger.Warnf("Skipping policy with missing data: %+v", included)
				continue
			}

			policy := SentinelPolicy{
				Name:     included.Attributes.Name,
				Checksum: fmt.Sprintf("sha256:%s", included.Attributes.Shasum),
				Source: fmt.Sprintf("https://registry.terraform.io/v2%s/policy/%s.sentinel?checksum=sha256:%s",
					policyID, included.Attributes.Name, included.Attributes.Shasum),
			}
			content.Policies = append(content.Policies, policy)
		}
	}

	return content, nil
}

// SentinelPolicyContent represents the content needed to generate Sentinel policies
type SentinelPolicyContent struct {
	PolicyID    string
	Description string
	Version     string
	Modules     []SentinelModule
	Policies    []SentinelPolicy
}

// SentinelModule represents a Sentinel module
type SentinelModule struct {
	Name   string
	Source string
}

// SentinelPolicy represents a Sentinel policy
type SentinelPolicy struct {
	Name     string
	Checksum string
	Source   string
}

// GenerateHCL generates HCL configuration for the policy
func (c *SentinelPolicyContent) GenerateHCL(enforcementLevel string) string {
	if err := validateEnforcementLevel(enforcementLevel); err != nil {
		// Default to advisory if invalid
		enforcementLevel = "advisory"
	}

	var builder strings.Builder

	// Add header comment
	builder.WriteString(fmt.Sprintf(`# Sentinel Policy Configuration
# Policy: %s
# Version: %s
# Description: %s

`, c.PolicyID, c.Version, c.Description))

	// Add modules
	if len(c.Modules) > 0 {
		builder.WriteString("# Policy Modules\n")
		for _, module := range c.Modules {
			builder.WriteString(fmt.Sprintf(`module "%s" {
  source = "%s"
}

`, module.Name, module.Source))
		}
	}

	// Add policies
	if len(c.Policies) > 0 {
		builder.WriteString("# Policies\n")
		for _, policy := range c.Policies {
			builder.WriteString(fmt.Sprintf(`policy "%s" {
  source            = "%s"
  enforcement_level = "%s"
}

`, policy.Name, policy.Source, enforcementLevel))
		}
	}

	return builder.String()
}

// validatePolicyParams validates policy parameters
func validatePolicyParams(namespace, name, version string) error {
	var errs MultiError

	if namespace == "" {
		errs.Add(&ValidationError{
			Field:   "namespace",
			Value:   namespace,
			Message: "namespace cannot be empty",
		})
	} else if !isValidNamespace(namespace) {
		errs.Add(&ValidationError{
			Field:   "namespace",
			Value:   namespace,
			Message: "invalid namespace format",
		})
	}

	if name == "" {
		errs.Add(&ValidationError{
			Field:   "name",
			Value:   name,
			Message: "name cannot be empty",
		})
	} else if !isValidPolicyName(name) {
		errs.Add(&ValidationError{
			Field:   "name",
			Value:   name,
			Message: "invalid policy name format",
		})
	}

	if version == "" {
		errs.Add(&ValidationError{
			Field:   "version",
			Value:   version,
			Message: "version cannot be empty",
		})
	} else if !isValidVersion(version) {
		errs.Add(&ValidationError{
			Field:   "version",
			Value:   version,
			Message: "invalid version format",
		})
	}

	return errs.ErrorOrNil()
}

// isValidPolicyName validates policy name format
func isValidPolicyName(name string) bool {
	// Policy name should contain only alphanumeric characters, hyphens, and underscores
	for _, r := range name {
		if !isAlphaNumeric(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

// validateEnforcementLevel validates Sentinel enforcement level
func validateEnforcementLevel(level string) error {
	validLevels := []string{"advisory", "soft-mandatory", "hard-mandatory"}
	for _, valid := range validLevels {
		if level == valid {
			return nil
		}
	}
	return &ValidationError{
		Field:   "enforcementLevel",
		Value:   level,
		Message: fmt.Sprintf("invalid enforcement level, must be one of: %s", strings.Join(validLevels, ", ")),
	}
}
