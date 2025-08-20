package tests

import (
	"context"
	"fmt"

	"terralense-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// ProviderTests contains tests for the Providers API
type ProviderTests struct {
	*BaseTestSuite
}

// NewProviderTests creates a new provider test suite
func NewProviderTests(client *registry.Client, logger *logrus.Logger) TestSuite {
	suite := &ProviderTests{
		BaseTestSuite: NewBaseTestSuite("Providers", client, logger),
	}

	suite.setupTests()
	return suite
}

func (s *ProviderTests) setupTests() {
	s.AddTest("List Providers", "Test listing providers with various options", s.testListProviders)
	s.AddTest("Get Provider", "Test getting a specific provider", s.testGetProvider)
	s.AddTest("Get Latest Version", "Test getting latest provider version", s.testGetLatestVersion)
	s.AddTest("List Versions", "Test listing all provider versions", s.testListVersions)
	s.AddTest("Get Version ID", "Test getting version ID", s.testGetVersionID)
	s.AddTest("List Documentation", "Test listing provider documentation", s.testListDocs)
	s.AddTest("Get Documentation v2", "Test v2 documentation API", s.testGetDocsV2)
	s.AddTest("Filter by Tier", "Test filtering providers by tier", s.testFilterByTier)
	s.AddTest("Filter by Namespace", "Test filtering by namespace", s.testFilterByNamespace)
	s.AddTest("Invalid Provider", "Test error handling for invalid providers", s.testInvalidProvider)
}

func (s *ProviderTests) testListProviders(ctx context.Context) error {
	opts := &registry.ProviderListOptions{
		PageSize: 10,
	}

	result, err := s.client.Providers.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list providers: %w", err)
	}

	if err := AssertNotNil(result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		return fmt.Errorf("expected at least one provider, got none")
	}

	// Verify provider structure
	for _, provider := range result.Data {
		if provider.ID == "" {
			return fmt.Errorf("provider has empty ID")
		}
		if provider.Type != "providers" {
			return fmt.Errorf("unexpected provider type: %s", provider.Type)
		}
		if provider.Attributes.Namespace == "" {
			return fmt.Errorf("provider has empty namespace")
		}
		if provider.Attributes.Name == "" {
			return fmt.Errorf("provider has empty name")
		}
	}

	s.logger.Debugf("Listed %d providers", len(result.Data))
	return nil
}

func (s *ProviderTests) testGetProvider(ctx context.Context) error {
	// Test with well-known providers
	testCases := []struct {
		namespace string
		name      string
	}{
		{"hashicorp", "aws"},
		{"hashicorp", "azurerm"},
		{"hashicorp", "google"},
	}

	for _, tc := range testCases {
		provider, err := s.client.Providers.Get(ctx, tc.namespace, tc.name)
		if err != nil {
			if registry.IsNotFound(err) {
				s.logger.Warnf("Provider %s/%s not found, skipping", tc.namespace, tc.name)
				continue
			}
			return fmt.Errorf("failed to get provider %s/%s: %w", tc.namespace, tc.name, err)
		}

		// Verify provider details
		if provider.Attributes.Namespace != tc.namespace {
			return fmt.Errorf("namespace mismatch: expected %s, got %s",
				tc.namespace, provider.Attributes.Namespace)
		}
		if provider.Attributes.Name != tc.name {
			return fmt.Errorf("name mismatch: expected %s, got %s",
				tc.name, provider.Attributes.Name)
		}

		// Verify required fields
		if provider.Attributes.FullName == "" {
			return fmt.Errorf("provider full name is empty")
		}
		if provider.Attributes.Tier == "" {
			return fmt.Errorf("provider tier is empty")
		}

		s.logger.Debugf("Got provider %s with %d downloads, tier: %s",
			provider.Attributes.FullName, provider.Attributes.Downloads, provider.Attributes.Tier)

		return nil // Test at least one provider successfully
	}

	return fmt.Errorf("no test providers found")
}

func (s *ProviderTests) testGetLatestVersion(ctx context.Context) error {
	latest, err := s.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	if latest.Version == "" {
		return fmt.Errorf("latest version is empty")
	}

	// Verify version format
	if err := registry.ValidateProviderVersion(latest.Version); err != nil {
		return fmt.Errorf("invalid version format %s: %w", latest.Version, err)
	}

	s.logger.Debugf("Latest AWS provider version: %s", latest.Version)
	return nil
}

func (s *ProviderTests) testListVersions(ctx context.Context) error {
	versions, err := s.client.Providers.ListVersions(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions.Included) == 0 {
		return fmt.Errorf("no versions found")
	}

	// Verify version data
	for _, version := range versions.Included {
		if version.Type != "provider-versions" {
			return fmt.Errorf("unexpected version type: %s", version.Type)
		}
		if version.Attributes.Version == "" {
			return fmt.Errorf("empty version string")
		}

		// Verify version format
		if err := registry.ValidateProviderVersion(version.Attributes.Version); err != nil {
			s.logger.Warnf("Invalid version format: %s - %v", version.Attributes.Version, err)
		}
	}

	s.logger.Debugf("Found %d versions for hashicorp/aws", len(versions.Included))
	return nil
}

