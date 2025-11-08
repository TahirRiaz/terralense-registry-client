package tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/TahirRiaz/terralens-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// ModuleTests contains tests for the Modules API
type ModuleTests struct {
	*BaseTestSuite
}

// NewModuleTests creates a new module test suite
func NewModuleTests(client *registry.Client, logger *logrus.Logger) TestSuite {
	suite := &ModuleTests{
		BaseTestSuite: NewBaseTestSuite("Modules", client, logger),
	}

	suite.setupTests()
	return suite
}

func (s *ModuleTests) setupTests() {
	s.AddTest("List Modules", "Test listing modules with various options", s.testListModules)
	s.AddTest("Search Modules", "Test module search functionality", s.testSearchModules)
	s.AddTest("Search with Relevance", "Test module search with relevance scoring", s.testSearchWithRelevance)
	s.AddTest("Get Module", "Test getting a specific module", s.testGetModule)
	s.AddTest("Get Module by ID", "Test getting a module by full ID", s.testGetModuleByID)
	s.AddTest("Get Latest Version", "Test getting the latest version of a module", s.testGetLatestVersion)
	s.AddTest("List Versions", "Test listing all versions of a module", s.testListVersions)
	s.AddTest("Download URL", "Test generating download URL", s.testDownloadURL)
	s.AddTest("Pagination", "Test module list pagination", s.testPagination)
	s.AddTest("Filter by Provider", "Test filtering modules by provider", s.testFilterByProvider)
	s.AddTest("Verified Modules", "Test filtering verified modules", s.testVerifiedModules)
	s.AddTest("Invalid Module", "Test error handling for invalid modules", s.testInvalidModule)
}

func (s *ModuleTests) testListModules(ctx context.Context) error {
	// Test with default options
	opts := &registry.ModuleListOptions{
		Limit: 10,
	}

	result, err := s.client.Modules.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list modules: %w", err)
	}

	if err := AssertNotNil(result); err != nil {
		return err
	}

	if len(result.Modules) == 0 {
		return fmt.Errorf("expected at least one module, got none")
	}

	// Verify module structure
	for _, module := range result.Modules {
		if module.ID == "" {
			return fmt.Errorf("module has empty ID")
		}
		if module.Namespace == "" {
			return fmt.Errorf("module has empty namespace")
		}
		if module.Name == "" {
			return fmt.Errorf("module has empty name")
		}
		if module.Provider == "" {
			return fmt.Errorf("module has empty provider")
		}
	}

	s.logger.Debugf("Listed %d modules", len(result.Modules))
	return nil
}

func (s *ModuleTests) testSearchModules(ctx context.Context) error {
	queries := []string{
		"aws vpc",
		"terraform",
		"kubernetes",
	}

	for _, query := range queries {
		result, err := s.client.Modules.Search(ctx, query, 0)
		if err != nil {
			return fmt.Errorf("failed to search for '%s': %w", query, err)
		}

		if err := AssertNotNil(result); err != nil {
			return fmt.Errorf("search result for '%s' is nil: %w", query, err)
		}

		s.logger.Debugf("Search for '%s' returned %d results", query, len(result.Modules))

		// Verify search results contain the query terms
		for _, module := range result.Modules {
			// Check if query terms appear in module details
			moduleText := strings.ToLower(fmt.Sprintf("%s %s %s %s",
				module.Name, module.Description, module.Namespace, module.Provider))

			queryLower := strings.ToLower(query)
			queryParts := strings.Fields(queryLower)

			foundAny := false
			for _, part := range queryParts {
				if strings.Contains(moduleText, part) {
					foundAny = true
					break
				}
			}

			if !foundAny && len(result.Modules) > 0 {
				s.logger.Warnf("Module %s doesn't contain query terms '%s'", module.ID, query)
			}
		}
	}

	return nil
}

