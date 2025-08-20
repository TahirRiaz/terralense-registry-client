package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"terralense-registry-client/registry"
	"terralense-registry-client/tests"

	"github.com/sirupsen/logrus"
)

// Config holds the application configuration
type Config struct {
	Mode         string
	LogLevel     string
	Timeout      time.Duration
	BaseURL      string
	RateLimit    int
	RatePeriod   time.Duration
	OutputFormat string
	// Test-specific configurations
	TestSuite string
	TestCase  string
	ListTests bool
}

func main() {
	config := parseFlags()

	// Setup logger
	logger := setupLogger(config.LogLevel)

	// Handle list tests request
	if config.ListTests {
		listAvailableTests()
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Create client
	client, err := createClient(config, logger)
	if err != nil {
		log.Fatalf("Failed to create registry client: %v", err)
	}

	// Run based on mode
	switch config.Mode {
	case "demo":
		runDemo(ctx, client, logger)
	case "test":
		runTests(ctx, client, logger, config)
	case "all":
		runDemo(ctx, client, logger)
		fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
		runTests(ctx, client, logger, config)
	default:
		log.Fatalf("Unknown mode: %s", config.Mode)
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Mode, "mode", "demo", "Run mode: demo, test, or all")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level: debug, info, warn, error")
	flag.DurationVar(&config.Timeout, "timeout", 5*time.Minute, "Request timeout")
	flag.StringVar(&config.BaseURL, "base-url", registry.DefaultBaseURL, "Registry base URL")
	flag.IntVar(&config.RateLimit, "rate-limit", 100, "Rate limit requests per period")
	flag.DurationVar(&config.RatePeriod, "rate-period", time.Minute, "Rate limit period")
	flag.StringVar(&config.OutputFormat, "output", "table", "Output format: table, json, yaml")

	// Test-specific flags
	flag.StringVar(&config.TestSuite, "suite", "", "Run specific test suite (e.g., 'Modules', 'Providers')")
	flag.StringVar(&config.TestCase, "test", "", "Run specific test case (requires -suite)")
	flag.BoolVar(&config.ListTests, "list-tests", false, "List all available test suites and cases")

	flag.Parse()

	// Validate test-specific flags
	if config.TestCase != "" && config.TestSuite == "" {
		log.Fatal("Error: -test flag requires -suite flag to be specified")
	}

	return config
}

func setupLogger(level string) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// Set formatter
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "15:04:05",
		DisableQuote:     true,
	})

	return logger
}

func createClient(config *Config, logger *logrus.Logger) (*registry.Client, error) {
	return registry.NewClient(
		registry.WithBaseURL(config.BaseURL),
		registry.WithLogger(logger),
		registry.WithTimeout(30*time.Second),
		registry.WithRateLimit(config.RateLimit, config.RatePeriod),
		registry.WithUserAgent("terralense-registry-client/1.0"),
	)
}

func runDemo(ctx context.Context, client *registry.Client, logger *logrus.Logger) {
	fmt.Println("=== Terraform Registry Client Demo ===")
	fmt.Println("Running Azure VNet Resources Demo")
	fmt.Println(strings.Repeat("=", 50) + "\n")

	demo := NewAzureVNetDemo(client, logger)

	if err := demo.Run(ctx); err != nil {
		logger.Errorf("Demo failed: %v", err)
		os.Exit(1)
	}
}

func runTests(ctx context.Context, client *registry.Client, logger *logrus.Logger, config *Config) {
	fmt.Println("=== Terraform Registry Client Test Suite ===")

	// Create test runner
	runner := tests.NewTestRunner(client, logger)

	// Register all test suites
	allSuites := registerAllTestSuites(runner, client, logger)

	// Check if specific suite/test requested
	if config.TestSuite != "" {
		runSpecificTests(ctx, runner, allSuites, config)
		return
	}

	// Run all tests
	fmt.Println("Running comprehensive tests")
	fmt.Println(strings.Repeat("=", 50) + "\n")

	results := runner.RunAll(ctx)

	// Print results
	runner.PrintResults(results)

	// Exit with error if tests failed
	if results.Failed > 0 {
		os.Exit(1)
	}
}

