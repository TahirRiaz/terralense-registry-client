package registry

import (
	"encoding/json"
	"time"
)

// Provider represents a Terraform provider
type Provider struct {
	ID          string    `json:"id"`
	Owner       string    `json:"owner"`
	Namespace   string    `json:"namespace"`
	Name        string    `json:"name"`
	Alias       string    `json:"alias,omitempty"`
	Version     string    `json:"version"`
	Tag         string    `json:"tag,omitempty"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	Downloads   int64     `json:"downloads"`
	Tier        string    `json:"tier"`
	LogoURL     string    `json:"logo_url,omitempty"`
	Versions    []string  `json:"versions,omitempty"`
}

// ProviderDoc represents a provider documentation item
type ProviderDoc struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Path        string `json:"path"`
	Slug        string `json:"slug"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory,omitempty"`
	Language    string `json:"language"`
}

// ProviderDocs represents a provider with its documentation
type ProviderDocs struct {
	Provider
	Docs []ProviderDoc `json:"docs"`
}

// ProviderList represents a paginated list of providers (v2 API)
type ProviderList struct {
	Data  []ProviderData `json:"data"`
	Links Links          `json:"links"`
	Meta  Meta           `json:"meta"`
}

// ProviderData represents provider data in v2 API responses
type ProviderData struct {
	Type       string             `json:"type"`
	ID         string             `json:"id"`
	Attributes ProviderAttributes `json:"attributes"`
	Links      SelfLink           `json:"links"`
}

// ProviderAttributes represents provider attributes in v2 API
type ProviderAttributes struct {
	Alias         string `json:"alias,omitempty"`
	Description   string `json:"description"`
	Downloads     int64  `json:"downloads"`
	Featured      bool   `json:"featured"`
	FullName      string `json:"full-name"`
	LogoURL       string `json:"logo-url,omitempty"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	OwnerName     string `json:"owner-name"`
	RobotsNoindex bool   `json:"robots-noindex"`
	Source        string `json:"source"`
	Tier          string `json:"tier"`
	Unlisted      bool   `json:"unlisted"`
	Warning       string `json:"warning,omitempty"`
}

// ProviderVersionList represents a provider with its versions
type ProviderVersionList struct {
	Data     ProviderVersionData `json:"data"`
	Included []VersionData       `json:"included"`
}

// ProviderVersionData represents provider version data
type ProviderVersionData struct {
	Type          string                   `json:"type"`
	ID            string                   `json:"id"`
	Attributes    ProviderAttributes       `json:"attributes"`
	Relationships ProviderVersionRelations `json:"relationships"`
	Links         SelfLink                 `json:"links"`
}

// ProviderVersionRelations represents provider version relationships
type ProviderVersionRelations struct {
	ProviderVersions RelationshipData `json:"provider-versions"`
}

// RelationshipData represents relationship data
type RelationshipData struct {
	Data []ResourceIdentifier `json:"data"`
}

// ResourceIdentifier identifies a resource
type ResourceIdentifier struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// VersionData represents version data
type VersionData struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Attributes VersionAttributes `json:"attributes"`
	Links      SelfLink          `json:"links"`
}

// VersionAttributes represents version attributes
type VersionAttributes struct {
	Description string    `json:"description"`
	Downloads   int       `json:"downloads"`
	PublishedAt time.Time `json:"published-at"`
	Tag         string    `json:"tag,omitempty"`
	Version     string    `json:"version"`
}

// ProviderDocDetails represents detailed provider documentation
type ProviderDocDetails struct {
	Data ProviderDocData `json:"data"`
}

// ProviderDocData represents provider documentation data
type ProviderDocData struct {
	Type       string        `json:"type"`
	ID         string        `json:"id"`
	Attributes DocAttributes `json:"attributes"`
	Links      SelfLink      `json:"links"`
}

// DocAttributes represents documentation attributes
type DocAttributes struct {
	Category    string `json:"category"`
	Content     string `json:"content"`
	Language    string `json:"language"`
	Path        string `json:"path"`
	Slug        string `json:"slug"`
	Subcategory string `json:"subcategory,omitempty"`
	Title       string `json:"title"`
	Truncated   bool   `json:"truncated"`
}

// Module represents a Terraform module
type Module struct {
	ID          string    `json:"id"`
	Owner       string    `json:"owner"`
	Namespace   string    `json:"namespace"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Provider    string    `json:"provider"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Tag         string    `json:"tag,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	Downloads   int64     `json:"downloads"`
	Verified    bool      `json:"verified"`
}

// ModuleList represents a paginated list of modules
type ModuleList struct {
	Meta    ModuleMeta `json:"meta"`
	Modules []Module   `json:"modules"`
}

// ModuleMeta represents module list metadata
type ModuleMeta struct {
	Limit         int    `json:"limit"`
	CurrentOffset int    `json:"current_offset"`
	NextOffset    int    `json:"next_offset,omitempty"`
	PrevOffset    int    `json:"prev_offset,omitempty"`
	NextURL       string `json:"next_url,omitempty"`
	PrevURL       string `json:"prev_url,omitempty"`
}

// ModuleDetails represents detailed information about a module version
type ModuleDetails struct {
	Module
	ProviderLogoURL string          `json:"provider_logo_url,omitempty"`
	Root            ModulePart      `json:"root"`
	Submodules      []ModulePart    `json:"submodules,omitempty"`
	Examples        []ModulePart    `json:"examples,omitempty"`
	Providers       []string        `json:"providers,omitempty"`
	Versions        []string        `json:"versions,omitempty"`
	Deprecation     json.RawMessage `json:"deprecation,omitempty"`
}

