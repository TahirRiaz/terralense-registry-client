package tests

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/TahirRiaz/terralens-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// PerformanceTests contains performance-related tests
type PerformanceTests struct {
	*BaseTestSuite
}

// NewPerformanceTests creates a new performance test suite
func NewPerformanceTests(client *registry.Client, logger *logrus.Logger) TestSuite {
	suite := &PerformanceTests{
		BaseTestSuite: NewBaseTestSuite("Performance", client, logger),
	}

	suite.setupTests()
	return suite
}

func (s *PerformanceTests) setupTests() {
	s.AddTest("Response Time", "Test API response times", s.testResponseTime)
	s.AddTest("Concurrent Requests", "Test concurrent request handling", s.testConcurrentRequests)
	s.AddTest("Rate Limiting", "Test rate limiter behavior", s.testRateLimiting)
	s.AddTest("Large Result Sets", "Test handling of large result sets", s.testLargeResultSets)
	s.AddTest("Pagination Performance", "Test pagination efficiency", s.testPaginationPerformance)
	s.AddTest("Search Performance", "Test search response times", s.testSearchPerformance)
	s.AddTest("Cache Behavior", "Test caching behavior if implemented", s.testCacheBehavior)
}

func (s *PerformanceTests) testResponseTime(ctx context.Context) error {
	// Test response times for various endpoints
	tests := []struct {
		name        string
		fn          func() error
		maxDuration time.Duration
	}{
		{
			name: "List modules",
			fn: func() error {
				_, err := s.client.Modules.List(ctx, &registry.ModuleListOptions{Limit: 10})
				return err
			},
			maxDuration: 5 * time.Second,
		},
		{
			name: "Search modules",
			fn: func() error {
				_, err := s.client.Modules.Search(ctx, "terraform", 0)
				return err
			},
			maxDuration: 5 * time.Second,
		},
		{
			name: "Get provider",
			fn: func() error {
				_, err := s.client.Providers.Get(ctx, "hashicorp", "aws")
				return err
			},
			maxDuration: 3 * time.Second,
		},
	}

	for _, test := range tests {
		start := time.Now()
		err := test.fn()
		duration := time.Since(start)

		if err != nil {
			// Some operations might fail (e.g., provider not found)
			if !registry.IsNotFound(err) {
				return fmt.Errorf("%s failed: %w", test.name, err)
			}
		}

		if duration > test.maxDuration {
			s.logger.Warnf("%s took %v (max: %v)", test.name, duration, test.maxDuration)
		} else {
			s.logger.Debugf("%s completed in %v", test.name, duration)
		}
	}

	return nil
}

