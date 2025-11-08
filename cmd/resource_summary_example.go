package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TahirRiaz/terralens-registry-client/registry"
	"github.com/sirupsen/logrus"
)

// ResourceSummaryExample demonstrates how to get a structured summary
// of provider resources organized by subcategory
type ResourceSummaryExample struct {
	client *registry.Client
	logger *logrus.Logger
}

// NewResourceSummaryExample creates a new resource summary example
func NewResourceSummaryExample(client *registry.Client, logger *logrus.Logger) *ResourceSummaryExample {
	return &ResourceSummaryExample{
		client: client,
		logger: logger,
	}
}

// Run executes the resource summary examples
func (e *ResourceSummaryExample) Run(ctx context.Context) error {
	fmt.Println("\n=== Provider Resource Summary Examples ===")
	fmt.Println("This demonstrates how to get structured resource summaries")
	fmt.Println(strings.Repeat("=", 70) + "\n")

	// Example 1: Get AWS provider resource summary
	if err := e.exampleAWSResourceSummary(ctx); err != nil {
		return err
	}

	// Example 2: Get Azure provider resource summary
	if err := e.exampleAzureResourceSummary(ctx); err != nil {
		return err
	}

	// Example 3: Export summary as JSON
	if err := e.exampleExportJSON(ctx); err != nil {
		return err
	}

	// Example 4: Filter specific subcategories
	if err := e.exampleFilterSubcategories(ctx); err != nil {
		return err
	}

	return nil
}

func (e *ResourceSummaryExample) exampleAWSResourceSummary(ctx context.Context) error {
	fmt.Println("Example 1: Getting AWS Provider Resource Summary")
	fmt.Println(strings.Repeat("-", 70))

	// Get complete resource summary
	summary, err := e.client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "aws", "latest")
	if err != nil {
		return fmt.Errorf("failed to get resource summary: %w", err)
	}

	// Display summary statistics
	fmt.Printf("Provider: %s/%s\n", summary.ProviderNamespace, summary.ProviderName)
	fmt.Printf("Version: %s\n", summary.Version)
	fmt.Printf("Total Resources: %d\n", summary.TotalResources)
	fmt.Printf("Total Data Sources: %d\n", summary.TotalDataSources)
	fmt.Printf("Subcategories: %d\n\n", len(summary.AllSubcategories))

	// Display resources by subcategory
	fmt.Println("Resources by Subcategory (Top 5 subcategories):")
	fmt.Println(strings.Repeat("-", 70))

	displayLimit := 5
	for i, subcategory := range summary.AllSubcategories {
		if i >= displayLimit {
			break
		}

		resources := summary.ResourcesBySubcategory[subcategory]
		dataSources := summary.DataSourcesBySubcategory[subcategory]

		fmt.Printf("\n%s:\n", subcategory)
		fmt.Printf("  Resources: %d\n", len(resources))
		fmt.Printf("  Data Sources: %d\n", len(dataSources))

		// Show first 3 resources
		if len(resources) > 0 {
			fmt.Println("  Sample Resources:")
			sampleCount := 3
			if len(resources) < sampleCount {
				sampleCount = len(resources)
			}
			for j := 0; j < sampleCount; j++ {
				fmt.Printf("    - %s\n", resources[j].Title)
			}
		}
	}

	if len(summary.AllSubcategories) > displayLimit {
		fmt.Printf("\n... and %d more subcategories\n", len(summary.AllSubcategories)-displayLimit)
	}

	fmt.Println()
	return nil
}

func (e *ResourceSummaryExample) exampleAzureResourceSummary(ctx context.Context) error {
	fmt.Println("Example 2: Getting Azure Provider Resource Summary")
	fmt.Println(strings.Repeat("-", 70))

	summary, err := e.client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "azurerm", "latest")
	if err != nil {
		return fmt.Errorf("failed to get resource summary: %w", err)
	}

	fmt.Printf("Provider: %s/%s v%s\n", summary.ProviderNamespace, summary.ProviderName, summary.Version)
	fmt.Printf("Total Resources: %d\n", summary.TotalResources)
	fmt.Printf("Total Data Sources: %d\n\n", summary.TotalDataSources)

	// Show networking-specific resources
	if networkingResources, ok := summary.ResourcesBySubcategory["Networking"]; ok {
		fmt.Printf("Networking Resources: %d\n", len(networkingResources))
		fmt.Println("Sample networking resources:")
		for i, resource := range networkingResources {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(networkingResources)-5)
				break
			}
			fmt.Printf("  - %s (slug: %s)\n", resource.Title, resource.Slug)
		}
	}

	fmt.Println()
	return nil
}

