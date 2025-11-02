package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/TahirRiaz/terralense-registry-client/registry"
	"github.com/sirupsen/logrus"
)

// SubcategoryExample demonstrates how to use subcategory filtering
// to get specific types of resources (Networking, Compute, Storage, etc.)
type SubcategoryExample struct {
	client *registry.Client
	logger *logrus.Logger
}

// NewSubcategoryExample creates a new subcategory example
func NewSubcategoryExample(client *registry.Client, logger *logrus.Logger) *SubcategoryExample {
	return &SubcategoryExample{
		client: client,
		logger: logger,
	}
}

// Run executes the subcategory filtering examples
func (e *SubcategoryExample) Run(ctx context.Context) error {
	fmt.Println("\n=== Subcategory Filtering Examples ===")
	fmt.Println("This demonstrates how to filter provider resources by subcategory")
	fmt.Println(strings.Repeat("=", 70) + "\n")

	// Example 1: Get networking resources
	if err := e.exampleNetworkingResources(ctx); err != nil {
		return err
	}

	// Example 2: Get multiple subcategories
	if err := e.exampleMultipleSubcategories(ctx); err != nil {
		return err
	}

	// Example 3: Get data sources by subcategory
	if err := e.exampleDataSourcesBySubcategory(ctx); err != nil {
		return err
	}

	// Example 4: Compare subcategories across providers
	if err := e.exampleCompareProviders(ctx); err != nil {
		return err
	}

	return nil
}

func (e *SubcategoryExample) exampleNetworkingResources(ctx context.Context) error {
	fmt.Println("Example 1: Getting Networking Resources from Azure Provider")
	fmt.Println(strings.Repeat("-", 70))

	// Get the Azure provider
	provider, err := e.client.Providers.Get(ctx, "hashicorp", "azurerm")
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Get latest version
	latest, err := e.client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	// Get version ID
	versionID, err := e.client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	fmt.Printf("Provider: %s\n", provider.Attributes.FullName)
	fmt.Printf("Version: %s\n\n", latest.Version)

	// Method 1: Using the convenience method
	fmt.Println("Method 1: Using GetNetworkingResources() convenience method")
	networkingResources, err := e.client.Providers.GetNetworkingResources(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get networking resources: %w", err)
	}

	fmt.Printf("Found %d networking resources\n", len(networkingResources))
	e.displaySampleResources(ctx, networkingResources, 5)

	// Method 2: Using the generic method with subcategory constant
	fmt.Println("\nMethod 2: Using GetResourcesBySubcategory() with constant")
	resources, err := e.client.Providers.GetResourcesBySubcategory(
		ctx,
		versionID,
		registry.SubcategoryNetworking,
	)
	if err != nil {
		return fmt.Errorf("failed to get resources: %w", err)
	}

	fmt.Printf("Found %d resources (should match Method 1)\n", len(resources))

	// Method 3: Using ListDocsV2 with full control
	fmt.Println("\nMethod 3: Using ListDocsV2() for full control")
	opts := &registry.ProviderDocListOptions{
		ProviderVersionID: versionID,
		Category:          "resources",
		Subcategory:       registry.SubcategoryNetworking,
		Language:          "hcl",
	}

	docs, err := e.client.Providers.ListDocsV2(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list docs: %w", err)
	}

	fmt.Printf("Found %d docs (should match previous methods)\n\n", len(docs))

	return nil
}

