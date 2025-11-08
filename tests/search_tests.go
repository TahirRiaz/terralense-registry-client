package tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/TahirRiaz/terralens-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// SearchTests contains tests for search functionality across all services
type SearchTests struct {
	*BaseTestSuite
}

// NewSearchTests creates a new search test suite
func NewSearchTests(client *registry.Client, logger *logrus.Logger) TestSuite {
	suite := &SearchTests{
		BaseTestSuite: NewBaseTestSuite("Search", client, logger),
	}

	suite.setupTests()
	return suite
}

func (s *SearchTests) setupTests() {
	s.AddTest("Module Search Relevance", "Test module search relevance scoring", s.testModuleSearchRelevance)
	s.AddTest("Policy Search Relevance", "Test policy search relevance scoring", s.testPolicySearchRelevance)
	s.AddTest("Cross-Provider Search", "Test searching across different providers", s.testCrossProviderSearch)
	s.AddTest("Empty Search", "Test handling of empty search queries", s.testEmptySearch)
	s.AddTest("Special Characters", "Test search with special characters", s.testSpecialCharacters)
	s.AddTest("Case Sensitivity", "Test case sensitivity in search", s.testCaseSensitivity)
	s.AddTest("Partial Matches", "Test partial word matching", s.testPartialMatches)
	s.AddTest("Multi-Word Search", "Test multi-word search queries", s.testMultiWordSearch)
}

func (s *SearchTests) testModuleSearchRelevance(ctx context.Context) error {
	// Test that exact matches rank higher
	queries := []struct {
		query         string
		expectedInTop string // Expected module name in top results
	}{
		{"vpc", "vpc"},
		{"kubernetes", "kubernetes"},
		{"consul", "consul"},
	}

	for _, test := range queries {
		results, err := s.client.Modules.SearchWithRelevance(ctx, test.query, 0)
		if err != nil {
			return fmt.Errorf("search failed for '%s': %w", test.query, err)
		}

		if len(results) == 0 {
			s.logger.Warnf("No results for query '%s'", test.query)
			continue
		}

		// Check if expected result is in top 5
		foundInTop := false
		for i := 0; i < len(results) && i < 5; i++ {
			if strings.Contains(strings.ToLower(results[i].Name), test.expectedInTop) {
				foundInTop = true

				// Verify this has high relevance
				if results[i].Relevance < 5.0 {
					s.logger.Warnf("Expected higher relevance for exact match '%s', got %.2f",
						results[i].Name, results[i].Relevance)
				}
				break
			}
		}

		if !foundInTop {
			s.logger.Warnf("Expected '%s' in top results for query '%s'",
				test.expectedInTop, test.query)
		}

		// Verify relevance scores are descending
		for i := 1; i < len(results); i++ {
			if results[i].Relevance > results[i-1].Relevance {
				return fmt.Errorf("relevance scores not in descending order at position %d", i)
			}
		}

		s.logger.Debugf("Query '%s' returned %d results, top relevance: %.2f",
			test.query, len(results), results[0].Relevance)
	}

	return nil
}

func (s *SearchTests) testPolicySearchRelevance(ctx context.Context) error {
	// Test policy search with relevance
	queries := []string{"aws", "compliance", "security"}

	for _, query := range queries {
		results, err := s.client.Policies.Search(ctx, query)
		if err != nil {
			return fmt.Errorf("policy search failed for '%s': %w", query, err)
		}

		if len(results) == 0 {
			s.logger.Warnf("No policy results for query '%s'", query)
			continue
		}

		// Verify relevance calculation
		for _, result := range results {
			// Name match should have higher relevance
			if strings.Contains(strings.ToLower(result.Policy.Attributes.Name), strings.ToLower(query)) {
				if result.Relevance < 5.0 {
					s.logger.Warnf("Expected higher relevance for name match in policy %s",
						result.Policy.Attributes.Name)
				}
			}

			// Verified policies should have bonus relevance
			if result.Policy.Attributes.Verified && result.Relevance < 2.0 {
				s.logger.Warnf("Verified policy %s has low relevance: %.2f",
					result.Policy.Attributes.Name, result.Relevance)
			}
		}

		s.logger.Debugf("Policy search '%s' returned %d results", query, len(results))
	}

	return nil
}

func (s *SearchTests) testCrossProviderSearch(ctx context.Context) error {
	// Search for modules across different providers
	providers := map[string]string{
		"aws":        "aws",
		"azure":      "azurerm",
		"google":     "google",
		"kubernetes": "kubernetes",
	}

	for query, expectedProvider := range providers {
		results, err := s.client.Modules.Search(ctx, query, 0)
		if err != nil {
			return fmt.Errorf("search failed for '%s': %w", query, err)
		}

		if len(results.Modules) == 0 {
			s.logger.Warnf("No results for provider query '%s'", query)
			continue
		}

		// Count modules from expected provider
		providerCount := 0
		for _, module := range results.Modules {
			if module.Provider == expectedProvider {
				providerCount++
			}
		}

		if providerCount == 0 {
			s.logger.Warnf("No modules found for provider %s when searching '%s'",
				expectedProvider, query)
		} else {
			s.logger.Debugf("Found %d modules for provider %s", providerCount, expectedProvider)
		}
	}

	return nil
}

