package registry

import (
	"context"
)

// ProvidersServiceInterface defines the interface for provider operations
type ProvidersServiceInterface interface {
	// List returns a list of providers
	List(ctx context.Context, opts *ProviderListOptions) (*ProviderList, error)

	// Get returns details about a specific provider
	Get(ctx context.Context, namespace, name string) (*ProviderData, error)

	// GetLatest returns the latest version info for a provider
	GetLatest(ctx context.Context, namespace, name string) (*ProviderLatestVersion, error)

	// GetVersion returns details about a specific provider version
	GetVersion(ctx context.Context, namespace, name, version string) (*Provider, error)

	// ListVersions returns all versions of a provider
	ListVersions(ctx context.Context, namespace, name string) (*ProviderVersionList, error)

	// GetVersionID returns the version ID for a specific provider version
	GetVersionID(ctx context.Context, namespace, name, version string) (string, error)

	// ListDocs returns documentation for a provider version
	ListDocs(ctx context.Context, namespace, name, version string) (*ProviderDocs, error)

	// ListDocsV2 returns documentation using the v2 API with pagination support
	ListDocsV2(ctx context.Context, opts *ProviderDocListOptions) ([]ProviderData, error)

	// GetDoc returns detailed documentation for a specific provider doc
	GetDoc(ctx context.Context, docID string) (*ProviderDocDetails, error)

	// GetOverviewDocs returns the overview documentation for a provider version
	GetOverviewDocs(ctx context.Context, providerVersionID string) (string, error)
}

// ModulesServiceInterface defines the interface for module operations
type ModulesServiceInterface interface {
	// List returns a list of all modules
	List(ctx context.Context, opts *ModuleListOptions) (*ModuleList, error)

	// Search searches for modules based on a query string
	Search(ctx context.Context, query string, offset int) (*ModuleList, error)

	// SearchWithRelevance searches for modules and calculates relevance scores
	SearchWithRelevance(ctx context.Context, query string, offset int) ([]ModuleSearchResult, error)

	// Get returns details about a specific module version
	Get(ctx context.Context, namespace, name, provider, version string) (*ModuleDetails, error)

	// GetByID returns details about a module using its full ID
	GetByID(ctx context.Context, moduleID string) (*ModuleDetails, error)

	// GetLatest returns the latest version of a module
	GetLatest(ctx context.Context, namespace, name, provider string) (*ModuleDetails, error)

	// ListVersions returns all versions of a module
	ListVersions(ctx context.Context, namespace, name, provider string) ([]string, error)

	// Download returns the download URL for a module
	Download(ctx context.Context, namespace, name, provider, version string) (string, error)
}

// PoliciesServiceInterface defines the interface for policy operations
type PoliciesServiceInterface interface {
	// List returns a list of policies
	List(ctx context.Context, opts *PolicyListOptions) (*PolicyList, error)

	// Get returns details about a specific policy version
	Get(ctx context.Context, namespace, name, version string) (*PolicyDetails, error)

	// GetByID returns details about a policy using its full ID
	GetByID(ctx context.Context, policyID string) (*PolicyDetails, error)

	// Search searches for policies based on a query string
	Search(ctx context.Context, query string) ([]PolicySearchResult, error)

	// GetSentinelContent generates Sentinel policy content for a policy
	GetSentinelContent(ctx context.Context, policyID string) (*SentinelPolicyContent, error)
}
