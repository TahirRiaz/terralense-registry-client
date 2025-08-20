package registry

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// ModulesService handles communication with the module related
// methods of the Terraform Registry API.
type ModulesService struct {
	client *Client
}

// ModuleListOptions specifies optional parameters to module list methods
type ModuleListOptions struct {
	// Offset specifies the offset for pagination
	Offset int `url:"offset,omitempty"`

	// Limit specifies the number of items to return (max 100)
	Limit int `url:"limit,omitempty"`

	// Provider filters modules by provider
	Provider string `url:"provider,omitempty"`

	// Verified filters to only show verified modules
	Verified bool `url:"verified,omitempty"`
}

// Validate validates the module list options
func (o *ModuleListOptions) Validate() error {
	if o == nil {
		return nil
	}

	if o.Offset < 0 {
		return &ValidationError{
			Field:   "Offset",
			Value:   o.Offset,
			Message: "offset cannot be negative",
		}
	}

	if o.Limit < 0 || o.Limit > 100 {
		return &ValidationError{
			Field:   "Limit",
			Value:   o.Limit,
			Message: "limit must be between 0 and 100",
		}
	}

	if o.Provider != "" && !isValidProviderName(o.Provider) {
		return &ValidationError{
			Field:   "Provider",
			Value:   o.Provider,
			Message: "invalid provider name format",
		}
	}

	return nil
}

// List returns a list of all modules
func (s *ModulesService) List(ctx context.Context, opts *ModuleListOptions) (*ModuleList, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	path := "modules"
	if opts != nil {
		values := url.Values{}
		if opts.Offset > 0 {
			values.Add("offset", fmt.Sprintf("%d", opts.Offset))
		}
		if opts.Limit > 0 {
			values.Add("limit", fmt.Sprintf("%d", opts.Limit))
		} else {
			values.Add("limit", "50") // Default limit
		}
		if opts.Provider != "" {
			values.Add("provider", opts.Provider)
		}
		if opts.Verified {
			values.Add("verified", "true")
		}
		if len(values) > 0 {
			path = fmt.Sprintf("%s?%s", path, values.Encode())
		}
	}

	var result ModuleList
	if err := s.client.get(ctx, path, "v1", &result); err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}

	return &result, nil
}

// Search searches for modules based on a query string
func (s *ModulesService) Search(ctx context.Context, query string, offset int) (*ModuleList, error) {
	if query == "" {
		return nil, &ValidationError{
			Field:   "query",
			Value:   query,
			Message: "search query cannot be empty",
		}
	}

	if offset < 0 {
		return nil, &ValidationError{
			Field:   "offset",
			Value:   offset,
			Message: "offset cannot be negative",
		}
	}

	path := fmt.Sprintf("modules/search?q=%s&offset=%d", url.QueryEscape(query), offset)

	var result ModuleList
	if err := s.client.get(ctx, path, "v1", &result); err != nil {
		return nil, fmt.Errorf("failed to search modules: %w", err)
	}

	return &result, nil
}

// Get returns details about a specific module version
func (s *ModulesService) Get(ctx context.Context, namespace, name, provider, version string) (*ModuleDetails, error) {
	if err := validateModuleParams(namespace, name, provider, version); err != nil {
		return nil, err
	}

	moduleID := fmt.Sprintf("%s/%s/%s/%s", namespace, name, provider, version)
	path := fmt.Sprintf("modules/%s", moduleID)

	var result ModuleDetails
	if err := s.client.get(ctx, path, "v1", &result); err != nil {
		return nil, fmt.Errorf("failed to get module %s: %w", moduleID, err)
	}

	return &result, nil
}

// GetByID returns details about a module using its full ID
func (s *ModulesService) GetByID(ctx context.Context, moduleID string) (*ModuleDetails, error) {
	if moduleID == "" {
		return nil, &ValidationError{
			Field:   "moduleID",
			Value:   moduleID,
			Message: "module ID cannot be empty",
		}
	}

	// Validate module ID format
	parts := strings.Split(moduleID, "/")
	if len(parts) != 4 {
		return nil, &ValidationError{
			Field:   "moduleID",
			Value:   moduleID,
			Message: "invalid module ID format, expected namespace/name/provider/version",
		}
	}

	return s.Get(ctx, parts[0], parts[1], parts[2], parts[3])
}