func (s *SearchTests) testEmptySearch(ctx context.Context) error {
	// Test empty search query
	_, err := s.client.Modules.Search(ctx, "", 0)

	if err == nil {
		return fmt.Errorf("expected error for empty search query, got nil")
	}

	if !registry.IsValidationError(err) {
		return fmt.Errorf("expected validation error for empty query, got: %v", err)
	}

	// Test whitespace-only query
	_, err = s.client.Modules.Search(ctx, "   ", 0)
	if err != nil && registry.IsValidationError(err) {
		s.logger.Debug("Whitespace-only query correctly rejected")
	}

	return nil
}

func (s *SearchTests) testSpecialCharacters(ctx context.Context) error {
	// Test search with special characters
	queries := []string{
		"aws-vpc",
		"terraform_module",
		"azure.network",
		"module/test", // This might be rejected
	}

	for _, query := range queries {
		results, err := s.client.Modules.Search(ctx, query, 0)

		if err != nil {
			// Some special characters might cause errors
			if strings.Contains(query, "/") {
				s.logger.Debugf("Query with '/' correctly rejected: %s", query)
				continue
			}
			return fmt.Errorf("unexpected error for query '%s': %w", query, err)
		}

		s.logger.Debugf("Query '%s' returned %d results", query, len(results.Modules))
	}

	return nil
}

func (s *SearchTests) testCaseSensitivity(ctx context.Context) error {
	// Test that search is case-insensitive
	queries := [][]string{
		{"AWS", "aws", "Aws"},
		{"VPC", "vpc", "Vpc"},
		{"KUBERNETES", "kubernetes", "Kubernetes"},
	}

	for _, querySet := range queries {
		var previousCount int

		for i, query := range querySet {
			results, err := s.client.Modules.Search(ctx, query, 0)
			if err != nil {
				return fmt.Errorf("search failed for '%s': %w", query, err)
			}

			if i > 0 && len(results.Modules) != previousCount {
				s.logger.Warnf("Different result count for case variations: '%s' (%d) vs previous (%d)",
					query, len(results.Modules), previousCount)
			}

			previousCount = len(results.Modules)
			s.logger.Debugf("Query '%s' returned %d results", query, previousCount)
		}
	}

	return nil
}

func (s *SearchTests) testPartialMatches(ctx context.Context) error {
	// Test partial word matching
	partialQueries := []struct {
		partial  string
		fullWord string
	}{
		{"kube", "kubernetes"},
		{"terra", "terraform"},
		{"elas", "elasticsearch"},
	}

	for _, test := range partialQueries {
		partialResults, err := s.client.Modules.Search(ctx, test.partial, 0)
		if err != nil {
			return fmt.Errorf("search failed for partial '%s': %w", test.partial, err)
		}

		fullResults, err := s.client.Modules.Search(ctx, test.fullWord, 0)
		if err != nil {
			return fmt.Errorf("search failed for full '%s': %w", test.fullWord, err)
		}

		// Check if partial search includes results from full word search
		partialContainsFull := false
		for _, partialModule := range partialResults.Modules {
			if strings.Contains(strings.ToLower(partialModule.Name), test.fullWord) ||
				strings.Contains(strings.ToLower(partialModule.Description), test.fullWord) {
				partialContainsFull = true
				break
			}
		}

		if len(fullResults.Modules) > 0 && !partialContainsFull {
			s.logger.Warnf("Partial search '%s' doesn't include results for '%s'",
				test.partial, test.fullWord)
		}

		s.logger.Debugf("Partial '%s': %d results, Full '%s': %d results",
			test.partial, len(partialResults.Modules),
			test.fullWord, len(fullResults.Modules))
	}

	return nil
}

func (s *SearchTests) testMultiWordSearch(ctx context.Context) error {
	// Test multi-word search queries
	multiWordQueries := []string{
		"aws vpc",
		"azure virtual network",
		"google cloud storage",
		"kubernetes ingress controller",
	}

	for _, query := range multiWordQueries {
		results, err := s.client.Modules.SearchWithRelevance(ctx, query, 0)
		if err != nil {
			return fmt.Errorf("search failed for '%s': %w", query, err)
		}

		if len(results) == 0 {
			s.logger.Warnf("No results for multi-word query '%s'", query)
			continue
		}

		// Verify that results contain at least some of the query terms
		queryWords := strings.Fields(strings.ToLower(query))

		relevantResults := 0
		for _, result := range results {
			moduleText := strings.ToLower(fmt.Sprintf("%s %s",
				result.Name, result.Description))

			matchCount := 0
			for _, word := range queryWords {
				if strings.Contains(moduleText, word) {
					matchCount++
				}
			}

			// Consider it relevant if it contains at least half the query words
			if matchCount >= len(queryWords)/2 {
				relevantResults++
			}
		}

		relevanceRatio := float64(relevantResults) / float64(len(results))
		if relevanceRatio < 0.5 {
			s.logger.Warnf("Low relevance ratio (%.2f) for query '%s'",
				relevanceRatio, query)
		}

		s.logger.Debugf("Multi-word query '%s': %d results, %.0f%% relevant",
			query, len(results), relevanceRatio*100)
	}

	return nil
}