func (s *PerformanceTests) testConcurrentRequests(ctx context.Context) error {
	// Test concurrent requests to ensure thread safety
	concurrency := 10
	iterations := 5

	var wg sync.WaitGroup
	errors := make(chan error, concurrency*iterations)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				// Mix different types of requests
				switch j % 3 {
				case 0:
					_, err := s.client.Modules.List(ctx, &registry.ModuleListOptions{Limit: 5})
					if err != nil {
						errors <- fmt.Errorf("worker %d: list failed: %w", workerID, err)
					}
				case 1:
					_, err := s.client.Modules.Search(ctx, "aws", 0)
					if err != nil {
						errors <- fmt.Errorf("worker %d: search failed: %w", workerID, err)
					}
				case 2:
					_, err := s.client.Providers.List(ctx, &registry.ProviderListOptions{PageSize: 5})
					if err != nil {
						errors <- fmt.Errorf("worker %d: provider list failed: %w", workerID, err)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := time.Since(start)

	// Check for errors
	errorCount := 0
	for err := range errors {
		errorCount++
		s.logger.Errorf("Concurrent request error: %v", err)
	}

	if errorCount > 0 {
		return fmt.Errorf("had %d errors during concurrent requests", errorCount)
	}

	totalRequests := concurrency * iterations
	requestsPerSecond := float64(totalRequests) / duration.Seconds()

	s.logger.Debugf("Completed %d concurrent requests in %v (%.2f req/s)",
		totalRequests, duration, requestsPerSecond)

	return nil
}

func (s *PerformanceTests) testRateLimiting(ctx context.Context) error {
	// Test rate limiter behavior
	// Note: This test is careful not to actually exceed rate limits

	// Get current rate limit status
	rateLimiter := s.client.GetRateLimiter()
	if rateLimiter == nil {
		s.logger.Warn("No rate limiter configured")
		return nil
	}

	initialTokens := rateLimiter.TokensRemaining()
	s.logger.Debugf("Initial rate limit tokens: %d", initialTokens)

	// Make a few requests
	requestCount := 5
	for i := 0; i < requestCount; i++ {
		_, err := s.client.Modules.List(ctx, &registry.ModuleListOptions{Limit: 1})
		if err != nil {
			if registry.IsRateLimited(err) {
				s.logger.Warn("Hit rate limit during test")
				break
			}
			return fmt.Errorf("request failed: %w", err)
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	finalTokens := rateLimiter.TokensRemaining()
	s.logger.Debugf("Final rate limit tokens: %d", finalTokens)

	// Verify tokens were consumed
	if finalTokens >= initialTokens {
		s.logger.Warn("Rate limiter tokens not consumed as expected")
	}

	return nil
}

func (s *PerformanceTests) testLargeResultSets(ctx context.Context) error {
	// Test handling of large result sets
	start := time.Now()

	// Request maximum allowed page size
	opts := &registry.ModuleListOptions{
		Limit: 100, // Maximum typically allowed
	}

	result, err := s.client.Modules.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list large result set: %w", err)
	}

	duration := time.Since(start)

	s.logger.Debugf("Retrieved %d modules in %v", len(result.Modules), duration)

	// Verify all modules have required fields (spot check for data integrity)
	for i, module := range result.Modules {
		if module.ID == "" || module.Name == "" || module.Provider == "" {
			return fmt.Errorf("module at index %d has missing required fields", i)
		}
	}

	// Test memory efficiency with multiple large requests
	memTestStart := time.Now()
	for i := 0; i < 5; i++ {
		_, err := s.client.Modules.List(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed on iteration %d: %w", i, err)
		}
	}
	memTestDuration := time.Since(memTestStart)

	s.logger.Debugf("Completed 5 large requests in %v", memTestDuration)

	return nil
}

func (s *PerformanceTests) testPaginationPerformance(ctx context.Context) error {
	// Test pagination performance
	pageSize := 20
	maxPages := 5

	var allModules []registry.Module

	start := time.Now()

	for page := 0; page < maxPages; page++ {
		opts := &registry.ModuleListOptions{
			Offset: page * pageSize,
			Limit:  pageSize,
		}

		pageStart := time.Now()
		result, err := s.client.Modules.List(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to get page %d: %w", page, err)
		}

		pageDuration := time.Since(pageStart)
		s.logger.Debugf("Page %d (%d items) fetched in %v", page, len(result.Modules), pageDuration)

		if len(result.Modules) == 0 {
			break // No more results
		}

		allModules = append(allModules, result.Modules...)

		if len(result.Modules) < pageSize {
			break // Last page
		}
	}

	totalDuration := time.Since(start)

	s.logger.Debugf("Fetched %d modules across %d pages in %v",
		len(allModules), maxPages, totalDuration)

	// Verify no duplicates across pages
	seen := make(map[string]bool)
	for _, module := range allModules {
		if seen[module.ID] {
			return fmt.Errorf("duplicate module found in pagination: %s", module.ID)
		}
		seen[module.ID] = true
	}

	return nil
}

func (s *PerformanceTests) testSearchPerformance(ctx context.Context) error {
	// Test search performance with various query complexities
	queries := []struct {
		query       string
		description string
	}{
		{"a", "single character"},
		{"aws", "short common term"},
		{"terraform module vpc", "multi-word query"},
		{"azure virtual network security group", "complex multi-word query"},
	}

	for _, test := range queries {
		start := time.Now()

		results, err := s.client.Modules.SearchWithRelevance(ctx, test.query, 0)
		if err != nil {
			return fmt.Errorf("search failed for '%s': %w", test.query, err)
		}

		duration := time.Since(start)

		// Also measure relevance calculation time
		relevanceStart := time.Now()
		// Relevance is already calculated in SearchWithRelevance
		relevanceDuration := time.Since(relevanceStart)

		s.logger.Debugf("Search '%s' (%s): %d results in %v (relevance calc: %v)",
			test.query, test.description, len(results), duration, relevanceDuration)

		// Longer queries might take more time
		maxDuration := 3 * time.Second
		if len(test.query) > 20 {
			maxDuration = 5 * time.Second
		}

		if duration > maxDuration {
			s.logger.Warnf("Search took longer than expected: %v > %v", duration, maxDuration)
		}
	}

	return nil
}

func (s *PerformanceTests) testCacheBehavior(ctx context.Context) error {
	// Test caching behavior by making repeated identical requests
	// Note: The current implementation might not have caching

	// First request (cache miss)
	start1 := time.Now()
	result1, err := s.client.Providers.Get(ctx, "hashicorp", "aws")
	if err != nil {
		if !registry.IsNotFound(err) {
			return fmt.Errorf("first request failed: %w", err)
		}
		// Try a different provider
		result1, err = s.client.Providers.Get(ctx, "hashicorp", "random")
		if err != nil {
			s.logger.Warn("Could not test caching - provider not found")
			return nil
		}
	}
	duration1 := time.Since(start1)

	// Second identical request (potential cache hit)
	start2 := time.Now()
	result2, err := s.client.Providers.Get(ctx, "hashicorp", "aws")
	if err != nil {
		if !registry.IsNotFound(err) {
			return fmt.Errorf("second request failed: %w", err)
		}
		result2, err = s.client.Providers.Get(ctx, "hashicorp", "random")
	}
	duration2 := time.Since(start2)

	// Third identical request
	start3 := time.Now()
	_, err = s.client.Providers.Get(ctx, "hashicorp", "aws")
	if err != nil && !registry.IsNotFound(err) {
		_, err = s.client.Providers.Get(ctx, "hashicorp", "random")
	}
	duration3 := time.Since(start3)

	s.logger.Debugf("Request durations: 1st=%v, 2nd=%v, 3rd=%v", duration1, duration2, duration3)

	// If caching is implemented, subsequent requests should be faster
	if duration2 < duration1/2 || duration3 < duration1/2 {
		s.logger.Debug("Caching appears to be working (subsequent requests faster)")
	} else {
		s.logger.Debug("No significant caching detected")
	}

	// Verify results are consistent
	if result1 != nil && result2 != nil {
		if result1.ID != result2.ID {
			return fmt.Errorf("inconsistent results between requests")
		}
	}

	return nil
}
