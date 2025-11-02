package tests

import (
	"context"
	"fmt"

	"github.com/TahirRiaz/terralense-registry-client/registry"
	"github.com/sirupsen/logrus"
)

// SubcategoryTests encapsulates all subcategory-related tests
type SubcategoryTests struct {
	*BaseTestSuite
}

// NewSubcategoryTests creates a new subcategory test suite
func NewSubcategoryTests(client *registry.Client, logger *logrus.Logger) TestSuite {
	suite := &SubcategoryTests{
		BaseTestSuite: NewBaseTestSuite("Subcategory", client, logger),
	}

	suite.setupTests()
	return suite
}

func (s *SubcategoryTests) setupTests() {
	s.AddTest("List Networking Resources", "Test getting networking resources for a provider", s.testListNetworkingResources)
	s.AddTest("List Compute Resources", "Test getting compute resources for a provider", s.testListComputeResources)
	s.AddTest("List Storage Resources", "Test getting storage resources for a provider", s.testListStorageResources)
	s.AddTest("List Database Resources", "Test getting database resources for a provider", s.testListDatabaseResources)
	s.AddTest("List Security Resources", "Test getting security resources for a provider", s.testListSecurityResources)
	s.AddTest("List Resources by Subcategory", "Test getting resources by custom subcategory", s.testListResourcesBySubcategory)
	s.AddTest("List Data Sources by Subcategory", "Test getting data sources by subcategory", s.testListDataSourcesBySubcategory)
	s.AddTest("Validate Subcategory Filtering", "Test subcategory filtering accuracy", s.testSubcategoryFiltering)
	s.AddTest("Test Subcategory Validation", "Test subcategory parameter validation", s.testSubcategoryValidation)
	s.AddTest("Test Multiple Providers", "Test subcategory filtering across multiple providers", s.testMultipleProviders)
}

func (t *SubcategoryTests) testListNetworkingResources(ctx context.Context) error {
	// Test with Azure provider
	provider, err := t.client.Providers.Get(ctx, "hashicorp", "azurerm")
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Get networking resources
	resources, err := t.client.Providers.GetNetworkingResources(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get networking resources: %w", err)
	}

	fmt.Printf("Provider: %s\n", provider.Attributes.FullName)
	fmt.Printf("Version: %s\n", latest.Version)
	fmt.Printf("Networking Resources Found: %d\n", len(resources))

	if len(resources) == 0 {
		return fmt.Errorf("expected networking resources, got none")
	}

	// Display first few resources
	displayCount := 5
	if len(resources) < displayCount {
		displayCount = len(resources)
	}

	fmt.Println("\nSample Networking Resources:")
	for i := 0; i < displayCount; i++ {
		// Get doc details to see the title
		doc, err := t.client.Providers.GetDoc(ctx, resources[i].ID)
		if err != nil {
			fmt.Printf("  %d. ID: %s (error getting details: %v)\n", i+1, resources[i].ID, err)
			continue
		}
		fmt.Printf("  %d. %s (category: %s, subcategory: %s)\n",
			i+1,
			doc.Data.Attributes.Title,
			doc.Data.Attributes.Category,
			doc.Data.Attributes.Subcategory)
	}

	return nil
}