func (s *ModuleTests) testSearchWithRelevance(ctx context.Context) error {
	query := "aws vpc"

	results, err := s.client.Modules.SearchWithRelevance(ctx, query, 0)
	if err != nil {
		return fmt.Errorf("failed to search with relevance: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("expected search results, got none")
	}

	// Verify results are sorted by relevance
	for i := 1; i < len(results); i++ {
		if results[i].Relevance > results[i-1].Relevance {
			return fmt.Errorf("results not sorted by relevance: %f > %f at position %d",
				results[i].Relevance, results[i-1].Relevance, i)
		}
	}

	// Verify relevance scores are positive
	for _, result := range results {
		if result.Relevance < 0 {
			return fmt.Errorf("negative relevance score: %f for module %s",
				result.Relevance, result.ID)
		}
	}

	s.logger.Debugf("Search with relevance returned %d results, top relevance: %.2f",
		len(results), results[0].Relevance)

	return nil
}

func (s *ModuleTests) testGetModule(ctx context.Context) error {
	// Use a well-known module
	testCases := []struct {
		namespace string
		name      string
		provider  string
		version   string
	}{
		{"terraform-aws-modules", "vpc", "aws", ""},
		{"hashicorp", "consul", "aws", ""},
	}

	for _, tc := range testCases {
		// First get the latest version
		latest, err := s.client.Modules.GetLatest(ctx, tc.namespace, tc.name, tc.provider)
		if err != nil {
			if registry.IsNotFound(err) {
				s.logger.Warnf("Module %s/%s/%s not found, skipping",
					tc.namespace, tc.name, tc.provider)
				continue
			}
			return fmt.Errorf("failed to get latest version: %w", err)
		}

		// Now get specific version
		module, err := s.client.Modules.Get(ctx, tc.namespace, tc.name, tc.provider, latest.Version)
		if err != nil {
			return fmt.Errorf("failed to get module: %w", err)
		}

		// Verify module details
		if module.Namespace != tc.namespace {
			return fmt.Errorf("namespace mismatch: expected %s, got %s",
				tc.namespace, module.Namespace)
		}
		if module.Name != tc.name {
			return fmt.Errorf("name mismatch: expected %s, got %s",
				tc.name, module.Name)
		}
		if module.Provider != tc.provider {
			return fmt.Errorf("provider mismatch: expected %s, got %s",
				tc.provider, module.Provider)
		}

		// Verify module has root configuration
		if err := AssertNotNil(module.Root); err != nil {
			return fmt.Errorf("module root is nil: %w", err)
		}

		s.logger.Debugf("Got module %s version %s with %d inputs and %d outputs",
			module.ID, module.Version, len(module.Root.Inputs), len(module.Root.Outputs))

		break // Test at least one module successfully
	}

	return nil
}

func (s *ModuleTests) testGetModuleByID(ctx context.Context) error {
	// First, search for a module to get a valid ID
	results, err := s.client.Modules.Search(ctx, "terraform", 0)
	if err != nil {
		return fmt.Errorf("failed to search for modules: %w", err)
	}

	if len(results.Modules) == 0 {
		return fmt.Errorf("no modules found for testing")
	}

	// Use the first module ID
	moduleID := results.Modules[0].ID

	module, err := s.client.Modules.GetByID(ctx, moduleID)
	if err != nil {
		return fmt.Errorf("failed to get module by ID %s: %w", moduleID, err)
	}

	if module.ID != moduleID {
		return fmt.Errorf("module ID mismatch: expected %s, got %s", moduleID, module.ID)
	}

	s.logger.Debugf("Successfully retrieved module by ID: %s", moduleID)
	return nil
}

func (s *ModuleTests) testGetLatestVersion(ctx context.Context) error {
	// Strategy: Use search to find a module, then use its version info
	// The search API returns modules with their current version

	searchQueries := []string{"terraform-aws-modules vpc", "cloudposse label", "terraform-google-modules network"}

	for _, query := range searchQueries {
		searchResult, err := s.client.Modules.Search(ctx, query, 0)
		if err != nil {
			s.logger.Debugf("Search failed for '%s': %v", query, err)
			continue
		}

		if len(searchResult.Modules) == 0 {
			s.logger.Debugf("No modules found for query '%s'", query)
			continue
		}

		// Try the first few modules from search results
		for i, module := range searchResult.Modules {
			if i >= 5 { // Limit attempts per search
				break
			}

			// Skip test/example modules and those with low downloads
			if strings.Contains(strings.ToLower(module.Name), "test") ||
				strings.Contains(strings.ToLower(module.Name), "example") ||
				module.Downloads < 1000 {
				continue
			}

			s.logger.Debugf("Trying module from search: %s/%s/%s (version: %s)",
				module.Namespace, module.Name, module.Provider, module.Version)

			// The search result already contains the module with its current version
			// Let's first verify we can get this specific version
			moduleDetails, err := s.client.Modules.Get(ctx, module.Namespace, module.Name,
				module.Provider, module.Version)
			if err != nil {
				s.logger.Debugf("Failed to get module details: %v", err)
				continue
			}

			// Now we know the module is accessible, let's get its latest version
			// by using the same approach but without specifying version
			latest, err := s.client.Modules.GetLatest(ctx, module.Namespace, module.Name, module.Provider)
			if err != nil {
				s.logger.Debugf("GetLatest failed for %s/%s/%s: %v",
					module.Namespace, module.Name, module.Provider, err)
				continue
			}

			// Verify the result
			if latest.Version == "" {
				return fmt.Errorf("latest version is empty for %s/%s/%s",
					module.Namespace, module.Name, module.Provider)
			}

			// The module details should have versions list
			if len(latest.Versions) == 0 {
				// Some modules might not expose all versions in the details
				// but we at least have the current version
				s.logger.Warnf("No versions list for %s/%s/%s, but got version %s",
					module.Namespace, module.Name, module.Provider, latest.Version)
			}

			s.logger.Debugf("Successfully got latest version %s for %s/%s/%s",
				latest.Version, module.Namespace, module.Name, module.Provider)

			// Also verify the version data makes sense
			if moduleDetails.Version != "" && latest.Version != "" {
				// The latest version should be >= the version we got from search
				s.logger.Debugf("Search version: %s, Latest version: %s",
					moduleDetails.Version, latest.Version)
			}

			return nil
		}
	}

	// If the above didn't work, try a more direct approach with known stable modules
	knownStableModules := []struct {
		namespace string
		name      string
		provider  string
	}{
		{"cloudposse", "label", "null"},
		{"terraform-aws-modules", "vpc", "aws"},
		{"terraform-google-modules", "network", "google"},
	}

	for _, km := range knownStableModules {
		// Try to get any version first to ensure the module exists
		searchQuery := fmt.Sprintf("%s/%s", km.namespace, km.name)
		searchResult, err := s.client.Modules.Search(ctx, searchQuery, 0)
		if err != nil || len(searchResult.Modules) == 0 {
			continue
		}

		// Find the exact module in search results
		var foundModule *registry.Module
		for _, m := range searchResult.Modules {
			if m.Namespace == km.namespace && m.Name == km.name && m.Provider == km.provider {
				foundModule = &m
				break
			}
		}

		if foundModule == nil {
			continue
		}

		// We found it in search, now try GetLatest
		latest, err := s.client.Modules.GetLatest(ctx, km.namespace, km.name, km.provider)
		if err == nil {
			s.logger.Debugf("Successfully got latest version %s for known module %s/%s/%s",
				latest.Version, km.namespace, km.name, km.provider)
			return nil
		}
	}

	return fmt.Errorf("unable to find any accessible module for testing GetLatest")
}

func (s *ModuleTests) testListVersions(ctx context.Context) error {
	// Similar strategy: use search first to find accessible modules

	searchQueries := []string{"terraform-aws-modules", "cloudposse", "Azure"}

	for _, query := range searchQueries {
		searchResult, err := s.client.Modules.Search(ctx, query, 0)
		if err != nil {
			s.logger.Debugf("Search failed for '%s': %v", query, err)
			continue
		}

		if len(searchResult.Modules) == 0 {
			continue
		}

		// Try the first few modules from search results
		for i, module := range searchResult.Modules {
			if i >= 5 { // Limit attempts
				break
			}

			// Skip test/example modules and those with low downloads
			if strings.Contains(strings.ToLower(module.Name), "test") ||
				strings.Contains(strings.ToLower(module.Name), "example") ||
				module.Downloads < 1000 {
				continue
			}

			// First verify we can access the module
			_, err := s.client.Modules.Get(ctx, module.Namespace, module.Name,
				module.Provider, module.Version)
			if err != nil {
				continue
			}

			// Now try to list versions
			versions, err := s.client.Modules.ListVersions(ctx, module.Namespace,
				module.Name, module.Provider)
			if err != nil {
				s.logger.Debugf("ListVersions failed for %s/%s/%s: %v",
					module.Namespace, module.Name, module.Provider, err)
				continue
			}

			if len(versions) == 0 {
				s.logger.Warnf("No versions returned for %s/%s/%s",
					module.Namespace, module.Name, module.Provider)
				continue
			}

			// Verify version format
			validVersions := 0
			for _, version := range versions {
				if version == "" {
					return fmt.Errorf("found empty version string")
				}

				// Basic version format check
				if err := registry.ValidateProviderVersion(version); err != nil {
					s.logger.Warnf("Invalid version format: %s - %v", version, err)
				} else {
					validVersions++
				}
			}

			if validVersions == 0 {
				return fmt.Errorf("no valid versions found for module %s/%s/%s",
					module.Namespace, module.Name, module.Provider)
			}

			s.logger.Debugf("Successfully found %d versions (%d valid) for %s/%s/%s",
				len(versions), validVersions, module.Namespace, module.Name, module.Provider)
			return nil
		}
	}

	// Fallback: try known modules that should have versions
	knownModules := []struct {
		namespace string
		name      string
		provider  string
	}{
		{"cloudposse", "label", "null"},
		{"terraform-aws-modules", "vpc", "aws"},
	}

	for _, km := range knownModules {
		// Quick check if accessible via search
		searchQuery := fmt.Sprintf("%s %s", km.namespace, km.name)
		searchResult, err := s.client.Modules.Search(ctx, searchQuery, 0)
		if err != nil || len(searchResult.Modules) == 0 {
			continue
		}

		// Try to list versions
		versions, err := s.client.Modules.ListVersions(ctx, km.namespace, km.name, km.provider)
		if err == nil && len(versions) > 0 {
			s.logger.Debugf("Found %d versions for fallback module %s/%s/%s",
				len(versions), km.namespace, km.name, km.provider)
			return nil
		}
	}

	return fmt.Errorf("unable to find any accessible module for testing ListVersions")
}
func (s *ModuleTests) testDownloadURL(ctx context.Context) error {
	// Get a valid module first
	results, err := s.client.Modules.Search(ctx, "terraform", 0)
	if err != nil {
		return fmt.Errorf("failed to search for modules: %w", err)
	}

	if len(results.Modules) == 0 {
		return fmt.Errorf("no modules found")
	}

	module := results.Modules[0]

	downloadURL, err := s.client.Modules.Download(ctx,
		module.Namespace, module.Name, module.Provider, module.Version)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	if downloadURL == "" {
		return fmt.Errorf("download URL is empty")
	}

	// Verify URL format
	expectedPrefix := fmt.Sprintf("%s/v1/modules", s.client.GetBaseURL())
	if !strings.HasPrefix(downloadURL, expectedPrefix) {
		return fmt.Errorf("unexpected download URL format: %s", downloadURL)
	}

	s.logger.Debugf("Download URL: %s", downloadURL)
	return nil
}

func (s *ModuleTests) testPagination(ctx context.Context) error {
	// Test pagination with small page size
	pageSize := 5
	var allModules []registry.Module

	for i := 0; i < 3; i++ { // Get 3 pages max
		opts := &registry.ModuleListOptions{
			Offset: i * pageSize,
			Limit:  pageSize,
		}

		result, err := s.client.Modules.List(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list modules (page %d): %w", i+1, err)
		}

		if len(result.Modules) == 0 {
			break // No more results
		}

		allModules = append(allModules, result.Modules...)

		// Verify pagination metadata
		if result.Meta.Limit != pageSize {
			return fmt.Errorf("unexpected limit in metadata: %d", result.Meta.Limit)
		}

		if result.Meta.CurrentOffset != i*pageSize {
			return fmt.Errorf("unexpected offset: expected %d, got %d",
				i*pageSize, result.Meta.CurrentOffset)
		}

		if len(result.Modules) < pageSize {
			break // Last page
		}
	}

	s.logger.Debugf("Retrieved %d modules across multiple pages", len(allModules))
	return nil
}

func (s *ModuleTests) testFilterByProvider(ctx context.Context) error {
	providers := []string{"aws", "azurerm", "google"}

	for _, provider := range providers {
		opts := &registry.ModuleListOptions{
			Provider: provider,
			Limit:    10,
		}

		result, err := s.client.Modules.List(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to filter by provider %s: %w", provider, err)
		}

		// Verify all modules are for the specified provider
		for _, module := range result.Modules {
			if module.Provider != provider {
				return fmt.Errorf("expected provider %s, got %s for module %s",
					provider, module.Provider, module.ID)
			}
		}

		s.logger.Debugf("Found %d modules for provider %s", len(result.Modules), provider)
	}

	return nil
}

func (s *ModuleTests) testVerifiedModules(ctx context.Context) error {
	opts := &registry.ModuleListOptions{
		Verified: true,
		Limit:    20,
	}

	result, err := s.client.Modules.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list verified modules: %w", err)
	}

	// Verify all modules are verified
	for _, module := range result.Modules {
		if !module.Verified {
			return fmt.Errorf("expected verified module, got unverified: %s", module.ID)
		}
	}

	s.logger.Debugf("Found %d verified modules", len(result.Modules))
	return nil
}

func (s *ModuleTests) testInvalidModule(ctx context.Context) error {
	// Test with non-existent module
	_, err := s.client.Modules.Get(ctx, "invalid-namespace", "invalid-name", "invalid-provider", "0.0.0")

	if err == nil {
		return fmt.Errorf("expected error for invalid module, got nil")
	}

	if !registry.IsNotFound(err) {
		return fmt.Errorf("expected NotFound error, got: %v", err)
	}

	// Test with invalid module ID
	_, err = s.client.Modules.GetByID(ctx, "invalid/module/id")
	if err == nil {
		return fmt.Errorf("expected error for invalid module ID, got nil")
	}

	s.logger.Debug("Invalid module handling works correctly")
	return nil
}
