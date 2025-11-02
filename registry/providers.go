package registry

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// Common provider documentation subcategories
const (
	// SubcategoryNetworking represents networking-related resources
	SubcategoryNetworking = "Networking"

	// SubcategoryCompute represents compute-related resources
	SubcategoryCompute = "Compute"

	// SubcategoryStorage represents storage-related resources
	SubcategoryStorage = "Storage"

	// SubcategoryDatabase represents database-related resources
	SubcategoryDatabase = "Database"

	// SubcategorySecurity represents security-related resources
	SubcategorySecurity = "Security"

	// SubcategoryIdentity represents identity and access management resources
	SubcategoryIdentity = "Identity"

	// SubcategoryMonitoring represents monitoring and logging resources
	SubcategoryMonitoring = "Monitoring"

	// SubcategoryContainer represents container-related resources
	SubcategoryContainer = "Container"

	// SubcategoryServerless represents serverless-related resources
	SubcategoryServerless = "Serverless"

	// SubcategoryAnalytics represents analytics-related resources
	SubcategoryAnalytics = "Analytics"

	// SubcategoryMessaging represents messaging and queueing resources
	SubcategoryMessaging = "Messaging"

	// SubcategoryDeveloper represents developer tools and resources
	SubcategoryDeveloper = "Developer"

	// SubcategoryManagement represents management and governance resources
	SubcategoryManagement = "Management"
)

// ProvidersService handles communication with the provider related
// methods of the Terraform Registry API.
type ProvidersService struct {
	client *Client
}

// ProviderListOptions specifies optional parameters to the List method
type ProviderListOptions struct {
	// Tier filters providers by tier (official, partner, community)
	Tier string `url:"filter[tier],omitempty"`

	// Namespace filters providers by namespace
	Namespace string `url:"filter[namespace],omitempty"`

	// Page specifies the page number for pagination
	Page int `url:"page[number],omitempty"`

	// PageSize specifies the number of items per page
	PageSize int `url:"page[size],omitempty"`
}

// Validate validates the provider list options
func (o *ProviderListOptions) Validate() error {
	if o == nil {
		return nil
	}

	if o.Tier != "" && !isValidTier(o.Tier) {
		return &ValidationError{
			Field:   "Tier",
			Value:   o.Tier,
			Message: "tier must be one of: official, partner, community",
		}
	}

	if o.Namespace != "" && !isValidNamespace(o.Namespace) {
		return &ValidationError{
			Field:   "Namespace",
			Value:   o.Namespace,
			Message: "invalid namespace format",
		}
	}

	if o.Page < 0 {
		return &ValidationError{
			Field:   "Page",
			Value:   o.Page,
			Message: "page cannot be negative",
		}
	}

	if o.PageSize < 0 || o.PageSize > 100 {
		return &ValidationError{
			Field:   "PageSize",
			Value:   o.PageSize,
			Message: "page size must be between 0 and 100",
		}
	}

	return nil
}

// List returns a list of providers
func (s *ProvidersService) List(ctx context.Context, opts *ProviderListOptions) (*ProviderList, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	path := "providers"
	if opts != nil {
		values := url.Values{}
		if opts.Tier != "" {
			values.Add("filter[tier]", opts.Tier)
		}
		if opts.Namespace != "" {
			values.Add("filter[namespace]", opts.Namespace)
		}
		if opts.Page > 0 {
			values.Add("page[number]", fmt.Sprintf("%d", opts.Page))
		}
		if opts.PageSize > 0 {
			values.Add("page[size]", fmt.Sprintf("%d", opts.PageSize))
		} else {
			values.Add("page[size]", "50") // Default page size
		}
		if len(values) > 0 {
			path = fmt.Sprintf("%s?%s", path, values.Encode())
		}
	}

	var result ProviderList
	if err := s.client.get(ctx, path, "v2", &result); err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	return &result, nil
}