// ListVersions returns all versions of a module
func (s *ModulesService) ListVersions(ctx context.Context, namespace, name, provider string) ([]string, error) {
	if err := validateModuleParams(namespace, name, provider, ""); err != nil {
		return nil, err
	}

	// Call the dedicated versions endpoint instead of going via search/latest
	path := fmt.Sprintf("modules/%s/%s/%s/versions", url.PathEscape(namespace), url.PathEscape(name), url.PathEscape(provider))

	var resp struct {
		Modules []struct {
			Versions []struct {
				Version string `json:"version"`
			} `json:"versions"`
		} `json:"modules"`
	}

	if err := s.client.get(ctx, path, "v1", &resp); err != nil {
		return nil, fmt.Errorf("failed to list module versions: %w", err)
	}

	if len(resp.Modules) == 0 {
		return nil, &APIError{
			StatusCode: 404,
			Message:    fmt.Sprintf("module %s/%s/%s not found", namespace, name, provider),
		}
	}

	versions := make([]string, 0, len(resp.Modules[0].Versions))
	for _, v := range resp.Modules[0].Versions {
		if v.Version != "" {
			versions = append(versions, v.Version)
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for module %s/%s/%s", namespace, name, provider)
	}

	return versions, nil
}

// GetLatest returns the latest version of a module
func (s *ModulesService) GetLatest(ctx context.Context, namespace, name, provider string) (*ModuleDetails, error) {
	if err := validateModuleParams(namespace, name, provider, ""); err != nil {
		return nil, err
	}

	// Use ListVersions to get all versions, then pick the greatest semver
	versions, err := s.ListVersions(ctx, namespace, name, provider)
	if err != nil {
		return nil, err
	}

	latest := versions[0]
	for i := 1; i < len(versions); i++ {
		if CompareVersions(versions[i], latest) > 0 {
			latest = versions[i]
		}
	}

	// Return full details for the latest version
	return s.Get(ctx, namespace, name, provider, latest)
}

// Download returns the download URL for a module
func (s *ModulesService) Download(ctx context.Context, namespace, name, provider, version string) (string, error) {
	if err := validateModuleParams(namespace, name, provider, version); err != nil {
		return "", err
	}

	// Verify the module exists
	if _, err := s.Get(ctx, namespace, name, provider, version); err != nil {
		return "", fmt.Errorf("failed to verify module exists: %w", err)
	}

	// The download URL follows a specific pattern
	downloadURL := fmt.Sprintf("%s/v1/modules/%s/%s/%s/%s/download",
		s.client.baseURL, namespace, name, provider, version)

	return downloadURL, nil
}

// ModuleSearchResult represents a search result with relevance information
type ModuleSearchResult struct {
	Module
	Relevance float64 // Calculated relevance score
}

// SearchWithRelevance searches for modules and calculates relevance scores
func (s *ModulesService) SearchWithRelevance(ctx context.Context, query string, offset int) ([]ModuleSearchResult, error) {
	result, err := s.Search(ctx, query, offset)
	if err != nil {
		return nil, err
	}

	var searchResults []ModuleSearchResult
	queryLower := strings.ToLower(query)
	queryParts := strings.Fields(queryLower)

	for _, mod := range result.Modules {
		searchResult := ModuleSearchResult{
			Module: mod,
		}

		// Calculate relevance based on various factors
		relevance := 0.0

		nameLower := strings.ToLower(mod.Name)
		descLower := strings.ToLower(mod.Description)

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

		// Description match
		if strings.Contains(descLower, queryLower) {
			relevance += 3.0
		} else {
			// Check if all query parts are in the description
			allPartsInDesc := true
			for _, part := range queryParts {
				if !strings.Contains(descLower, part) {
					allPartsInDesc = false
					break
				}
			}
			if allPartsInDesc {
				relevance += 1.5
			}
		}

		// Namespace match
		if strings.Contains(strings.ToLower(mod.Namespace), queryLower) {
			relevance += 2.0
		}

		// Provider match
		if strings.Contains(strings.ToLower(mod.Provider), queryLower) {
			relevance += 1.0
		}

		// Verification status
		if mod.Verified {
			relevance += 2.0
		}

		// Download count (normalized, logarithmic scale)
		if mod.Downloads > 0 {
			downloadScore := logScale(float64(mod.Downloads), 1, 10000000, 0, 3)
			relevance += downloadScore
		}

		// Recency (if published recently)
		daysSincePublished := timeSince(mod.PublishedAt).Hours() / 24
		if daysSincePublished < 30 {
			relevance += 1.0
		} else if daysSincePublished < 90 {
			relevance += 0.5
		}

		searchResult.Relevance = relevance
		searchResults = append(searchResults, searchResult)
	}

	// Sort by relevance
	sort.Slice(searchResults, func(i, j int) bool {
		return searchResults[i].Relevance > searchResults[j].Relevance
	})

	return searchResults, nil
}

// validateModuleParams validates module parameters
func validateModuleParams(namespace, name, provider, version string) error {
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
	} else if !isValidModuleName(name) {
		errs.Add(&ValidationError{
			Field:   "name",
			Value:   name,
			Message: "invalid module name format",
		})
	}

	if provider == "" {
		errs.Add(&ValidationError{
			Field:   "provider",
			Value:   provider,
			Message: "provider cannot be empty",
		})
	} else if !isValidProviderName(provider) {
		errs.Add(&ValidationError{
			Field:   "provider",
			Value:   provider,
			Message: "invalid provider name format",
		})
	}

	if version != "" && !isValidVersion(version) {
		errs.Add(&ValidationError{
			Field:   "version",
			Value:   version,
			Message: "invalid version format",
		})
	}

	return errs.ErrorOrNil()
}

