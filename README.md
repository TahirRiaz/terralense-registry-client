# Terraform Registry Client

[![Go Reference](https://pkg.go.dev/badge/github.com/TahirRiaz/terralense-registry-client.svg)](https://pkg.go.dev/github.com/TahirRiaz/terralense-registry-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/TahirRiaz/terralense-registry-client)](https://goreportcard.com/report/github.com/TahirRiaz/terralense-registry-client)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A robust Go client library for interacting with the Terraform Registry API. This library provides comprehensive access to modules, providers, and policies with built-in rate limiting, retries, and error handling.

## Features

- 🚀 **Full API Coverage** - Complete implementation of Terraform Registry v1 and v2 APIs
- 🔄 **Automatic Retries** - Built-in retry logic with exponential backoff
- ⚡ **Rate Limiting** - Configurable rate limiter to respect API limits
- 🔍 **Advanced Search** - Search with relevance scoring for better results
- 📚 **Provider Documentation** - Access provider documentation and schemas
- 🛡️ **Type Safety** - Fully typed responses with Go structs
- 🧪 **Well Tested** - Comprehensive test suite with real API integration tests

## Installation

```bash
go get github.com/TahirRiaz/terralense-registry-client/registry
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/TahirRiaz/terralense-registry-client/registry"
)

func main() {
    // Create a client with default settings
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
    
    for _, module := range results.Modules {
        fmt.Printf("Found: %s/%s/%s (v%s) - %d downloads\n",
            module.Namespace, module.Name, module.Provider,
            module.Version, module.Downloads)
    }
}
```

## Configuration

```go
// Create a client with custom configuration
client, err := registry.NewClient(
    registry.WithBaseURL("https://registry.terraform.io"),
    registry.WithTimeout(30 * time.Second),
    registry.WithRateLimit(100, time.Minute),
    registry.WithLogger(logrus.New()),
    registry.WithUserAgent("my-app/1.0"),
)
```

## API Usage

### Modules

```go
// List modules with pagination
modules, err := client.Modules.List(ctx, &registry.ModuleListOptions{
    Offset:   0,
    Limit:    20,
    Provider: "aws",
    Verified: true,
})

// Get specific module details
module, err := client.Modules.Get(ctx, "terraform-aws-modules", "vpc", "aws", "5.0.0")

// Get latest version
latest, err := client.Modules.GetLatest(ctx, "terraform-aws-modules", "vpc", "aws")

// Search with relevance scoring
results, err := client.Modules.SearchWithRelevance(ctx, "kubernetes ingress", 0)
```

### Providers

```go
// List providers
providers, err := client.Providers.List(ctx, &registry.ProviderListOptions{
    Tier:     "official",
    PageSize: 50,
})

// Get provider details
provider, err := client.Providers.Get(ctx, "hashicorp", "aws")

// Get provider documentation
docs, err := client.Providers.ListDocs(ctx, "hashicorp", "aws", "5.0.0")

// Get resources by subcategory (NEW!)
latest, _ := client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
versionID, _ := client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latest.Version)

// Method 1: Use convenience methods
networkingResources, err := client.Providers.GetNetworkingResources(ctx, versionID)
computeResources, err := client.Providers.GetComputeResources(ctx, versionID)
storageResources, err := client.Providers.GetStorageResources(ctx, versionID)
databaseResources, err := client.Providers.GetDatabaseResources(ctx, versionID)
securityResources, err := client.Providers.GetSecurityResources(ctx, versionID)

// Method 2: Use generic method with subcategory constant
resources, err := client.Providers.GetResourcesBySubcategory(ctx, versionID, registry.SubcategoryNetworking)

// Method 3: Get data sources by subcategory
dataSources, err := client.Providers.GetDataSourcesBySubcategory(ctx, versionID, registry.SubcategoryNetworking)

// Method 4: Use ListDocsV2 for full control
opts := &registry.ProviderDocListOptions{
    ProviderVersionID: versionID,
    Category:          "resources",
    Subcategory:       registry.SubcategoryNetworking,
    Language:          "hcl",
}
docs, err := client.Providers.ListDocsV2(ctx, opts)

// Method 5: Get a complete structured summary (NEW!)
summary, err := client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "aws", "latest")
// Returns:
// - summary.TotalResources
// - summary.TotalDataSources
// - summary.ResourcesBySubcategory (map[string][]ResourceInfo)
// - summary.DataSourcesBySubcategory (map[string][]ResourceInfo)
// - summary.AllSubcategories (sorted list)
```

#### Available Subcategory Constants

```go
registry.SubcategoryNetworking   // Networking resources (VPC, VNet, Subnets, etc.)
registry.SubcategoryCompute      // Compute resources (VMs, EC2, etc.)
registry.SubcategoryStorage      // Storage resources (S3, Blob Storage, etc.)
registry.SubcategoryDatabase     // Database resources (RDS, SQL Database, etc.)
registry.SubcategorySecurity     // Security resources (IAM, Security Groups, etc.)
registry.SubcategoryIdentity     // Identity and access management
registry.SubcategoryMonitoring   // Monitoring and logging resources
registry.SubcategoryContainer    // Container resources (ECS, AKS, etc.)
registry.SubcategoryServerless   // Serverless resources (Lambda, Functions, etc.)
registry.SubcategoryAnalytics    // Analytics resources
registry.SubcategoryMessaging    // Messaging and queueing resources
registry.SubcategoryDeveloper    // Developer tools and resources
registry.SubcategoryManagement   // Management and governance resources
```

### Policies

```go
// Search policies
policies, err := client.Policies.Search(ctx, "aws compliance")

// Get policy details
policy, err := client.Policies.Get(ctx, "hashicorp", "cis-policy", "1.0.0")

// Generate Sentinel configuration
content, err := client.Policies.GetSentinelContent(ctx, policyID)
hcl := content.GenerateHCL("soft-mandatory")
```

## Error Handling

The library provides typed errors with helper functions:

```go
module, err := client.Modules.Get(ctx, namespace, name, provider, version)
if err != nil {
    switch {
    case registry.IsNotFound(err):
        // Handle 404
    case registry.IsRateLimited(err):
        // Handle rate limiting
    case registry.IsValidationError(err):
        // Handle validation errors
    default:
        // Handle other errors
    }
}
```

## Examples

Check the `tests` directory for comprehensive examples:

- [Module operations](tests/module_tests.go)
- [Provider operations](tests/provider_tests.go)
- [Search functionality](tests/search_tests.go)
- [Error handling](tests/error_tests.go)

## Running Tests

```bash
# Run all tests
go run ./cmd -mode=test

# Run specific test suite
go run ./cmd -mode=test -suite="Modules"

# List available tests
go run ./cmd -list-tests
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) for robust HTTP handling
- Uses [logrus](https://github.com/sirupsen/logrus) for structured logging