// Get returns details about a specific provider using v2 API
func (s *ProvidersService) Get(ctx context.Context, namespace, name string) (*ProviderData, error) {
	if err := validateProviderParams(namespace, name); err != nil {
		return nil, err
	}

	// Use v2 API with proper endpoint structure
	path := fmt.Sprintf("providers?filter[namespace]=%s&filter[name]=%s",
		url.QueryEscape(namespace), url.QueryEscape(name))

	var result struct {
		Data []ProviderData `json:"data"`
	}

	if err := s.client.get(ctx, path, "v2", &result); err != nil {
		return nil, fmt.Errorf("failed to get provider %s/%s: %w", namespace, name, err)
	}

	if len(result.Data) == 0 {
		return nil, &APIError{
			StatusCode: 404,
			Message:    fmt.Sprintf("provider %s/%s not found", namespace, name),
		}
	}

	return &result.Data[0], nil
}

// GetLatest returns the latest version info for a provider
func (s *ProvidersService) GetLatest(ctx context.Context, namespace, name string) (*ProviderLatestVersion, error) {
	if err := validateProviderParams(namespace, name); err != nil {
		return nil, err
	}

	// First get the provider
	provider, err := s.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	// Get versions with included data
	path := fmt.Sprintf("providers/%s?include=provider-versions", provider.ID)

	var result struct {
		Data     ProviderData  `json:"data"`
		Included []VersionData `json:"included"`
	}

	if err := s.client.get(ctx, path, "v2", &result); err != nil {
		return nil, fmt.Errorf("failed to get provider versions: %w", err)
	}

	// Find the latest version
	var latestVersion string
	for _, version := range result.Included {
		if latestVersion == "" || CompareVersions(version.Attributes.Version, latestVersion) > 0 {
			latestVersion = version.Attributes.Version
		}
	}

	if latestVersion == "" {
		return nil, fmt.Errorf("no versions found for provider %s/%s", namespace, name)
	}

	return &ProviderLatestVersion{
		Provider: result.Data,
		Version:  latestVersion,
	}, nil
}

// GetVersion returns details about a specific provider version
func (s *ProvidersService) GetVersion(ctx context.Context, namespace, name, version string) (*Provider, error) {
	if err := validateProviderParams(namespace, name); err != nil {
		return nil, err
	}

	if err := ValidateProviderVersion(version); err != nil {
		return nil, &ValidationError{
			Field:   "version",
			Value:   version,
			Message: err.Error(),
		}
	}

	path := fmt.Sprintf("providers/%s/%s/%s", namespace, name, version)

	var result Provider
	if err := s.client.get(ctx, path, "v1", &result); err != nil {
		return nil, fmt.Errorf("failed to get provider version: %w", err)
	}

	return &result, nil
}

