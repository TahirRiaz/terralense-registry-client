# Terraform Registry Client Testing Guide

This comprehensive guide covers all aspects of testing the Terraform Registry Client library, including test structure, running specific tests, and best practices.

## Table of Contents

- [Quick Start](#quick-start)
- [Command Line Options](#command-line-options)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Test Suites Overview](#test-suites-overview)
- [Writing New Tests](#writing-new-tests)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Quick Start

```bash
# List all available tests
go run ./cmd -list-tests

# Run all tests
go run ./cmd -mode=test

# Run specific test suite
go run ./cmd -mode=test -suite="Modules"

# Run single test
go run ./cmd -mode=test -suite="Modules" -test="List Modules"

# Run with debug logging
go run ./cmd -mode=test -suite="Providers" -log-level=debug
```

## Command Line Options

### General Flags

- `-mode`: Run mode (demo, test, or all)
- `-log-level`: Log level (debug, info, warn, error)
- `-timeout`: Request timeout duration (default: 5m)
- `-base-url`: Registry base URL (default: https://registry.terraform.io)
- `-rate-limit`: Rate limit requests per period (default: 100)
- `-rate-period`: Rate limit period duration (default: 1m)
- `-output`: Output format: table, json, yaml (default: table)

### Test-Specific Flags

- `-suite="SuiteName"`: Run all tests in a specific suite
- `-test="TestName"`: Run a specific test case (requires -suite)
- `-list-tests`: List all available test suites and cases

## Test Structure

The test suite is organized into seven main categories:

```
tests/
├── test_runner.go      # Core test execution framework
├── module_tests.go     # Module API tests
├── provider_tests.go   # Provider API tests
├── policy_tests.go     # Policy API tests
├── search_tests.go     # Search functionality tests
├── validation_tests.go # Input validation tests
├── error_tests.go      # Error handling tests
└── performance_tests.go # Performance benchmarks
```

## Running Tests

### List All Available Tests

```bash
go run ./cmd -list-tests
```

Output example:
```
=== Available Test Suites and Cases ===

Modules:
  - List Modules
  - Search Modules
  - Search with Relevance
  - Get Module
  - Get Module by ID
  - Get Latest Version
  - List Versions
  - Download URL
  - Invalid Module

Providers:
  - List Providers
  - Get Provider
  - Get Latest Version
  - List Versions
  - Get Version ID
  - List Provider Docs
  - Get Provider Docs V2
  - Invalid Provider

[... more suites ...]
```

### Run All Tests

```bash
# Run all test suites
go run ./cmd -mode=test

# Run demo and all tests
go run ./cmd -mode=all
```

### Run Specific Test Suite

```bash
# Run all module tests
go run ./cmd -mode=test -suite="Modules"

# Run all provider tests
go run ./cmd -mode=test -suite="Providers"

# Run all search tests
go run ./cmd -mode=test -suite="Search"
```

### Run Single Test

```bash
# Run specific test with exact name from -list-tests output
go run ./cmd -mode=test -suite="Modules" -test="List Modules"

# Run with debug logging
go run ./cmd -mode=test -suite="Providers" -test="Get Provider" -log-level=debug

# Run with custom timeout
go run ./cmd -mode=test -suite="Performance" -test="Large Result Set" -timeout=10m
```

### Build and Run

```bash
# Build the binary
go build -o terralense-client ./cmd

# Run tests using binary
./terralense-client -mode=test -suite="Modules"
./terralense-client -list-tests
```

## Test Suites Overview

### 1. Module Tests (`module_tests.go`)

Tests for the Modules API functionality:

| Test Name | Description |
|-----------|-------------|
| List Modules | Tests listing modules with pagination |
| Search Modules | Tests basic module search |
| Search with Relevance | Tests relevance scoring in search |
| Get Module | Tests getting specific module details |
| Get Module by ID | Tests getting module using full ID |
| Get Latest Version | Tests retrieving latest module version |
| List Versions | Tests listing all module versions |
| Download URL | Tests download URL generation |
| Module Pagination | Tests pagination parameters |
| Provider Filter | Tests filtering by provider |
| Verified Filter | Tests filtering verified modules |
| Invalid Module | Tests error handling for invalid modules |

### 2. Provider Tests (`provider_tests.go`)

Tests for the Providers API functionality:

| Test Name | Description |
|-----------|-------------|
| List Providers | Tests listing providers |
| Get Provider | Tests getting provider details |
| Get Latest Version | Tests getting latest provider version |
| List Versions | Tests listing all provider versions |
| Get Version ID | Tests version ID retrieval |
| List Provider Docs | Tests documentation listing (v1 API) |
| Get Provider Docs V2 | Tests documentation (v2 API) |
| Filter by Tier | Tests tier filtering (official/partner/community) |
| Filter by Namespace | Tests namespace filtering |
| Invalid Provider | Tests error handling |

### 3. Policy Tests (`policy_tests.go`)

Tests for the Policies API functionality:

| Test Name | Description |
|-----------|-------------|
| List Policies | Tests listing policies |
| Get Policy | Tests getting policy details |
| Get Policy by ID | Tests getting policy using ID |
| Search Policies | Tests policy search |
| Generate Sentinel | Tests Sentinel content generation |
| Policy Pagination | Tests pagination |
| Include Latest Version | Tests version inclusion |
| Invalid Policy | Tests error handling |

### 4. Search Tests (`search_tests.go`)

Tests for search functionality:

| Test Name | Description |
|-----------|-------------|
| Module Search Relevance | Tests relevance scoring for modules |
| Policy Search Relevance | Tests relevance scoring for policies |
| Cross Provider Search | Tests searching across providers |
| Empty Query | Tests empty search query handling |
| Special Characters | Tests special character handling |
| Case Sensitivity | Tests case-insensitive search |
| Partial Matching | Tests partial word matching |
| Multi-word Search | Tests multi-word queries |

### 5. Validation Tests (`validation_tests.go`)

Tests for input validation:

| Test Name | Description |
|-----------|-------------|
| Module Parameters | Tests module parameter validation |
| Provider Parameters | Tests provider parameter validation |
| Policy Parameters | Tests policy parameter validation |
| Version Validation | Tests version string validation |
| Pagination Limits | Tests pagination limit validation |
| Parse Module ID | Tests module ID parsing |
| Parse Policy ID | Tests policy ID parsing |
| Provider URI Parsing | Tests provider URI formats |

### 6. Error Handling Tests (`error_tests.go`)

Tests for error handling:

| Test Name | Description |
|-----------|-------------|
| Not Found Errors | Tests 404 error handling |
| Validation Errors | Tests validation error handling |
| Error Type Checking | Tests error type functions |
| Context Cancellation | Tests context cancellation |
| Timeout Handling | Tests timeout scenarios |
| API Error Structure | Tests API error parsing |
| Multi-error Handling | Tests aggregated errors |

### 7. Performance Tests (`performance_tests.go`)

Tests for performance characteristics:

| Test Name | Description |
|-----------|-------------|
| Response Times | Tests API response times |
| Concurrent Requests | Tests concurrent API calls |
| Rate Limiting | Tests rate limiter behavior |
| Large Result Set | Tests handling large responses |
| Pagination Performance | Tests pagination efficiency |
| Search Performance | Tests search speed |

## Writing New Tests

### 1. Create Test Suite

```go
package tests

import (
    "context"
    "terralense-registry-client/registry"
    "github.com/sirupsen/logrus"
)

type MyNewTests struct {
    *BaseTestSuite
}

func NewMyNewTests(client *registry.Client, logger *logrus.Logger) TestSuite {
    suite := &MyNewTests{
        BaseTestSuite: NewBaseTestSuite("My New Tests", client, logger),
    }
    
    // Add test cases
    suite.AddTest("Test Name", "Test description", suite.testFunction)
    
    return suite
}

func (s *MyNewTests) testFunction(ctx context.Context) error {
    // Test implementation
    result, err := s.client.SomeAPI.SomeMethod(ctx, params)
    if err != nil {
        return err
    }
    
    // Use assertions
    return AssertNotNil(result)
}
```

### 2. Register Test Suite

In `main.go`, add your test suite to the registration:

```go
func registerAllTestSuites(runner *tests.TestRunner, client *registry.Client, logger *logrus.Logger) map[string]tests.TestSuite {
    suites := make(map[string]tests.TestSuite)
    
    // Existing suites...
    suites["My New Tests"] = tests.NewMyNewTests(client, logger)
    
    // Register with runner
    for name, suite := range suites {
        runner.AddSuite(name, suite)
    }
    
    return suites
}
```

### 3. Use Assertion Helpers

```go
// Available assertions
AssertEqual(expected, actual)
AssertNotNil(value)
AssertNil(value)
AssertTrue(condition, message)
AssertNoError(err)
AssertError(err)
AssertContains(haystack, needle)
AssertGreaterThan(a, b)
AssertLessThan(a, b)
```

## Best Practices

### 1. Test Isolation

Each test should be independent:

```go
func (s *MyTests) testSomething(ctx context.Context) error {
    // Don't rely on state from other tests
    // Set up everything needed for this test
    
    // Clean up if necessary
    defer cleanup()
    
    // Perform test
    return nil
}
```

### 2. Error Handling

Always check and handle errors appropriately:

```go
result, err := s.client.Modules.List(ctx, opts)
if err != nil {
    // Check if it's expected error
    if registry.IsNotFound(err) {
        return nil // Expected
    }
    return fmt.Errorf("unexpected error: %w", err)
}
```

### 3. Context and Timeouts

Use context with appropriate timeouts:

```go
func (s *MyTests) testWithTimeout(ctx context.Context) error {
    // Create a shorter timeout for specific operation
    shortCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return s.client.SomeAPI.QuickOperation(shortCtx, params)
}
```

### 4. Logging

Provide helpful debug information:

```go
s.logger.Debugf("Testing with params: %+v", params)
result, err := s.client.SomeAPI.Method(ctx, params)
s.logger.Debugf("Got result: %+v, error: %v", result, err)
```

### 5. Rate Limiting

Be mindful of API rate limits:

```go
func (s *PerformanceTests) testConcurrent(ctx context.Context) error {
    // Add delays between requests if needed
    time.Sleep(100 * time.Millisecond)
    
    // Or check rate limiter status
    remaining := s.client.GetRateLimiter().TokensRemaining()
    if remaining < 10 {
        s.logger.Warnf("Low rate limit tokens: %d", remaining)
    }
}
```

## Troubleshooting

### Common Issues

#### 1. Rate Limiting Errors

```bash
# Reduce rate limit for testing
go run ./cmd -mode=test -rate-limit=10 -rate-period=1m
```

#### 2. Network Timeouts

```bash
# Increase timeout
go run ./cmd -mode=test -timeout=10m
```

#### 3. Test Data Dependencies

Some tests expect specific modules/providers to exist. If tests fail:

- Check if the expected resources still exist in the registry
- Update test data references
- Use more generic search terms

#### 4. Debug Failed Tests

```bash
# Run with debug logging
go run ./cmd -mode=test -suite="Modules" -test="Failed Test" -log-level=debug

# Check specific error
go run ./cmd -mode=test -suite="Error Handling" -log-level=debug
```

### Environment Variables

Set these for additional control:

```bash
# HTTP proxy
export HTTP_PROXY=http://proxy.example.com:8080

# Custom registry URL (for testing against different registries)
export REGISTRY_URL=https://custom.registry.io

# Disable color output
export NO_COLOR=1
```

### Exit Codes

- `0`: All tests passed
- `1`: One or more tests failed or invalid arguments

### Verbose Output

For maximum debugging information:

```bash
go run ./cmd -mode=test -log-level=debug 2>&1 | tee test-output.log
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run all tests
      run: go run ./cmd -mode=test
    
    - name: Run specific suite
      run: go run ./cmd -mode=test -suite="Modules"
```

### Tips for CI

1. Use appropriate timeouts
2. Handle rate limiting gracefully
3. Don't depend on specific registry data
4. Log enough information for debugging
5. Consider running different suites in parallel jobs

## Summary

This testing framework provides:

- Flexible test execution (all, suite, or single test)
- Comprehensive test coverage
- Easy debugging with detailed logging
- Performance benchmarking
- Proper error handling validation
- CI/CD ready implementation

Use `-list-tests` to explore available tests and `-log-level=debug` when troubleshooting issues.