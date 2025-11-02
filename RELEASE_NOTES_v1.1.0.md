# Release Notes - v1.1.0

## 🎉 Major Feature: Subcategory Filtering & Resource Summaries

We're excited to announce v1.1.0 of the Terraform Registry Client, which adds powerful subcategory filtering capabilities and structured resource summaries!

### ✨ What's New

#### 1. Subcategory Filtering
Filter provider resources and data sources by subcategory (Networking, Compute, Storage, etc.):

```go
// Get all networking resources from Azure provider
versionID, _ := client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", "latest")
networkingResources, err := client.Providers.GetNetworkingResources(ctx, versionID)
```

**13 Predefined Subcategory Constants:**
- Networking (VPC, VNet, Subnets)
- Compute (VMs, EC2)
- Storage (S3, Blob Storage)
- Database (RDS, SQL)
- Security (IAM, Security Groups)
- Identity, Monitoring, Container, Serverless, Analytics, Messaging, Developer, Management

#### 2. Resource Summary API
Get a complete structured summary of all provider resources:

```go
summary, err := client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "aws", "latest")

// Access organized data
fmt.Printf("Total Resources: %d\n", summary.TotalResources)
fmt.Printf("Total Data Sources: %d\n", summary.TotalDataSources)

// Browse by subcategory
for _, subcategory := range summary.AllSubcategories {
    resources := summary.ResourcesBySubcategory[subcategory]
    for _, res := range resources {
        fmt.Printf("- %s (%s)\n", res.Title, res.Slug)
    }
}
```

#### 3. New Convenience Methods

**Get Resources by Subcategory:**
- `GetNetworkingResources()`
- `GetComputeResources()`
- `GetStorageResources()`
- `GetDatabaseResources()`
- `GetSecurityResources()`
- `GetResourcesBySubcategory()` (generic)
- `GetDataSourcesBySubcategory()`

**Get Resource Summaries:**
- `GetProviderResourceSummary()` - Complete structured summary
- `ExtractResourceInfoFromProviderDocs()` - Extract from existing docs
- `BuildResourceInfoFromDocs()` - Build simplified lists

### 📊 Use Cases

#### Use Case 1: Find All Networking Resources
```go
summary, _ := client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "aws", "latest")
networkResources := summary.ResourcesBySubcategory["VPC (Virtual Private Cloud)"]

for _, resource := range networkResources {
    fmt.Printf("Resource: %s\n", resource.Title)
    fmt.Printf("  Slug: %s\n", resource.Slug)
    fmt.Printf("  Path: %s\n", resource.Path)
}
```

#### Use Case 2: Compare Providers
```go
providers := []string{"aws", "azurerm", "google"}

for _, provider := range providers {
    summary, _ := client.Providers.GetProviderResourceSummary(ctx, "hashicorp", provider, "latest")
    networkingResources := summary.ResourcesBySubcategory[registry.SubcategoryNetworking]
    fmt.Printf("%s: %d networking resources\n", provider, len(networkingResources))
}
```

#### Use Case 3: Build Resource Catalog
```go
summary, _ := client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "aws", "latest")

// Export to JSON for your application
type Catalog struct {
    Provider     string              `json:"provider"`
    Version      string              `json:"version"`
    Subcategories map[string][]string `json:"subcategories"`
}

catalog := Catalog{
    Provider:     "aws",
    Version:      summary.Version,
    Subcategories: make(map[string][]string),
}

for subcategory, resources := range summary.ResourcesBySubcategory {
    for _, res := range resources {
        catalog.Subcategories[subcategory] = append(
            catalog.Subcategories[subcategory],
            res.Slug,
        )
    }
}
```

### 🔧 Technical Improvements

- **Production-Grade Code**: All new features include comprehensive validation, error handling, and documentation
- **Backward Compatible**: No breaking changes - all existing code continues to work
- **Well Tested**: 10 new test cases covering all subcategory functionality
- **Lenient Validation**: Supports both standard and custom provider subcategories
- **Performance Optimized**: Efficient API calls with proper pagination handling

### 📦 New Data Structures

**ProviderResourceSummary**
```go
type ProviderResourceSummary struct {
    ProviderNamespace        string
    ProviderName             string
    Version                  string
    TotalResources           int
    TotalDataSources         int
    ResourcesBySubcategory   map[string][]ResourceInfo
    DataSourcesBySubcategory map[string][]ResourceInfo
    AllSubcategories         []string
}
```

**ResourceInfo**
```go
type ResourceInfo struct {
    ID          string
    Type        string
    Name        string
    Title       string
    Subcategory string
    Category    string
    Slug        string
    Path        string
}
```

### 🧪 Testing

Run the new subcategory tests:

```bash
# Run all subcategory tests
go run ./cmd -mode=test -suite="Subcategory"

# Run specific test
go run ./cmd -mode=test -suite="Subcategory" -test="List Networking Resources"

# List all available tests
go run ./cmd -list-tests
```

### 📚 Documentation

- Updated [README.md](README.md) with comprehensive examples
- Added [subcategory_example.go](cmd/subcategory_example.go) with 4 complete examples
- Added [resource_summary_example.go](cmd/resource_summary_example.go) demonstrating summary usage
- New [CHANGELOG.md](CHANGELOG.md) for tracking changes

### 🚀 Getting Started

**Installation:**
```bash
go get github.com/TahirRiaz/terralense-registry-client/registry@v1.1.0
```

**Quick Example:**
```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/TahirRiaz/terralense-registry-client/registry"
)

func main() {
    client, _ := registry.NewClient()
    ctx := context.Background()

    // Get networking resources
    latest, _ := client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
    versionID, _ := client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latest.Version)

    resources, err := client.Providers.GetNetworkingResources(ctx, versionID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d networking resources\n", len(resources))
}
```

### 🔗 Links

- [Documentation](README.md)
- [Examples](cmd/)
- [Tests](tests/subcategory_tests.go)
- [Changelog](CHANGELOG.md)

### 🙏 Acknowledgments

Thank you to all contributors and users who requested this feature!

### 📝 License

MIT License - See [LICENSE](LICENSE) file for details

---

**Full Changelog**: v1.0.0...v1.1.0
