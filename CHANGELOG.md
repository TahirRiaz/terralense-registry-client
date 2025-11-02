# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-11-02

### Added

#### Subcategory Filtering Support
- **NEW**: Added comprehensive subcategory filtering for provider resources and data sources
- Added `Subcategory` field to `ProviderDocListOptions` struct for filtering by resource subcategory
- Added 13 predefined subcategory constants:
  - `SubcategoryNetworking` - Networking resources (VPC, VNet, Subnets, etc.)
  - `SubcategoryCompute` - Compute resources (VMs, EC2, etc.)
  - `SubcategoryStorage` - Storage resources (S3, Blob Storage, etc.)
  - `SubcategoryDatabase` - Database resources (RDS, SQL Database, etc.)
  - `SubcategorySecurity` - Security resources (IAM, Security Groups, etc.)
  - `SubcategoryIdentity` - Identity and access management resources
  - `SubcategoryMonitoring` - Monitoring and logging resources
  - `SubcategoryContainer` - Container resources (ECS, AKS, etc.)
  - `SubcategoryServerless` - Serverless resources (Lambda, Functions, etc.)
  - `SubcategoryAnalytics` - Analytics resources
  - `SubcategoryMessaging` - Messaging and queueing resources
  - `SubcategoryDeveloper` - Developer tools and resources
  - `SubcategoryManagement` - Management and governance resources

#### New Convenience Methods
- `GetNetworkingResources(ctx, providerVersionID)` - Get all networking resources
- `GetComputeResources(ctx, providerVersionID)` - Get all compute resources
- `GetStorageResources(ctx, providerVersionID)` - Get all storage resources
- `GetDatabaseResources(ctx, providerVersionID)` - Get all database resources
- `GetSecurityResources(ctx, providerVersionID)` - Get all security resources
- `GetResourcesBySubcategory(ctx, providerVersionID, subcategory)` - Generic method for any subcategory
- `GetDataSourcesBySubcategory(ctx, providerVersionID, subcategory)` - Get data sources by subcategory

#### Resource Summary Feature
- **NEW**: Added `GetProviderResourceSummary(ctx, namespace, name, version)` method
  - Returns a complete structured summary of all provider resources and data sources
  - Organizes resources by subcategory
  - Provides total counts and sorted subcategory lists
  - Returns `ProviderResourceSummary` struct with:
    - `TotalResources` - Total count of resources
    - `TotalDataSources` - Total count of data sources
    - `ResourcesBySubcategory` - Map of subcategory to resource list
    - `DataSourcesBySubcategory` - Map of subcategory to data source list
    - `AllSubcategories` - Sorted list of all subcategories

#### New Data Structures
- `ProviderResourceSummary` - Structured summary of provider resources organized by subcategory
- `ResourceInfo` - Lightweight resource information structure containing:
  - ID, Type, Name, Title
  - Subcategory, Category
  - Slug, Path

#### Helper Functions
- `ExtractResourceInfoFromProviderDocs(docs)` - Extract key info from provider documentation
- `BuildResourceInfoFromDocs(docs)` - Create simplified resource lists
- Subcategory validation with lenient handling of custom provider subcategories

### Enhanced

#### API Support
- Updated `ListDocsV2` to support subcategory filtering via query parameters
- Enhanced provider documentation filtering with `filter[subcategory]` parameter
- Improved validation for subcategory fields (lenient to support custom provider subcategories)

#### Testing
- Added comprehensive subcategory test suite with 10 test cases:
  - List Networking Resources
  - List Compute Resources
  - List Storage Resources
  - List Database Resources
  - List Security Resources
  - List Resources by Custom Subcategory
  - List Data Sources by Subcategory
  - Validate Subcategory Filtering
  - Test Subcategory Validation
  - Test Multiple Providers
- Integrated subcategory tests into main test runner

#### Documentation
- Updated README.md with subcategory filtering examples
- Added documentation for all 13 subcategory constants
- Included 4 different methods to filter resources by subcategory
- Created example files:
  - `subcategory_example.go` - Demonstrates subcategory filtering
  - `resource_summary_example.go` - Shows resource summary usage

#### Interfaces
- Updated `ProvidersServiceInterface` with new subcategory methods
- Maintained backward compatibility with existing API

### Technical Details

#### Files Modified
- `registry/providers.go` - Added subcategory filtering logic and new methods
- `registry/types.go` - Added new data structures (ProviderResourceSummary, ResourceInfo)
- `registry/interfaces.go` - Updated interface definitions
- `README.md` - Enhanced documentation with examples

#### Files Added
- `tests/subcategory_tests.go` - Comprehensive test suite
- `cmd/subcategory_example.go` - Usage examples
- `cmd/resource_summary_example.go` - Resource summary examples
- `CHANGELOG.md` - This file

### Usage Examples

#### Get Networking Resources
```go
latest, _ := client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
versionID, _ := client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latest.Version)

// Method 1: Convenience method
networkingResources, err := client.Providers.GetNetworkingResources(ctx, versionID)

// Method 2: Generic method
resources, err := client.Providers.GetResourcesBySubcategory(ctx, versionID, registry.SubcategoryNetworking)

// Method 3: Full control with ListDocsV2
opts := &registry.ProviderDocListOptions{
    ProviderVersionID: versionID,
    Category:          "resources",
    Subcategory:       registry.SubcategoryNetworking,
    Language:          "hcl",
}
docs, err := client.Providers.ListDocsV2(ctx, opts)
```

#### Get Complete Resource Summary
```go
summary, err := client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "aws", "latest")

fmt.Printf("Total Resources: %d\n", summary.TotalResources)
fmt.Printf("Total Data Sources: %d\n", summary.TotalDataSources)

// Access by subcategory
for _, subcategory := range summary.AllSubcategories {
    resources := summary.ResourcesBySubcategory[subcategory]
    fmt.Printf("%s: %d resources\n", subcategory, len(resources))
}
```

### Breaking Changes
None. All changes are backward compatible.

### Deprecations
None.

---

## [1.0.0] - 2025-10-29

### Added
- Initial release
- Full Terraform Registry v1 and v2 API support
- Module operations (list, search, get, download)
- Provider operations (list, get, versions, documentation)
- Policy operations (search, get)
- Automatic retries with exponential backoff
- Built-in rate limiting
- Comprehensive error handling
- Type-safe responses
- Well-tested with integration tests