func (e *SubcategoryExample) exampleMultipleSubcategories(ctx context.Context) error {
	fmt.Println("Example 2: Getting Multiple Subcategories from AWS Provider")
	fmt.Println(strings.Repeat("-", 70))

	// Get AWS provider latest version ID
	latest, err := e.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := e.client.Providers.GetVersionID(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	fmt.Printf("Provider: hashicorp/aws\n")
	fmt.Printf("Version: %s\n\n", latest.Version)

	// Get resources for different subcategories
	subcategories := map[string]func(context.Context, string) ([]registry.ProviderData, error){
		"Networking": e.client.Providers.GetNetworkingResources,
		"Compute":    e.client.Providers.GetComputeResources,
		"Storage":    e.client.Providers.GetStorageResources,
		"Database":   e.client.Providers.GetDatabaseResources,
		"Security":   e.client.Providers.GetSecurityResources,
	}

	fmt.Println("Resource counts by subcategory:")
	for name, fn := range subcategories {
		resources, err := fn(ctx, versionID)
		if err != nil {
			e.logger.Warnf("Failed to get %s resources: %v", name, err)
			continue
		}
		fmt.Printf("  %-15s: %4d resources\n", name, len(resources))
	}

	fmt.Println()
	return nil
}

func (e *SubcategoryExample) exampleDataSourcesBySubcategory(ctx context.Context) error {
	fmt.Println("Example 3: Getting Data Sources by Subcategory")
	fmt.Println(strings.Repeat("-", 70))

	latest, err := e.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := e.client.Providers.GetVersionID(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	fmt.Printf("Getting networking data sources from AWS provider\n\n")

	// Get networking data sources
	dataSources, err := e.client.Providers.GetDataSourcesBySubcategory(
		ctx,
		versionID,
		registry.SubcategoryNetworking,
	)
	if err != nil {
		return fmt.Errorf("failed to get data sources: %w", err)
	}

	fmt.Printf("Found %d networking data sources\n", len(dataSources))
	e.displaySampleResources(ctx, dataSources, 5)

	fmt.Println()
	return nil
}

func (e *SubcategoryExample) exampleCompareProviders(ctx context.Context) error {
	fmt.Println("Example 4: Comparing Networking Resources Across Providers")
	fmt.Println(strings.Repeat("-", 70))

	providers := []struct {
		namespace string
		name      string
	}{
		{"hashicorp", "aws"},
		{"hashicorp", "azurerm"},
		{"hashicorp", "google"},
	}

	fmt.Println("Networking resources count comparison:\n")
	fmt.Printf("%-20s | %-10s | %s\n", "Provider", "Version", "Resources")
	fmt.Println(strings.Repeat("-", 70))

	for _, p := range providers {
		latest, err := e.client.Providers.GetLatest(ctx, p.namespace, p.name)
		if err != nil {
			fmt.Printf("%-20s | %-10s | Error: %v\n", p.namespace+"/"+p.name, "N/A", err)
			continue
		}

		versionID, err := e.client.Providers.GetVersionID(ctx, p.namespace, p.name, latest.Version)
		if err != nil {
			fmt.Printf("%-20s | %-10s | Error: %v\n", p.namespace+"/"+p.name, latest.Version, err)
			continue
		}

		resources, err := e.client.Providers.GetNetworkingResources(ctx, versionID)
		if err != nil {
			fmt.Printf("%-20s | %-10s | Error: %v\n", p.namespace+"/"+p.name, latest.Version, err)
			continue
		}

		fmt.Printf("%-20s | %-10s | %d\n", p.namespace+"/"+p.name, latest.Version, len(resources))
	}

	fmt.Println()
	return nil
}

func (e *SubcategoryExample) displaySampleResources(ctx context.Context, resources []registry.ProviderData, limit int) {
	if len(resources) == 0 {
		fmt.Println("  No resources to display")
		return
	}

	if limit > len(resources) {
		limit = len(resources)
	}

	fmt.Println("\nSample resources:")
	for i := 0; i < limit; i++ {
		// Get detailed info
		doc, err := e.client.Providers.GetDoc(ctx, resources[i].ID)
		if err != nil {
			fmt.Printf("  %d. [Error fetching details: %v]\n", i+1, err)
			continue
		}

		fmt.Printf("  %d. %s\n", i+1, doc.Data.Attributes.Title)
		if doc.Data.Attributes.Subcategory != "" {
			fmt.Printf("     Category: %s | Subcategory: %s\n",
				doc.Data.Attributes.Category,
				doc.Data.Attributes.Subcategory)
		}
	}

	if len(resources) > limit {
		fmt.Printf("  ... and %d more\n", len(resources)-limit)
	}
}

// Example usage code for documentation
const exampleUsageCode = `
// Example: Get all networking resources from Azure provider
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TahirRiaz/terralense-registry-client/registry"
)

func main() {
	// Create client
	client, err := registry.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get provider version ID
	latest, _ := client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
	versionID, _ := client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latest.Version)

	// Method 1: Use convenience method
	networkingResources, err := client.Providers.GetNetworkingResources(ctx, versionID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d networking resources\n", len(networkingResources))

	// Method 2: Use generic method with subcategory constant
	resources, err := client.Providers.GetResourcesBySubcategory(
		ctx,
		versionID,
		registry.SubcategoryNetworking,
	)

	// Method 3: Use ListDocsV2 for full control
	opts := &registry.ProviderDocListOptions{
		ProviderVersionID: versionID,
		Category:          "resources",
		Subcategory:       registry.SubcategoryNetworking,
		Language:          "hcl",
	}

	docs, err := client.Providers.ListDocsV2(ctx, opts)

	// Available subcategory constants:
	// - registry.SubcategoryNetworking
	// - registry.SubcategoryCompute
	// - registry.SubcategoryStorage
	// - registry.SubcategoryDatabase
	// - registry.SubcategorySecurity
	// - registry.SubcategoryIdentity
	// - registry.SubcategoryMonitoring
	// - registry.SubcategoryContainer
	// - registry.SubcategoryServerless
	// - registry.SubcategoryAnalytics
	// - registry.SubcategoryMessaging
	// - registry.SubcategoryDeveloper
	// - registry.SubcategoryManagement

	// Get data sources by subcategory
	dataSources, err := client.Providers.GetDataSourcesBySubcategory(
		ctx,
		versionID,
		registry.SubcategoryNetworking,
	)
}
`
