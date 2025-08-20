# Terraform Registry Client

A production-grade Go client library for interacting with the Terraform Registry API.

## Features

- **Complete API Coverage**: Support for Providers, Modules, and Policies
- **Production Ready**: Comprehensive error handling, rate limiting, and retry logic
- **Type Safe**: Strongly typed responses with full struct definitions
- **Configurable**: Flexible client configuration with sensible defaults
- **Well Tested**: Interfaces for easy mocking and testing
- **Performant**: Connection pooling, efficient pagination, and request optimization
- **Observable**: Built-in logging with configurable levels

## Installation

```bash
go get github.com/yourusername/terralense-registry-client
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/yourusername/terralense-registry-client/registry"
)

func main() {
    // Create a client with default configuration
    client, err := registry.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Search for modules
    results, err := client.Modules.Search(ctx, "aws vpc", 0)
    if err != nil {
        log.Fatal(err)
    }

    // Print results
    for _, module := range results.Modules {
        fmt.Printf("Module: %s/%s/%s (v%s)\n", 
            module.Namespace, module.Name, module.Provider, module.Version)
    }
}
```

## Configuration

The client supports various configuration options:

```go
client, err := registry.NewClient(
    registry.WithBaseURL("https://custom-registry.example.com"),
    registry.WithTimeout(60 * time.Second),
    registry.WithLogger(customLogger),
    registry.WithRateLimit(200, time.Minute),
    registry.WithUserAgent("my-app/1.0"),
    registry.WithAPIToken("your-token"), // For private registries
)
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithBaseURL` | Registry API base URL | `https://registry.terraform.io` |
| `WithTimeout` | HTTP request timeout | `30s` |
| `WithLogger` | Custom logger instance | `logrus.New()` |
| `WithRateLimit` | Rate limit (requests, period) | `100 requests/minute` |
| `WithUserAgent` | Custom User-Agent header | `terraform-registry-client/1.0` |
| `WithAPIToken` | API token for authentication | None |
| `WithHTTPClient` | Custom HTTP client | Auto-configured |

## API Usage

### Modules

```go
// List all modules with pagination
opts := &registry.ModuleListOptions{
    Offset:   0,
    Limit:    50,
    Provider: "aws",
    Verified: true,
}
modules, err := client.Modules.List(ctx, opts)

// Search for modules
results, err := client.Modules.Search(ctx, "kubernetes", 0)

// Search with relevance scoring
searchResults, err := client.Modules.SearchWithRelevance(ctx, "azure network", 0)

// Get specific module
module, err := client.Modules.Get(ctx, "hashicorp", "consul", "aws", "0.1.0")

// Get latest version
latest, err := client.Modules.GetLatest(ctx, "hashicorp", "consul", "aws")

// List all versions
versions, err := client.Modules.ListVersions(ctx, "hashicorp", "consul", "aws")

// Get download URL
downloadURL, err := client.Modules.Download(ctx, "hashicorp", "consul", "aws", "0.1.0")
```

### Providers

```go
// List providers
opts := &registry.ProviderListOptions{
    Tier:      "official",
    Namespace: "hashicorp",
    Page:      1,
    PageSize:  50,
}
providers, err := client.Providers.List(ctx, opts)

// Get provider details
provider, err := client.Providers.Get(ctx, "hashicorp", "aws")

// Get latest version
latest, err := client.Providers.GetLatest(ctx, "hashicorp", "aws")

// List versions
versions, err := client.Providers.ListVersions(ctx, "hashicorp", "aws")

// Get documentation
docs, err := client.Providers.ListDocs(ctx, "hashicorp", "aws", "4.0.0")

// Get specific documentation with v2 API
docOpts := &registry.ProviderDocListOptions{
    ProviderVersionID: versionID,
    Category:         "resources",
    Slug:            "instance",
    Language:        "hcl",
}
docs, err := client.Providers.ListDocsV2(ctx, docOpts)
```

### Policies

```go
// List policies
opts := &registry.PolicyListOptions{
    PageSize:             50,
    Page:                1,
    IncludeLatestVersion: true,
}
policies, err := client.Policies.List(ctx, opts)

// Search policies
results, err := client.Policies.Search(ctx, "aws compliance")

// Get policy details
policy, err := client.Policies.Get(ctx, "hashicorp", "aws-cis-framework", "1.0.0")

// Generate Sentinel configuration
content, err := client.Policies.GetSentinelContent(ctx, policyID)
hcl := content.GenerateHCL("soft-mandatory")
```

## Error Handling

The client provides comprehensive error handling with typed errors:

```go
module, err := client.Modules.Get(ctx, "invalid", "module", "provider", "version")
if err != nil {
    if registry.IsNotFound(err) {
        // Handle 404
    } else if registry.IsRateLimited(err) {
        // Handle rate limiting
    } else if registry.IsValidationError(err) {
        // Handle validation errors
    } else {
        // Handle other errors
    }
}
```

### Error Types

- `APIError`: API response errors with status codes
- `RequestError`: Request creation/sending errors
- `ResponseError`: Response processing errors
- `ValidationError`: Input validation errors
- `MultiError`: Multiple errors combined

### Error Helper Functions

- `IsNotFound(err)`: Check if error is 404
- `IsRateLimited(err)`: Check if error is 429
- `IsUnauthorized(err)`: Check if error is 401
- `IsForbidden(err)`: Check if error is 403
- `IsServerError(err)`: Check if error is 5xx
- `IsValidationError(err)`: Check if error is validation related

## Advanced Features

### Rate Limiting

The client includes built-in rate limiting to prevent hitting API limits:

```go
// Configure custom rate limits
client, err := registry.NewClient(
    registry.WithRateLimit(200, time.Minute), // 200 requests per minute
)
```

### Pagination

For endpoints that support pagination:

```go
var allModules []registry.Module
offset := 0
limit := 100

for {
    opts := &registry.ModuleListOptions{
        Offset: offset,
        Limit:  limit,
    }
    
    result, err := client.Modules.List(ctx, opts)
    if err != nil {
        return err
    }
    
    allModules = append(allModules, result.Modules...)
    
    if len(result.Modules) < limit {
        break // No more results
    }
    
    offset += limit
}
```

### Context Support

All API methods accept a context for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

module, err := client.Modules.Get(ctx, namespace, name, provider, version)
```

### Logging

The client uses logrus for structured logging:

```go
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
logger.SetFormatter(&logrus.JSONFormatter{})

client, err := registry.NewClient(
    registry.WithLogger(logger),
)
```

## Testing

The client provides interfaces for all services, making it easy to mock in tests:

```go
type mockModulesService struct {
    mock.Mock
}

func (m *mockModulesService) Search(ctx context.Context, query string, offset int) (*registry.ModuleList, error) {
    args := m.Called(ctx, query, offset)
    return args.Get(0).(*registry.ModuleList), args.Error(1)
}

// In your test
mockService := new(mockModulesService)
mockService.On("Search", mock.Anything, "test", 0).Return(&registry.ModuleList{}, nil)
```

## Examples

See the [examples](./examples) directory for more comprehensive examples:

- [Basic Usage](./examples/basic/main.go)
- [Advanced Search](./examples/search/main.go)
- [Provider Documentation](./examples/providers/main.go)
- [Policy Management](./examples/policies/main.go)

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built using [go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) for robust HTTP handling
- Uses [logrus](https://github.com/sirupsen/logrus) for structured logging