func (t *SubcategoryTests) testListComputeResources(ctx context.Context) error {
	// Test with AWS provider
	provider, err := t.client.Providers.Get(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Get compute resources
	resources, err := t.client.Providers.GetComputeResources(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get compute resources: %w", err)
	}

	fmt.Printf("Provider: %s\n", provider.Attributes.FullName)
	fmt.Printf("Version: %s\n", latest.Version)
	fmt.Printf("Compute Resources Found: %d\n", len(resources))

	if len(resources) == 0 {
		return fmt.Errorf("expected compute resources, got none")
	}

	// Display first few resources
	displayCount := 5
	if len(resources) < displayCount {
		displayCount = len(resources)
	}

	fmt.Println("\nSample Compute Resources:")
	for i := 0; i < displayCount; i++ {
		doc, err := t.client.Providers.GetDoc(ctx, resources[i].ID)
		if err != nil {
			fmt.Printf("  %d. ID: %s (error getting details: %v)\n", i+1, resources[i].ID, err)
			continue
		}
		fmt.Printf("  %d. %s\n", i+1, doc.Data.Attributes.Title)
	}

	return nil
}

func (t *SubcategoryTests) testListStorageResources(ctx context.Context) error {
	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Get storage resources
	resources, err := t.client.Providers.GetStorageResources(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get storage resources: %w", err)
	}

	fmt.Printf("Storage Resources Found: %d\n", len(resources))

	if len(resources) == 0 {
		return fmt.Errorf("expected storage resources, got none")
	}

	return nil
}

func (t *SubcategoryTests) testListDatabaseResources(ctx context.Context) error {
	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Get database resources
	resources, err := t.client.Providers.GetDatabaseResources(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get database resources: %w", err)
	}

	fmt.Printf("Database Resources Found: %d\n", len(resources))

	if len(resources) == 0 {
		return fmt.Errorf("expected database resources, got none")
	}

	return nil
}

func (t *SubcategoryTests) testListSecurityResources(ctx context.Context) error {
	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Get security resources
	resources, err := t.client.Providers.GetSecurityResources(ctx, versionID)
	if err != nil {
		return fmt.Errorf("failed to get security resources: %w", err)
	}

	fmt.Printf("Security Resources Found: %d\n", len(resources))

	if len(resources) == 0 {
		return fmt.Errorf("expected security resources, got none")
	}

	return nil
}

func (t *SubcategoryTests) testListResourcesBySubcategory(ctx context.Context) error {
	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Test with different subcategories
	subcategories := []string{
		registry.SubcategoryNetworking,
		registry.SubcategoryCompute,
		registry.SubcategoryStorage,
	}

	for _, subcategory := range subcategories {
		resources, err := t.client.Providers.GetResourcesBySubcategory(ctx, versionID, subcategory)
		if err != nil {
			return fmt.Errorf("failed to get resources for subcategory %s: %w", subcategory, err)
		}

		fmt.Printf("Subcategory: %s - Resources: %d\n", subcategory, len(resources))

		if len(resources) == 0 {
			fmt.Printf("  Warning: No resources found for subcategory %s\n", subcategory)
		}
	}

	return nil
}

func (t *SubcategoryTests) testListDataSourcesBySubcategory(ctx context.Context) error {
	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Get networking data sources
	dataSources, err := t.client.Providers.GetDataSourcesBySubcategory(ctx, versionID, registry.SubcategoryNetworking)
	if err != nil {
		return fmt.Errorf("failed to get data sources: %w", err)
	}

	fmt.Printf("Networking Data Sources Found: %d\n", len(dataSources))

	if len(dataSources) == 0 {
		fmt.Println("  Warning: No networking data sources found")
	} else {
		// Display first few
		displayCount := 3
		if len(dataSources) < displayCount {
			displayCount = len(dataSources)
		}

		fmt.Println("\nSample Data Sources:")
		for i := 0; i < displayCount; i++ {
			doc, err := t.client.Providers.GetDoc(ctx, dataSources[i].ID)
			if err != nil {
				continue
			}
			fmt.Printf("  %d. %s\n", i+1, doc.Data.Attributes.Title)
		}
	}

	return nil
}

func (t *SubcategoryTests) testSubcategoryFiltering(ctx context.Context) error {
	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Test direct API call with subcategory filter
	opts := &registry.ProviderDocListOptions{
		ProviderVersionID: versionID,
		Category:          "resources",
		Subcategory:       registry.SubcategoryNetworking,
		Language:          "hcl",
		Page:              1,
	}

	docs, err := t.client.Providers.ListDocsV2(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list docs with subcategory filter: %w", err)
	}

	fmt.Printf("Filtered Resources: %d\n", len(docs))

	// Verify all results have the correct subcategory
	if len(docs) > 0 {
		sampleSize := 3
		if len(docs) < sampleSize {
			sampleSize = len(docs)
		}

		fmt.Println("\nVerifying subcategory filter:")
		for i := 0; i < sampleSize; i++ {
			doc, err := t.client.Providers.GetDoc(ctx, docs[i].ID)
			if err != nil {
				continue
			}

			fmt.Printf("  %s - Subcategory: %s\n",
				doc.Data.Attributes.Title,
				doc.Data.Attributes.Subcategory)

			if doc.Data.Attributes.Subcategory != registry.SubcategoryNetworking {
				return fmt.Errorf("expected subcategory %s, got %s for %s",
					registry.SubcategoryNetworking,
					doc.Data.Attributes.Subcategory,
					doc.Data.Attributes.Title)
			}
		}
	}

	return nil
}

func (t *SubcategoryTests) testSubcategoryValidation(ctx context.Context) error {
	latest, err := t.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	versionID, err := t.client.Providers.GetVersionID(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Test with empty subcategory (should fail)
	_, err = t.client.Providers.GetResourcesBySubcategory(ctx, versionID, "")
	if err == nil {
		return fmt.Errorf("expected error for empty subcategory, got nil")
	}
	fmt.Printf("✓ Empty subcategory validation: %v\n", err)

	// Test with empty provider version ID (should fail)
	_, err = t.client.Providers.GetResourcesBySubcategory(ctx, "", registry.SubcategoryNetworking)
	if err == nil {
		return fmt.Errorf("expected error for empty provider version ID, got nil")
	}
	fmt.Printf("✓ Empty provider version ID validation: %v\n", err)

	// Test with valid subcategory constant (should succeed)
	_, err = t.client.Providers.GetResourcesBySubcategory(ctx, versionID, registry.SubcategoryNetworking)
	if err != nil {
		return fmt.Errorf("failed with valid subcategory: %w", err)
	}
	fmt.Printf("✓ Valid subcategory accepted\n")

	return nil
}

func (t *SubcategoryTests) testMultipleProviders(ctx context.Context) error {
	providers := []struct {
		namespace string
		name      string
	}{
		{"hashicorp", "aws"},
		{"hashicorp", "azurerm"},
		{"hashicorp", "google"},
	}

	results := make(map[string]int)

	for _, p := range providers {
		latest, err := t.client.Providers.GetLatest(ctx, p.namespace, p.name)
		if err != nil {
			fmt.Printf("  Warning: Failed to get %s/%s: %v\n", p.namespace, p.name, err)
			continue
		}

		versionID, err := t.client.Providers.GetVersionID(ctx, p.namespace, p.name, latest.Version)
		if err != nil {
			fmt.Printf("  Warning: Failed to get version ID for %s/%s: %v\n", p.namespace, p.name, err)
			continue
		}

		resources, err := t.client.Providers.GetNetworkingResources(ctx, versionID)
		if err != nil {
			fmt.Printf("  Warning: Failed to get networking resources for %s/%s: %v\n", p.namespace, p.name, err)
			continue
		}

		providerKey := fmt.Sprintf("%s/%s", p.namespace, p.name)
		results[providerKey] = len(resources)
		fmt.Printf("  %s: %d networking resources\n", providerKey, len(resources))
	}

	if len(results) == 0 {
		return fmt.Errorf("failed to get networking resources from any provider")
	}

	return nil
}