// Helper functions for validation
func isValidNamespace(namespace string) bool {
	// Namespace should contain only alphanumeric characters, hyphens, and underscores
	for _, r := range namespace {
		if !isAlphaNumeric(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func isValidModuleName(name string) bool {
	// Module name should contain only alphanumeric characters, hyphens, and underscores
	for _, r := range name {
		if !isAlphaNumeric(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

func isValidProviderName(provider string) bool {
	// Provider name should contain only lowercase letters and hyphens
	for _, r := range provider {
		if !isLowerAlpha(r) && r != '-' {
			return false
		}
	}
	return true
}

func isValidVersion(version string) bool {
	// Basic semantic version validation
	// Format: v1.2.3 or 1.2.3, optionally with pre-release
	if version == "" {
		return false
	}

	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Check basic format
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	// Each part should be a number
	for i, part := range parts {
		if i == 2 {
			// The patch version might have a pre-release suffix
			dashIndex := strings.Index(part, "-")
			if dashIndex > 0 {
				part = part[:dashIndex]
			}
		}

		for _, r := range part {
			if !isDigit(r) {
				return false
			}
		}
	}

	return true
}

// Character type checking functions
func isAlphaNumeric(r rune) bool {
	return isLowerAlpha(r) || isUpperAlpha(r) || isDigit(r)
}

func isLowerAlpha(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUpperAlpha(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// logScale maps a value from one range to another using logarithmic scaling
func logScale(value, minIn, maxIn, minOut, maxOut float64) float64 {
	if value <= minIn {
		return minOut
	}
	if value >= maxIn {
		return maxOut
	}

	// Use log10 for scaling
	logMin := log10(minIn)
	logMax := log10(maxIn)
	logValue := log10(value)

	// Linear interpolation in log space
	normalized := (logValue - logMin) / (logMax - logMin)
	return minOut + normalized*(maxOut-minOut)
}

// log10 computes the base-10 logarithm
func log10(x float64) float64 {
	// Simple implementation of log10
	// In production, use math.Log10
	if x <= 0 {
		return 0
	}

	// Count the number of times we can divide by 10
	count := 0.0
	for x >= 10 {
		x /= 10
		count++
	}

	// Add fractional part (simplified)
	if x > 1 {
		count += (x - 1) / 9
	}

	return count
}

// timeSince returns the duration since the given time
func timeSince(t time.Time) time.Duration {
	return time.Since(t)
}