// ModulePart represents a part of a module (root, submodule, or example)
type ModulePart struct {
	Path                 string                     `json:"path"`
	Name                 string                     `json:"name"`
	Readme               string                     `json:"readme,omitempty"`
	Empty                bool                       `json:"empty"`
	Inputs               []ModuleInput              `json:"inputs,omitempty"`
	Outputs              []ModuleOutput             `json:"outputs,omitempty"`
	Dependencies         []ModuleDependency         `json:"dependencies,omitempty"`
	ProviderDependencies []ModuleProviderDependency `json:"provider_dependencies,omitempty"`
	Resources            []ModuleResource           `json:"resources,omitempty"`
}

// ModuleInput represents a module input variable
type ModuleInput struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Default     json.RawMessage `json:"default,omitempty"`
	Required    bool            `json:"required"`
}

// ModuleOutput represents a module output value
type ModuleOutput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ModuleDependency represents a module dependency
type ModuleDependency struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

// ModuleProviderDependency represents a provider dependency
type ModuleProviderDependency struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Source    string `json:"source"`
	Version   string `json:"version"`
}

// ModuleResource represents a resource in a module
type ModuleResource struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Policy represents a Terraform policy
type Policy struct {
	Type          string              `json:"type"`
	ID            string              `json:"id"`
	Attributes    PolicyAttributes    `json:"attributes"`
	Relationships PolicyRelationships `json:"relationships"`
	Links         SelfLink            `json:"links"`
}

// PolicyAttributes represents policy attributes
type PolicyAttributes struct {
	Downloads int    `json:"downloads"`
	FullName  string `json:"full-name"`
	Ingress   string `json:"ingress"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	OwnerName string `json:"owner-name"`
	Source    string `json:"source"`
	Title     string `json:"title"`
	Verified  bool   `json:"verified"`
}

// PolicyRelationships represents policy relationships
type PolicyRelationships struct {
	LatestVersion LatestVersionRelation `json:"latest-version"`
}

// LatestVersionRelation represents the latest version relationship
type LatestVersionRelation struct {
	Data  ResourceIdentifier `json:"data"`
	Links RelatedLink        `json:"links"`
}

// RelatedLink represents a related link
type RelatedLink struct {
	Related string `json:"related"`
}

// PolicyList represents a paginated list of policies
type PolicyList struct {
	Data     []Policy                `json:"data"`
	Included []PolicyVersionIncluded `json:"included,omitempty"`
	Links    Links                   `json:"links"`
	Meta     Meta                    `json:"meta"`
}

// PolicyVersionIncluded represents included policy version data
type PolicyVersionIncluded struct {
	Type       string                  `json:"type"`
	ID         string                  `json:"id"`
	Attributes PolicyVersionAttributes `json:"attributes"`
	Links      SelfLink                `json:"links"`
}

// PolicyVersionAttributes represents policy version attributes
type PolicyVersionAttributes struct {
	Description string    `json:"description"`
	Downloads   int       `json:"downloads"`
	PublishedAt time.Time `json:"published-at"`
	Readme      string    `json:"readme,omitempty"`
	Source      string    `json:"source"`
	Tag         string    `json:"tag,omitempty"`
	Version     string    `json:"version"`
}

// PolicyDetails represents detailed policy information
type PolicyDetails struct {
	Data     PolicyDetailData `json:"data"`
	Included []PolicyIncluded `json:"included,omitempty"`
}

// PolicyDetailData represents policy detail data
type PolicyDetailData struct {
	Type          string                    `json:"type"`
	ID            string                    `json:"id"`
	Attributes    PolicyVersionAttributes   `json:"attributes"`
	Relationships PolicyDetailRelationships `json:"relationships"`
	Links         SelfLink                  `json:"links"`
}

// PolicyDetailRelationships represents policy detail relationships
type PolicyDetailRelationships struct {
	Policies      RelationshipData      `json:"policies"`
	PolicyLibrary PolicyLibraryRelation `json:"policy-library"`
	PolicyModules RelationshipData      `json:"policy-modules"`
}

// PolicyLibraryRelation represents policy library relationship
type PolicyLibraryRelation struct {
	Data ResourceIdentifier `json:"data"`
}

// PolicyIncluded represents included policy data
type PolicyIncluded struct {
	Type       string              `json:"type"`
	ID         string              `json:"id"`
	Attributes PolicyIncludedAttrs `json:"attributes"`
	Links      SelfLink            `json:"links"`
}

// PolicyIncludedAttrs represents included policy attributes
type PolicyIncludedAttrs struct {
	Description string `json:"description"`
	Downloads   int    `json:"downloads"`
	FullName    string `json:"full-name"`
	Name        string `json:"name"`
	Shasum      string `json:"shasum"`
	ShasumType  string `json:"shasum-type"`
	Title       string `json:"title"`
}

// Common types

// Links represents pagination links
type Links struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
}

// Meta represents metadata
type Meta struct {
	Pagination Pagination `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	PageSize    int `json:"page-size"`
	CurrentPage int `json:"current-page"`
	NextPage    int `json:"next-page,omitempty"`
	PrevPage    int `json:"prev-page,omitempty"`
	TotalPages  int `json:"total-pages"`
	TotalCount  int `json:"total-count"`
}

// SelfLink represents a self link
type SelfLink struct {
	Self string `json:"self"`
}

// Time format constants
const (
	// TimeFormat is the format used by the Terraform Registry API
	TimeFormat = time.RFC3339
)