func (e *ResourceSummaryExample) exampleExportJSON(ctx context.Context) error {
	fmt.Println("Example 3: Exporting Resource Summary as JSON")
	fmt.Println(strings.Repeat("-", 70))

	summary, err := e.client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "google", "latest")
	if err != nil {
		return fmt.Errorf("failed to get resource summary: %w", err)
	}

	// Create a simplified structure for JSON export
	type SimplifiedSummary struct {
		Provider         string              `json:"provider"`
		Version          string              `json:"version"`
		TotalResources   int                 `json:"total_resources"`
		TotalDataSources int                 `json:"total_data_sources"`
		Subcategories    []string            `json:"subcategories"`
		ResourceCounts   map[string]int      `json:"resource_counts_by_subcategory"`
		SampleResources  map[string][]string `json:"sample_resources"`
	}

	simplified := SimplifiedSummary{
		Provider:         fmt.Sprintf("%s/%s", summary.ProviderNamespace, summary.ProviderName),
		Version:          summary.Version,
		TotalResources:   summary.TotalResources,
		TotalDataSources: summary.TotalDataSources,
		Subcategories:    summary.AllSubcategories,
		ResourceCounts:   make(map[string]int),
		SampleResources:  make(map[string][]string),
	}

	// Populate counts and samples
	for subcategory, resources := range summary.ResourcesBySubcategory {
		simplified.ResourceCounts[subcategory] = len(resources)

		// Get first 3 resources as samples
		samples := make([]string, 0, 3)
		for i, res := range resources {
			if i >= 3 {
				break
			}
			samples = append(samples, res.Title)
		}
		simplified.SampleResources[subcategory] = samples
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(simplified, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println("JSON Export (truncated):")
	// Print first 500 characters
	if len(jsonData) > 500 {
		fmt.Println(string(jsonData[:500]))
		fmt.Printf("\n... (truncated, total size: %d bytes)\n", len(jsonData))
	} else {
		fmt.Println(string(jsonData))
	}

	fmt.Println()
	return nil
}

func (e *ResourceSummaryExample) exampleFilterSubcategories(ctx context.Context) error {
	fmt.Println("Example 4: Filtering Specific Subcategories")
	fmt.Println(strings.Repeat("-", 70))

	summary, err := e.client.Providers.GetProviderResourceSummary(ctx, "hashicorp", "aws", "latest")
	if err != nil {
		return fmt.Errorf("failed to get resource summary: %w", err)
	}

	// Filter for specific subcategories of interest
	interestedSubcategories := []string{
		"VPC (Virtual Private Cloud)",
		"EC2 (Elastic Compute Cloud)",
		"S3 (Simple Storage)",
		"RDS (Relational Database)",
		"Lambda",
	}

	fmt.Println("Filtering for specific subcategories:")
	fmt.Println()

	for _, subcategory := range interestedSubcategories {
		resources, hasResources := summary.ResourcesBySubcategory[subcategory]
		dataSources, hasDataSources := summary.DataSourcesBySubcategory[subcategory]

		if !hasResources && !hasDataSources {
			fmt.Printf("%-40s: Not found\n", subcategory)
			continue
		}

		fmt.Printf("%-40s:\n", subcategory)
		if hasResources {
			fmt.Printf("  Resources:     %3d items\n", len(resources))
		}
		if hasDataSources {
			fmt.Printf("  Data Sources:  %3d items\n", len(dataSources))
		}
		fmt.Println()
	}

	return nil
}

// Example usage code
const resourceSummaryUsageCode = `
// Get a complete resource summary for a provider
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/TahirRiaz/terralens-registry-client/registry"
)

func main() {
	client, _ := registry.NewClient()
	ctx := context.Background()

	// Get complete summary
	summary, err := client.Providers.GetProviderResourceSummary(
		ctx,
		"hashicorp",  // namespace
		"aws",        // provider name
		"latest",     // version (or specific version like "5.0.0")
	)
	if err != nil {
		log.Fatal(err)
	}

	// Access summary data
	fmt.Printf("Provider: %s/%s v%s\n",
		summary.ProviderNamespace,
		summary.ProviderName,
		summary.Version)

	fmt.Printf("Total Resources: %d\n", summary.TotalResources)
	fmt.Printf("Total Data Sources: %d\n", summary.TotalDataSources)

	// Iterate through subcategories
	for _, subcategory := range summary.AllSubcategories {
		resources := summary.ResourcesBySubcategory[subcategory]
		dataSources := summary.DataSourcesBySubcategory[subcategory]

		fmt.Printf("\n%s:\n", subcategory)
		fmt.Printf("  Resources: %d\n", len(resources))
		fmt.Printf("  Data Sources: %d\n", len(dataSources))

		// Access individual resource info
		for _, resource := range resources {
			fmt.Printf("    - %s (slug: %s, path: %s)\n",
				resource.Title,
				resource.Slug,
				resource.Path)
		}
	}

	// Filter for networking resources
	networkingResources := summary.ResourcesBySubcategory["VPC (Virtual Private Cloud)"]
	for _, res := range networkingResources {
		fmt.Printf("VPC Resource: %s\n", res.Title)
	}
}
`