func (s *ProviderTests) testGetVersionID(ctx context.Context) error {
	// Test with specific version
	versionID, err := s.client.Providers.GetVersionID(ctx, "hashicorp", "aws", "4.0.0")
	if err != nil {
		// Try with latest
		versionID, err = s.client.Providers.GetVersionID(ctx, "hashicorp", "aws", "latest")
		if err != nil {
			return fmt.Errorf("failed to get version ID: %w", err)
		}
	}

	if versionID == "" {
		return fmt.Errorf("version ID is empty")
	}

	s.logger.Debugf("Got version ID: %s", versionID)
	return nil
}

func (s *ProviderTests) testListDocs(ctx context.Context) error {
	// Get a specific version first
	latest, err := s.client.Providers.GetLatest(ctx, "hashicorp", "aws")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	docs, err := s.client.Providers.ListDocs(ctx, "hashicorp", "aws", latest.Version)
	if err != nil {
		return fmt.Errorf("failed to list docs: %w", err)
	}

	if len(docs.Docs) == 0 {
		return fmt.Errorf("no documentation found")
	}

	// Verify doc structure
	categoriesFound := make(map[string]int)
	for _, doc := range docs.Docs {
		if doc.ID == "" {
			return fmt.Errorf("doc has empty ID")
		}
		if doc.Title == "" {
			return fmt.Errorf("doc has empty title")
		}
		if doc.Category == "" {
			return fmt.Errorf("doc has empty category")
		}

		categoriesFound[doc.Category]++
	}

	// Should have at least resources and data sources
	if categoriesFound["resources"] == 0 {
		return fmt.Errorf("no resource documentation found")
	}
	if categoriesFound["data-sources"] == 0 {
		return fmt.Errorf("no data source documentation found")
	}

	s.logger.Debugf("Found %d docs with categories: %v", len(docs.Docs), categoriesFound)
	return nil
}

func (s *ProviderTests) testGetDocsV2(ctx context.Context) error {
	// Get version ID first
	versionID, err := s.client.Providers.GetVersionID(ctx, "hashicorp", "aws", "latest")
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// Test getting specific resource documentation
	opts := &registry.ProviderDocListOptions{
		ProviderVersionID: versionID,
		Category:          "resources",
		Slug:              "instance",
		Language:          "hcl",
		Page:              1,
	}

	docs, err := s.client.Providers.ListDocsV2(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list docs v2: %w", err)
	}

	if len(docs) == 0 {
		// Try a different resource
		opts.Slug = "s3_bucket"
		docs, err = s.client.Providers.ListDocsV2(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to list docs v2 for s3_bucket: %w", err)
		}
	}

	if len(docs) > 0 {
		// Get detailed documentation
		docID := docs[0].ID
		details, err := s.client.Providers.GetDoc(ctx, docID)
		if err != nil {
			return fmt.Errorf("failed to get doc details: %w", err)
		}

		if details.Data.Attributes.Content == "" {
			return fmt.Errorf("doc content is empty")
		}

		s.logger.Debugf("Got documentation for %s with %d characters of content",
			details.Data.Attributes.Title, len(details.Data.Attributes.Content))
	}

	return nil
}

func (s *ProviderTests) testFilterByTier(ctx context.Context) error {
	tiers := []string{"official", "partner", "community"}

	for _, tier := range tiers {
		opts := &registry.ProviderListOptions{
			Tier:     tier,
			PageSize: 5,
		}

		result, err := s.client.Providers.List(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to filter by tier %s: %w", tier, err)
		}

		// Verify all providers match the tier
		for _, provider := range result.Data {
			if provider.Attributes.Tier != tier {
				return fmt.Errorf("expected tier %s, got %s for provider %s",
					tier, provider.Attributes.Tier, provider.ID)
			}
		}

		s.logger.Debugf("Found %d providers with tier %s", len(result.Data), tier)
	}

	return nil
}

func (s *ProviderTests) testFilterByNamespace(ctx context.Context) error {
	namespaces := []string{"hashicorp", "mongodb", "oracle"}

	for _, namespace := range namespaces {
		opts := &registry.ProviderListOptions{
			Namespace: namespace,
			PageSize:  10,
		}

		result, err := s.client.Providers.List(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to filter by namespace %s: %w", namespace, err)
		}

		// Verify all providers match the namespace
		for _, provider := range result.Data {
			if provider.Attributes.Namespace != namespace {
				return fmt.Errorf("expected namespace %s, got %s for provider %s",
					namespace, provider.Attributes.Namespace, provider.ID)
			}
		}

		s.logger.Debugf("Found %d providers in namespace %s", len(result.Data), namespace)

		if len(result.Data) > 0 {
			break // At least one namespace has providers
		}
	}

	return nil
}

func (s *ProviderTests) testInvalidProvider(ctx context.Context) error {
	// Test with non-existent provider
	_, err := s.client.Providers.Get(ctx, "invalid-namespace", "invalid-provider")

	if err == nil {
		return fmt.Errorf("expected error for invalid provider, got nil")
	}

	if !registry.IsNotFound(err) {
		return fmt.Errorf("expected NotFound error, got: %v", err)
	}

	// Test with invalid version
	_, err = s.client.Providers.GetVersionID(ctx, "hashicorp", "aws", "invalid.version")
	if err == nil {
		return fmt.Errorf("expected error for invalid version, got nil")
	}

	// Test documentation with invalid version ID
	opts := &registry.ProviderDocListOptions{
		ProviderVersionID: "invalid-version-id",
		Category:          "resources",
	}

	_, err = s.client.Providers.ListDocsV2(ctx, opts)
	if err == nil {
		return fmt.Errorf("expected error for invalid version ID, got nil")
	}

	s.logger.Debug("Invalid provider handling works correctly")
	return nil
}