// ListVersions returns all versions of a provider
func (s *ProvidersService) ListVersions(ctx context.Context, namespace, name string) (*ProviderVersionList, error) {
	if err := validateProviderParams(namespace, name); err != nil {
		return nil, err
	}

	// First, get the provider to get its ID
	provider, err := s.Get(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("providers/%s?include=provider-versions", provider.ID)

	var result ProviderVersionList
	if err := s.client.get(ctx, path, "v2", &result); err != nil {
		return nil, fmt.Errorf("failed to list provider versions: %w", err)
	}

	return &result, nil
}

// GetVersionID returns the version ID for a specific provider version
func (s *ProvidersService) GetVersionID(ctx context.Context, namespace, name, version string) (string, error) {
	if err := validateProviderParams(namespace, name); err != nil {
		return "", err
	}

	// Handle latest version
	if version == "" || version == "latest" {
		latest, err := s.GetLatest(ctx, namespace, name)
		if err != nil {
			return "", err
		}
		version = latest.Version
	} else if err := ValidateProviderVersion(version); err != nil {
		return "", &ValidationError{
			Field:   "version",
			Value:   version,
			Message: err.Error(),
		}
	}

	// Get all versions to find the ID
	versions, err := s.ListVersions(ctx, namespace, name)
	if err != nil {
		return "", err
	}

	for _, v := range versions.Included {
		if v.Attributes.Version == version {
			return v.ID, nil
		}
	}

	return "", &APIError{
		StatusCode: 404,
		Message:    fmt.Sprintf("provider version %s/%s@%s not found", namespace, name, version),
	}
}

// ListDocs returns documentation for a provider version
func (s *ProvidersService) ListDocs(ctx context.Context, namespace, name, version string) (*ProviderDocs, error) {
	if err := validateProviderParams(namespace, name); err != nil {
		return nil, err
	}

	if err := ValidateProviderVersion(version); err != nil {
		return nil, &ValidationError{
			Field:   "version",
			Value:   version,
			Message: err.Error(),
		}
	}

	path := fmt.Sprintf("providers/%s/%s/%s", namespace, name, version)

	var result ProviderDocs
	if err := s.client.get(ctx, path, "v1", &result); err != nil {
		return nil, fmt.Errorf("failed to list provider docs: %w", err)
	}

	return &result, nil
}

// ProviderDocListOptions specifies optional parameters for listing provider docs
type ProviderDocListOptions struct {
	// ProviderVersionID is the provider version ID (required)
	ProviderVersionID string

	// Category filters docs by category (resources, data-sources, guides, etc.)
	Category string

	// Subcategory filters docs by subcategory (e.g., Networking, Compute, Storage)
	Subcategory string

	// Slug filters docs by slug
	Slug string

	// Language filters docs by language (default: hcl)
	Language string

	// Page specifies the page number for pagination
	Page int
}

// Validate validates the provider doc list options
func (o *ProviderDocListOptions) Validate() error {
	if o == nil {
		return &ValidationError{
			Field:   "options",
			Message: "options cannot be nil",
		}
	}

	if o.ProviderVersionID == "" {
		return &ValidationError{
			Field:   "ProviderVersionID",
			Message: "provider version ID is required",
		}
	}

	if o.Category != "" && !isValidDocCategory(o.Category) {
		return &ValidationError{
			Field:   "Category",
			Value:   o.Category,
			Message: "invalid category, must be one of: resources, data-sources, functions, guides, overview",
		}
	}

	if o.Language != "" && !isValidLanguage(o.Language) {
		return &ValidationError{
			Field:   "Language",
			Value:   o.Language,
			Message: "invalid language",
		}
	}

	if o.Page < 0 {
		return &ValidationError{
			Field:   "Page",
			Value:   o.Page,
			Message: "page cannot be negative",
		}
	}

	return nil
}

// ListDocsV2 returns documentation using the v2 API with pagination support
func (s *ProvidersService) ListDocsV2(ctx context.Context, opts *ProviderDocListOptions) ([]ProviderData, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	var allDocs []ProviderData
	page := 1
	if opts.Page > 0 {
		page = opts.Page
	}

	maxPages := 100 // Prevent infinite loops

	for pageCount := 0; pageCount < maxPages; pageCount++ {
		values := url.Values{}
		values.Add("filter[provider-version]", opts.ProviderVersionID)

		if opts.Category != "" {
			values.Add("filter[category]", opts.Category)
		}
		if opts.Subcategory != "" {
			values.Add("filter[subcategory]", opts.Subcategory)
		}
		if opts.Slug != "" {
			values.Add("filter[slug]", opts.Slug)
		}
		if opts.Language != "" {
			values.Add("filter[language]", opts.Language)
		} else {
			values.Add("filter[language]", "hcl")
		}

		values.Add("page[number]", fmt.Sprintf("%d", page))
		values.Add("page[size]", "50")

		path := fmt.Sprintf("provider-docs?%s", values.Encode())

		var result struct {
			Data []ProviderData `json:"data"`
			Meta struct {
				Pagination Pagination `json:"pagination"`
			} `json:"meta"`
		}

		if err := s.client.get(ctx, path, "v2", &result); err != nil {
			return nil, fmt.Errorf("failed to list provider docs: %w", err)
		}

		if len(result.Data) == 0 {
			break
		}

		allDocs = append(allDocs, result.Data...)

		// If we're only getting a specific page, don't continue
		if opts.Page > 0 {
			break
		}

		// Check if there are more pages
		if result.Meta.Pagination.NextPage == 0 {
			break
		}

		page = result.Meta.Pagination.NextPage
	}

	return allDocs, nil
}

// GetDoc returns detailed documentation for a specific provider doc
func (s *ProvidersService) GetDoc(ctx context.Context, docID string) (*ProviderDocDetails, error) {
	if docID == "" {
		return nil, &ValidationError{
			Field:   "docID",
			Value:   docID,
			Message: "doc ID cannot be empty",
		}
	}

	path := fmt.Sprintf("provider-docs/%s", docID)

	var result ProviderDocDetails
	if err := s.client.get(ctx, path, "v2", &result); err != nil {
		return nil, fmt.Errorf("failed to get provider doc: %w", err)
	}

	return &result, nil
}

// GetOverviewDocs returns the overview documentation for a provider version
func (s *ProvidersService) GetOverviewDocs(ctx context.Context, providerVersionID string) (string, error) {
	if providerVersionID == "" {
		return "", &ValidationError{
			Field:   "providerVersionID",
			Value:   providerVersionID,
			Message: "provider version ID cannot be empty",
		}
	}

	opts := &ProviderDocListOptions{
		ProviderVersionID: providerVersionID,
		Category:          "overview",
		Slug:              "index",
	}

	docs, err := s.ListDocsV2(ctx, opts)
	if err != nil {
		return "", err
	}

	if len(docs) == 0 {
		return "", &APIError{
			StatusCode: 404,
			Message:    "overview documentation not found",
		}
	}

	var content strings.Builder
	for _, doc := range docs {
		details, err := s.GetDoc(ctx, doc.ID)
		if err != nil {
			return "", err
		}
		content.WriteString(details.Data.Attributes.Content)
		content.WriteString("\n")
	}

	return content.String(), nil
}

// GetResourcesBySubcategory returns all resources for a specific subcategory
func (s *ProvidersService) GetResourcesBySubcategory(ctx context.Context, providerVersionID, subcategory string) ([]ProviderData, error) {
	if providerVersionID == "" {
		return nil, &ValidationError{
			Field:   "providerVersionID",
			Value:   providerVersionID,
			Message: "provider version ID cannot be empty",
		}
	}

	if subcategory == "" {
		return nil, &ValidationError{
			Field:   "subcategory",
			Value:   subcategory,
			Message: "subcategory cannot be empty",
		}
	}

	if !isValidSubcategory(subcategory) {
		return nil, &ValidationError{
			Field:   "subcategory",
			Value:   subcategory,
			Message: "invalid subcategory",
		}
	}

	opts := &ProviderDocListOptions{
		ProviderVersionID: providerVersionID,
		Category:          "resources",
		Subcategory:       subcategory,
		Language:          "hcl",
	}

	docs, err := s.ListDocsV2(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources for subcategory %s: %w", subcategory, err)
	}

	return docs, nil
}

// GetNetworkingResources returns all networking resources for a provider version
func (s *ProvidersService) GetNetworkingResources(ctx context.Context, providerVersionID string) ([]ProviderData, error) {
	return s.GetResourcesBySubcategory(ctx, providerVersionID, SubcategoryNetworking)
}

// GetComputeResources returns all compute resources for a provider version
func (s *ProvidersService) GetComputeResources(ctx context.Context, providerVersionID string) ([]ProviderData, error) {
	return s.GetResourcesBySubcategory(ctx, providerVersionID, SubcategoryCompute)
}

// GetStorageResources returns all storage resources for a provider version
func (s *ProvidersService) GetStorageResources(ctx context.Context, providerVersionID string) ([]ProviderData, error) {
	return s.GetResourcesBySubcategory(ctx, providerVersionID, SubcategoryStorage)
}

// GetDatabaseResources returns all database resources for a provider version
func (s *ProvidersService) GetDatabaseResources(ctx context.Context, providerVersionID string) ([]ProviderData, error) {
	return s.GetResourcesBySubcategory(ctx, providerVersionID, SubcategoryDatabase)
}

// GetSecurityResources returns all security resources for a provider version
func (s *ProvidersService) GetSecurityResources(ctx context.Context, providerVersionID string) ([]ProviderData, error) {
	return s.GetResourcesBySubcategory(ctx, providerVersionID, SubcategorySecurity)
}

// GetDataSourcesBySubcategory returns all data sources for a specific subcategory
func (s *ProvidersService) GetDataSourcesBySubcategory(ctx context.Context, providerVersionID, subcategory string) ([]ProviderData, error) {
	if providerVersionID == "" {
		return nil, &ValidationError{
			Field:   "providerVersionID",
			Value:   providerVersionID,
			Message: "provider version ID cannot be empty",
		}
	}

	if subcategory == "" {
		return nil, &ValidationError{
			Field:   "subcategory",
			Value:   subcategory,
			Message: "subcategory cannot be empty",
		}
	}

	if !isValidSubcategory(subcategory) {
		return nil, &ValidationError{
			Field:   "subcategory",
			Value:   subcategory,
			Message: "invalid subcategory",
		}
	}

	opts := &ProviderDocListOptions{
		ProviderVersionID: providerVersionID,
		Category:          "data-sources",
		Subcategory:       subcategory,
		Language:          "hcl",
	}

	docs, err := s.ListDocsV2(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get data sources for subcategory %s: %w", subcategory, err)
	}

	return docs, nil
}

// GetProviderResourceSummary creates a structured summary of all provider resources and data sources
// organized by subcategory, returning only key information for application use
func (s *ProvidersService) GetProviderResourceSummary(ctx context.Context, namespace, name, version string) (*ProviderResourceSummary, error) {
	if err := validateProviderParams(namespace, name); err != nil {
		return nil, err
	}

	// Get provider version ID
	var versionID string
	var actualVersion string
	var err error

	if version == "" || version == "latest" {
		latest, err := s.GetLatest(ctx, namespace, name)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version: %w", err)
		}
		actualVersion = latest.Version
		versionID, err = s.GetVersionID(ctx, namespace, name, actualVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to get version ID: %w", err)
		}
	} else {
		actualVersion = version
		versionID, err = s.GetVersionID(ctx, namespace, name, version)
		if err != nil {
			return nil, fmt.Errorf("failed to get version ID: %w", err)
		}
	}

	// Get all resources
	resourceOpts := &ProviderDocListOptions{
		ProviderVersionID: versionID,
		Category:          "resources",
		Language:          "hcl",
	}

	resources, err := s.ListDocsV2(ctx, resourceOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	// Get all data sources
	dataSourceOpts := &ProviderDocListOptions{
		ProviderVersionID: versionID,
		Category:          "data-sources",
		Language:          "hcl",
	}

	dataSources, err := s.ListDocsV2(ctx, dataSourceOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get data sources: %w", err)
	}

	// Build the summary
	summary := &ProviderResourceSummary{
		ProviderNamespace:        namespace,
		ProviderName:             name,
		Version:                  actualVersion,
		TotalResources:           len(resources),
		TotalDataSources:         len(dataSources),
		ResourcesBySubcategory:   make(map[string][]ResourceInfo),
		DataSourcesBySubcategory: make(map[string][]ResourceInfo),
		AllSubcategories:         make([]string, 0),
	}

	// Track unique subcategories
	subcategorySet := make(map[string]bool)

	// Process resources
	for _, resource := range resources {
		// Get detailed info to access subcategory
		doc, err := s.GetDoc(ctx, resource.ID)
		if err != nil {
			// If we can't get details, skip this resource
			continue
		}

		attrs := doc.Data.Attributes
		subcategory := attrs.Subcategory
		if subcategory == "" {
			subcategory = "Other"
		}

		resourceInfo := ResourceInfo{
			ID:          resource.ID,
			Name:        attrs.Slug,
			Title:       attrs.Title,
			Subcategory: subcategory,
			Category:    attrs.Category,
			Slug:        attrs.Slug,
			Path:        attrs.Path,
		}

		summary.ResourcesBySubcategory[subcategory] = append(
			summary.ResourcesBySubcategory[subcategory],
			resourceInfo,
		)

		subcategorySet[subcategory] = true
	}

	// Process data sources
	for _, dataSource := range dataSources {
		doc, err := s.GetDoc(ctx, dataSource.ID)
		if err != nil {
			continue
		}

		attrs := doc.Data.Attributes
		subcategory := attrs.Subcategory
		if subcategory == "" {
			subcategory = "Other"
		}

		resourceInfo := ResourceInfo{
			ID:          dataSource.ID,
			Name:        attrs.Slug,
			Title:       attrs.Title,
			Subcategory: subcategory,
			Category:    attrs.Category,
			Slug:        attrs.Slug,
			Path:        attrs.Path,
		}

		summary.DataSourcesBySubcategory[subcategory] = append(
			summary.DataSourcesBySubcategory[subcategory],
			resourceInfo,
		)

		subcategorySet[subcategory] = true
	}

	// Create sorted list of subcategories
	for subcategory := range subcategorySet {
		summary.AllSubcategories = append(summary.AllSubcategories, subcategory)
	}

	// Sort subcategories alphabetically
	sortSubcategories(summary.AllSubcategories)

	return summary, nil
}

// BuildResourceInfoFromDocs creates a simplified resource list from provider documentation
// This is a lighter-weight alternative to GetProviderResourceSummary that doesn't fetch detailed docs
func (s *ProvidersService) BuildResourceInfoFromDocs(docs []ProviderData) []ResourceInfo {
	resources := make([]ResourceInfo, 0, len(docs))

	for _, doc := range docs {
		// Note: ProviderData doesn't contain subcategory, so we need to get doc details
		// This method is provided for when you already have doc details
		resources = append(resources, ResourceInfo{
			ID:   doc.ID,
			Type: doc.Type,
		})
	}

	return resources
}

// ExtractResourceInfoFromProviderDocs extracts key resource information from raw provider documentation response
// This is optimized for when you already have the full doc details and want to avoid additional API calls
func ExtractResourceInfoFromProviderDocs(docs []ProviderDocDetails) []ResourceInfo {
	resources := make([]ResourceInfo, 0, len(docs))

	for _, doc := range docs {
		attrs := doc.Data.Attributes

		resources = append(resources, ResourceInfo{
			ID:          doc.Data.ID,
			Name:        attrs.Slug,
			Title:       attrs.Title,
			Subcategory: attrs.Subcategory,
			Category:    attrs.Category,
			Slug:        attrs.Slug,
			Path:        attrs.Path,
		})
	}

	return resources
}

// ProviderLatestVersion represents a provider with version info
type ProviderLatestVersion struct {
	Provider ProviderData
	Version  string
}

// Helper functions for validation

func validateProviderParams(namespace, name string) error {
	var errs MultiError

	if namespace == "" {
		errs.Add(&ValidationError{
			Field:   "namespace",
			Value:   namespace,
			Message: "namespace cannot be empty",
		})
	} else if !isValidNamespace(namespace) {
		errs.Add(&ValidationError{
			Field:   "namespace",
			Value:   namespace,
			Message: "invalid namespace format",
		})
	}

	if name == "" {
		errs.Add(&ValidationError{
			Field:   "name",
			Value:   name,
			Message: "name cannot be empty",
		})
	} else if !isValidProviderName(name) {
		errs.Add(&ValidationError{
			Field:   "name",
			Value:   name,
			Message: "invalid provider name format",
		})
	}

	return errs.ErrorOrNil()
}

func isValidTier(tier string) bool {
	validTiers := []string{"official", "partner", "community"}
	for _, valid := range validTiers {
		if tier == valid {
			return true
		}
	}
	return false
}

func isValidDocCategory(category string) bool {
	validCategories := []string{"resources", "data-sources", "functions", "guides", "overview"}
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

func isValidSubcategory(subcategory string) bool {
	// Common subcategories across major cloud providers
	// Note: This validation is lenient - providers may use custom subcategories
	// We only validate against known common subcategories
	validSubcategories := []string{
		SubcategoryNetworking,
		SubcategoryCompute,
		SubcategoryStorage,
		SubcategoryDatabase,
		SubcategorySecurity,
		SubcategoryIdentity,
		SubcategoryMonitoring,
		SubcategoryContainer,
		SubcategoryServerless,
		SubcategoryAnalytics,
		SubcategoryMessaging,
		SubcategoryDeveloper,
		SubcategoryManagement,
	}

	for _, valid := range validSubcategories {
		if subcategory == valid {
			return true
		}
	}

	// Allow any subcategory that's not empty (providers may have custom ones)
	// This makes the validation lenient but still provides helpful constants
	return subcategory != ""
}

func isValidLanguage(language string) bool {
	// Add more languages as needed
	validLanguages := []string{"hcl", "terraform", "json"}
	for _, valid := range validLanguages {
		if language == valid {
			return true
		}
	}
	return false
}

func sortSubcategories(subcategories []string) {
	// Simple bubble sort for small lists
	n := len(subcategories)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if subcategories[j] > subcategories[j+1] {
				subcategories[j], subcategories[j+1] = subcategories[j+1], subcategories[j]
			}
		}
	}
}