func registerAllTestSuites(runner *tests.TestRunner, client *registry.Client, logger *logrus.Logger) map[string]tests.TestSuite {
	suites := make(map[string]tests.TestSuite)

	// Create all test suites
	suites["Modules"] = tests.NewModuleTests(client, logger)
	suites["Providers"] = tests.NewProviderTests(client, logger)
	suites["Policies"] = tests.NewPolicyTests(client, logger)
	suites["Search"] = tests.NewSearchTests(client, logger)
	suites["Validation"] = tests.NewValidationTests(client, logger)
	suites["Error Handling"] = tests.NewErrorTests(client, logger)
	suites["Performance"] = tests.NewPerformanceTests(client, logger)

	// Register with runner
	for name, suite := range suites {
		runner.AddSuite(name, suite)
	}

	return suites
}

func runSpecificTests(ctx context.Context, runner *tests.TestRunner, allSuites map[string]tests.TestSuite, config *Config) {
	// Find the requested suite
	suite, exists := allSuites[config.TestSuite]
	if !exists {
		fmt.Printf("Error: Test suite '%s' not found\n\n", config.TestSuite)
		fmt.Println("Available test suites:")
		for name := range allSuites {
			fmt.Printf("  - %s\n", name)
		}
		os.Exit(1)
	}

	// If specific test case requested
	if config.TestCase != "" {
		runSingleTest(ctx, runner, suite, config.TestSuite, config.TestCase)
		return
	}

	// Run all tests in the suite
	fmt.Printf("Running all tests in suite: %s\n", config.TestSuite)
	fmt.Println(strings.Repeat("=", 50) + "\n")

	results := runner.RunSuite(ctx, config.TestSuite, suite)
	runner.PrintResults(results)

	if results.Failed > 0 {
		os.Exit(1)
	}
}

func runSingleTest(ctx context.Context, runner *tests.TestRunner, suite tests.TestSuite, suiteName, testName string) {
	// Find the specific test
	var targetTest *tests.TestCase
	for _, test := range suite.Tests() {
		if test.Name == testName {
			targetTest = &test
			break
		}
	}

	if targetTest == nil {
		fmt.Printf("Error: Test case '%s' not found in suite '%s'\n\n", testName, suiteName)
		fmt.Printf("Available tests in %s suite:\n", suiteName)
		for _, test := range suite.Tests() {
			fmt.Printf("  - %s\n", test.Name)
		}
		os.Exit(1)
	}

	// Run the single test
	fmt.Printf("Running single test: %s/%s\n", suiteName, testName)
	fmt.Println(strings.Repeat("=", 50) + "\n")

	results := runner.RunSingleTest(ctx, suiteName, *targetTest)
	runner.PrintResults(results)

	if results.Failed > 0 {
		os.Exit(1)
	}
}

func listAvailableTests() {
	fmt.Println("=== Available Test Suites and Cases ===")
	fmt.Println()

	// Create a dummy client and logger just to get test suite info
	logger := logrus.New()
	client, _ := registry.NewClient(registry.WithLogger(logger))

	runner := tests.NewTestRunner(client, logger)
	allSuites := registerAllTestSuites(runner, client, logger)

	// List all suites and their tests
	for suiteName, suite := range allSuites {
		fmt.Printf("%s:\n", suiteName)
		for _, test := range suite.Tests() {
			fmt.Printf("  - %s", test.Name)
			if test.Description != "" {
				fmt.Printf(" (%s)", test.Description)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Print usage examples
	fmt.Println("Usage Examples:")
	fmt.Println("  # Run all tests")
	fmt.Println("  go run . -mode=test")
	fmt.Println()
	fmt.Println("  # Run all tests in a specific suite")
	fmt.Println("  go run . -mode=test -suite=\"Modules\"")
	fmt.Println()
	fmt.Println("  # Run a specific test")
	fmt.Println("  go run . -mode=test -suite=\"Modules\" -test=\"List Modules\"")
	fmt.Println()
	fmt.Println("  # Run with debug logging")
	fmt.Println("  go run . -mode=test -suite=\"Providers\" -log-level=debug")
}
