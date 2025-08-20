package tests

import (
	"context"
	"fmt"
	"strings"
	"time"

	"terralense-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// TestCase represents a single test case
type TestCase struct {
	Name        string
	Description string
	Run         func(ctx context.Context) error
}

// TestSuite represents a collection of related tests
type TestSuite interface {
	Name() string
	Tests() []TestCase
}

// TestResult represents the result of a single test
type TestResult struct {
	Suite    string
	Test     string
	Passed   bool
	Error    error
	Duration time.Duration
}

// TestResults aggregates all test results
type TestResults struct {
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Results  []TestResult
}

// TestRunner manages test execution
type TestRunner struct {
	client  *registry.Client
	logger  *logrus.Logger
	suites  map[string]TestSuite
	verbose bool
}

// NewTestRunner creates a new test runner
func NewTestRunner(client *registry.Client, logger *logrus.Logger) *TestRunner {
	return &TestRunner{
		client:  client,
		logger:  logger,
		suites:  make(map[string]TestSuite),
		verbose: logger.Level == logrus.DebugLevel,
	}
}

// AddSuite adds a test suite
func (r *TestRunner) AddSuite(name string, suite TestSuite) {
	r.suites[name] = suite
}

// GetSuite returns a specific test suite
func (r *TestRunner) GetSuite(name string) (TestSuite, bool) {
	suite, exists := r.suites[name]
	return suite, exists
}

// RunAll runs all test suites
func (r *TestRunner) RunAll(ctx context.Context) *TestResults {
	results := &TestResults{
		Results: make([]TestResult, 0),
	}

	startTime := time.Now()

	for _, suite := range r.suites {
		suiteResults := r.runSuite(ctx, suite)
		results.Results = append(results.Results, suiteResults...)
	}

	results.Duration = time.Since(startTime)

	// Calculate totals
	for _, result := range results.Results {
		results.Total++
		if result.Passed {
			results.Passed++
		} else {
			results.Failed++
		}
	}

	return results
}

// RunSuite runs a specific test suite and returns results
func (r *TestRunner) RunSuite(ctx context.Context, suiteName string, suite TestSuite) *TestResults {
	results := &TestResults{
		Results: make([]TestResult, 0),
	}

	startTime := time.Now()

	suiteResults := r.runSuite(ctx, suite)
	results.Results = append(results.Results, suiteResults...)

	results.Duration = time.Since(startTime)

	// Calculate totals
	for _, result := range results.Results {
		results.Total++
		if result.Passed {
			results.Passed++
		} else {
			results.Failed++
		}
	}

	return results
}

// RunSingleTest runs a single test case
func (r *TestRunner) RunSingleTest(ctx context.Context, suiteName string, test TestCase) *TestResults {
	results := &TestResults{
		Results: make([]TestResult, 0),
	}

	startTime := time.Now()

	result := r.runTest(ctx, suiteName, test)
	results.Results = append(results.Results, result)

	results.Duration = time.Since(startTime)
	results.Total = 1

	if result.Passed {
		results.Passed = 1
	} else {
		results.Failed = 1
	}

	// Print immediate result
	status := "✓ PASS"
	if !result.Passed {
		status = "✗ FAIL"
	}

	fmt.Printf("%s: %s/%s (%v)\n", status, suiteName, test.Name, result.Duration)

	if !result.Passed && result.Error != nil {
		fmt.Printf("  Error: %v\n", result.Error)
	}

	return results
}

// runSuite runs a single test suite
func (r *TestRunner) runSuite(ctx context.Context, suite TestSuite) []TestResult {
	r.logger.Infof("Running test suite: %s", suite.Name())
	fmt.Printf("\n%s Test Suite\n", suite.Name())
	fmt.Println(strings.Repeat("-", 50))

	var results []TestResult

	for _, test := range suite.Tests() {
		result := r.runTest(ctx, suite.Name(), test)
		results = append(results, result)

		// Print test result
		status := "✓ PASS"
		if !result.Passed {
			status = "✗ FAIL"
		}

		fmt.Printf("%s: %s (%v)\n", status, test.Name, result.Duration)

		if !result.Passed && result.Error != nil {
			fmt.Printf("  Error: %v\n", result.Error)
		}
	}

	return results
}

// runTest runs a single test
func (r *TestRunner) runTest(ctx context.Context, suiteName string, test TestCase) TestResult {
	result := TestResult{
		Suite: suiteName,
		Test:  test.Name,
	}

	// Create test context with timeout
	testCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	startTime := time.Now()

	// Run the test
	err := test.Run(testCtx)

	result.Duration = time.Since(startTime)
	result.Passed = err == nil
	result.Error = err

	if r.verbose {
		if result.Passed {
			r.logger.Debugf("Test passed: %s/%s", suiteName, test.Name)
		} else {
			r.logger.Errorf("Test failed: %s/%s - %v", suiteName, test.Name, err)
		}
	}

	return result
}

// PrintResults prints test results in a formatted way
func (r *TestRunner) PrintResults(results *TestResults) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Test Results Summary")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("Total Tests:    %d\n", results.Total)

	if results.Total > 0 {
		fmt.Printf("Passed:         %d (%.1f%%)\n", results.Passed, float64(results.Passed)/float64(results.Total)*100)
		fmt.Printf("Failed:         %d (%.1f%%)\n", results.Failed, float64(results.Failed)/float64(results.Total)*100)
	} else {
		fmt.Printf("Passed:         %d\n", results.Passed)
		fmt.Printf("Failed:         %d\n", results.Failed)
	}

	fmt.Printf("Total Duration: %v\n", results.Duration)

	if results.Failed > 0 {
		fmt.Println("\nFailed Tests:")
		fmt.Println(strings.Repeat("-", 30))

		for _, result := range results.Results {
			if !result.Passed {
				fmt.Printf("  • %s/%s\n", result.Suite, result.Test)
				if result.Error != nil {
					fmt.Printf("    Error: %v\n", result.Error)
				}
			}
		}
	}

	fmt.Println()
}

// ListSuites returns a list of all registered test suites
func (r *TestRunner) ListSuites() []string {
	suites := make([]string, 0, len(r.suites))
	for name := range r.suites {
		suites = append(suites, name)
	}
	return suites
}

// BaseTestSuite provides common functionality for test suites
type BaseTestSuite struct {
	client *registry.Client
	logger *logrus.Logger
	name   string
	tests  []TestCase
}

// NewBaseTestSuite creates a new base test suite
func NewBaseTestSuite(name string, client *registry.Client, logger *logrus.Logger) *BaseTestSuite {
	return &BaseTestSuite{
		name:   name,
		client: client,
		logger: logger,
		tests:  make([]TestCase, 0),
	}
}

// Name returns the suite name
func (s *BaseTestSuite) Name() string {
	return s.name
}

// Tests returns all test cases
func (s *BaseTestSuite) Tests() []TestCase {
	return s.tests
}

// AddTest adds a test case to the suite
func (s *BaseTestSuite) AddTest(name, description string, testFunc func(ctx context.Context) error) {
	s.tests = append(s.tests, TestCase{
		Name:        name,
		Description: description,
		Run:         testFunc,
	})
}

// AssertEqual checks if two values are equal
func AssertEqual(expected, actual interface{}) error {
	if expected != actual {
		return fmt.Errorf("expected %v, got %v", expected, actual)
	}
	return nil
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(value interface{}) error {
	if value == nil {
		return fmt.Errorf("expected non-nil value, got nil")
	}
	return nil
}

// AssertNil checks if a value is nil
func AssertNil(value interface{}) error {
	if value != nil {
		return fmt.Errorf("expected nil, got %v", value)
	}
	return nil
}

// AssertTrue checks if a condition is true
func AssertTrue(condition bool, message string) error {
	if !condition {
		return fmt.Errorf("assertion failed: %s", message)
	}
	return nil
}

// AssertNoError checks if there is no error
func AssertNoError(err error) error {
	if err != nil {
		return fmt.Errorf("expected no error, got: %v", err)
	}
	return nil
}

// AssertError checks if there is an error
func AssertError(err error) error {
	if err == nil {
		return fmt.Errorf("expected error, got nil")
	}
	return nil
}

// AssertContains checks if a string contains a substring
func AssertContains(haystack, needle string) error {
	if !strings.Contains(haystack, needle) {
		return fmt.Errorf("expected '%s' to contain '%s'", haystack, needle)
	}
	return nil
}

// AssertGreaterThan checks if a > b
func AssertGreaterThan(a, b int) error {
	if a <= b {
		return fmt.Errorf("expected %d > %d", a, b)
	}
	return nil
}

// AssertLessThan checks if a < b
func AssertLessThan(a, b int) error {
	if a >= b {
		return fmt.Errorf("expected %d < %d", a, b)
	}
	return nil
}
