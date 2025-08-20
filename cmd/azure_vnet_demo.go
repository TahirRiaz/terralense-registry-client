package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/TahirRiaz/terralense-registry-client/registry"

	"github.com/sirupsen/logrus"
)

// AzureVNetDemo encapsulates the Azure VNet demo functionality
type AzureVNetDemo struct {
	client *registry.Client
	logger *logrus.Logger
}

// NewAzureVNetDemo creates a new Azure VNet demo instance
func NewAzureVNetDemo(client *registry.Client, logger *logrus.Logger) *AzureVNetDemo {
	return &AzureVNetDemo{
		client: client,
		logger: logger,
	}
}

func (d *AzureVNetDemo) displayModuleInputs(inputs []registry.ModuleInput) {
	// Separate required and optional inputs
	var requiredInputs, optionalInputs []registry.ModuleInput

	for _, input := range inputs {
		if input.Required {
			requiredInputs = append(requiredInputs, input)
		} else {
			optionalInputs = append(optionalInputs, input)
		}
	}

	// Sort by name
	sort.Slice(requiredInputs, func(i, j int) bool {
		return requiredInputs[i].Name < requiredInputs[j].Name
	})
	sort.Slice(optionalInputs, func(i, j int) bool {
		return optionalInputs[i].Name < optionalInputs[j].Name
	})

	// Display required inputs
	if len(requiredInputs) > 0 {
		fmt.Println("\n  Required Inputs:")
		d.displayInputTable(requiredInputs, 5)
	}

	// Display optional inputs (limited)
	if len(optionalInputs) > 0 {
		fmt.Println("\n  Optional Inputs (showing first 5):")
		d.displayInputTable(optionalInputs, 5)

		if len(optionalInputs) > 5 {
			fmt.Printf("  ... and %d more optional inputs\n", len(optionalInputs)-5)
		}
	}
}

func (d *AzureVNetDemo) displayInputTable(inputs []registry.ModuleInput, limit int) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  NAME\tTYPE\tDESCRIPTION")
	fmt.Fprintln(w, "  ----\t----\t-----------")

	count := 0
	for _, input := range inputs {
		if count >= limit {
			break
		}

		desc := input.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		fmt.Fprintf(w, "  %s\t%s\t%s\n", input.Name, input.Type, desc)
		count++
	}
	w.Flush()
}

func (d *AzureVNetDemo) displayModuleOutputs(outputs []registry.ModuleOutput) {
	// Filter for important outputs
	var importantOutputs []registry.ModuleOutput

	for _, output := range outputs {
		nameLower := strings.ToLower(output.Name)
		if strings.Contains(nameLower, "vnet") ||
			strings.Contains(nameLower, "subnet") ||
			strings.Contains(nameLower, "id") ||
			strings.Contains(nameLower, "name") ||
			strings.Contains(nameLower, "address") {
			importantOutputs = append(importantOutputs, output)
		}
	}

	if len(importantOutputs) == 0 {
		importantOutputs = outputs
	}

	// Sort by name
	sort.Slice(importantOutputs, func(i, j int) bool {
		return importantOutputs[i].Name < importantOutputs[j].Name
	})

	// Display outputs
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  NAME\tDESCRIPTION")
	fmt.Fprintln(w, "  ----\t-----------")

	count := 0
	maxOutputs := 10
	for _, output := range importantOutputs {
		if count >= maxOutputs {
			break
		}

		desc := output.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}

		fmt.Fprintf(w, "  %s\t%s\n", output.Name, desc)
		count++
	}
	w.Flush()

	if len(importantOutputs) > maxOutputs {
		fmt.Printf("  ... and %d more outputs\n", len(importantOutputs)-maxOutputs)
	}
}

// Run executes the Azure VNet demo
func (d *AzureVNetDemo) Run(ctx context.Context) error {
	// Step 1: Search for Azure VNet modules
	fmt.Println("\n1. Searching for Azure VNet Terraform Modules")
	fmt.Println(strings.Repeat("-", 50))

	modules, err := d.searchAzureVNetModules(ctx)
	if err != nil {
		return fmt.Errorf("module search failed: %w", err)
	}

	if err := d.displayModuleResults(ctx, modules); err != nil {
		return fmt.Errorf("failed to display module results: %w", err)
	}

	// Step 2: Get Azure provider VNet documentation
	fmt.Println("\n2. Getting Azure Provider VNet Documentation")
	fmt.Println(strings.Repeat("-", 50))

	if err := d.getAzureProviderDocs(ctx); err != nil {
		return fmt.Errorf("provider docs failed: %w", err)
	}

	// Step 3: Get specific module with VNet configuration
	fmt.Println("\n3. Getting Popular Azure VNet Module Example")
	fmt.Println(strings.Repeat("-", 50))

	if err := d.getSpecificVNetModule(ctx); err != nil {
		return fmt.Errorf("specific module failed: %w", err)
	}

	return nil
}

func (d *AzureVNetDemo) searchAzureVNetModules(ctx context.Context) ([]registry.ModuleSearchResult, error) {
	searchQueries := []string{
		"azure vnet",
		"azure virtual network",
		"azurerm vnet",
	}

	var allResults []registry.ModuleSearchResult
	seen := make(map[string]bool)

	for _, query := range searchQueries {
		d.logger.Infof("Searching for: %s", query)

		results, err := d.client.Modules.SearchWithRelevance(ctx, query, 0)
		if err != nil {
			d.logger.Warnf("Search failed for '%s': %v", query, err)
			continue
		}

		// Deduplicate results
		for _, result := range results {
			if !seen[result.ID] {
				seen[result.ID] = true
				allResults = append(allResults, result)
			}
		}
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("no modules found")
	}

	// Sort by relevance
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Relevance > allResults[j].Relevance
	})

	return allResults, nil
}

func (d *AzureVNetDemo) displayModuleResults(ctx context.Context, results []registry.ModuleSearchResult) error {
	fmt.Printf("\nFound %d unique modules. Top 5 results:\n\n", len(results))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MODULE\tVERSION\tDOWNLOADS\tVERIFIED\tRELEVANCE")
	fmt.Fprintln(w, "------\t-------\t---------\t--------\t---------")

	maxResults := 5
	if len(results) < maxResults {
		maxResults = len(results)
	}

	for i := 0; i < maxResults; i++ {
		result := results[i]
		verified := "No"
		if result.Verified {
			verified = "Yes"
		}

		fmt.Fprintf(w, "%s/%s/%s\t%s\t%d\t%s\t%.1f\n",
			result.Namespace, result.Name, result.Provider,
			result.Version, result.Downloads, verified, result.Relevance)
	}
	w.Flush()

	// Get detailed configuration for the top result
	if len(results) > 0 {
		fmt.Printf("\nGetting configuration details for top module...\n")
		module, err := d.client.Modules.GetByID(ctx, results[0].ID)
		if err != nil {
			d.logger.Warnf("Failed to get module details: %v", err)
			return nil
		}

		d.displayModuleConfiguration(module)
	}

	return nil
}

func (d *AzureVNetDemo) displayModuleConfiguration(module *registry.ModuleDetails) {
	fmt.Println("\nModule Configuration:")
	fmt.Println(strings.Repeat("-", 40))

	// Display example configuration if available
	if len(module.Examples) > 0 && module.Examples[0].Readme != "" {
		examples := registry.ExtractTerraformExamples(module.Examples[0].Readme)
		if len(examples) > 0 {
			fmt.Println("Example Usage:")
			fmt.Println("```hcl")
			fmt.Println(examples[0])
			fmt.Println("```")
		}
	}

	// Display key inputs
	if len(module.Root.Inputs) > 0 {
		d.displayKeyInputs(module.Root.Inputs)
	}
}

func (d *AzureVNetDemo) displayKeyInputs(inputs []registry.ModuleInput) {
	fmt.Println("\nKey VNet-related Inputs:")

	// Filter VNet-related inputs
	var vnetInputs []registry.ModuleInput
	for _, input := range inputs {
		nameLower := strings.ToLower(input.Name)
		if input.Required ||
			strings.Contains(nameLower, "vnet") ||
			strings.Contains(nameLower, "subnet") ||
			strings.Contains(nameLower, "address") ||
			input.Name == "name" ||
			input.Name == "location" ||
			input.Name == "resource_group_name" {
			vnetInputs = append(vnetInputs, input)
		}
	}

	// Sort by required first, then by name
	sort.Slice(vnetInputs, func(i, j int) bool {
		if vnetInputs[i].Required != vnetInputs[j].Required {
			return vnetInputs[i].Required
		}
		return vnetInputs[i].Name < vnetInputs[j].Name
	})

	// Display in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tREQUIRED\tDESCRIPTION")
	fmt.Fprintln(w, "----\t----\t--------\t-----------")

	maxInputs := 10
	for i, input := range vnetInputs {
		if i >= maxInputs {
			break
		}

		required := "No"
		if input.Required {
			required = "Yes"
		}

		desc := input.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", input.Name, input.Type, required, desc)
	}
	w.Flush()

	if len(vnetInputs) > maxInputs {
		fmt.Printf("... and %d more inputs\n", len(vnetInputs)-maxInputs)
	}
}

func (d *AzureVNetDemo) getAzureProviderDocs(ctx context.Context) error {
	// Get the Azure provider info
	provider, err := d.client.Providers.Get(ctx, "hashicorp", "azurerm")
	if err != nil {
		return fmt.Errorf("failed to get Azure provider: %w", err)
	}

	fmt.Printf("Azure Provider: %s\n", provider.Attributes.FullName)
	fmt.Printf("Namespace: %s\n", provider.Attributes.Namespace)
	fmt.Printf("Downloads: %d\n", provider.Attributes.Downloads)
	fmt.Printf("Tier: %s\n", provider.Attributes.Tier)

	// Get latest version
	latestInfo, err := d.client.Providers.GetLatest(ctx, "hashicorp", "azurerm")
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	fmt.Printf("Latest Version: %s\n", latestInfo.Version)

	// Get provider version ID
	versionID, err := d.client.Providers.GetVersionID(ctx, "hashicorp", "azurerm", latestInfo.Version)
	if err != nil {
		return fmt.Errorf("failed to get version ID: %w", err)
	}

	// List of VNet-related resources to fetch
	vnetResources := []struct {
		slug string
		name string
	}{
		{"virtual_network", "azurerm_virtual_network"},
		{"subnet", "azurerm_subnet"},
		{"virtual_network_peering", "azurerm_virtual_network_peering"},
	}

	fmt.Println("\nFetching VNet-related resource documentation...")

	for _, resource := range vnetResources {
		fmt.Printf("\n%s:\n", resource.name)

		opts := &registry.ProviderDocListOptions{
			ProviderVersionID: versionID,
			Category:          "resources",
			Slug:              resource.slug,
			Language:          "hcl",
			Page:              1,
		}

		docs, err := d.client.Providers.ListDocsV2(ctx, opts)
		if err != nil {
			d.logger.Warnf("Failed to get docs for %s: %v", resource.name, err)
			fmt.Printf("  ✗ Failed to fetch documentation\n")
			continue
		}

		if len(docs) > 0 {
			fmt.Printf("  ✓ Documentation available\n")

			if resource.slug == "virtual_network" {
				// Get detailed docs for virtual_network
				details, err := d.client.Providers.GetDoc(ctx, docs[0].ID)
				if err != nil {
					d.logger.Warnf("Failed to get doc details: %v", err)
					continue
				}

				d.displayProviderDocumentation(details)
			}
		} else {
			fmt.Printf("  ✗ No documentation found\n")
		}
	}

	return nil
}

func (d *AzureVNetDemo) displayProviderDocumentation(details *registry.ProviderDocDetails) {
	fmt.Println("\nVirtual Network Resource Documentation:")
	fmt.Println(strings.Repeat("-", 40))

	// Extract configuration examples
	examples := registry.ExtractTerraformExamples(details.Data.Attributes.Content)
	if len(examples) > 0 {
		fmt.Println("Configuration Example:")
		fmt.Println("```hcl")
		// Limit example length for display
		example := examples[0]
		if len(example) > 500 {
			example = example[:500] + "\n... (truncated)"
		}
		fmt.Println(example)
		fmt.Println("```")
	}
}

func (d *AzureVNetDemo) getSpecificVNetModule(ctx context.Context) error {
	// Try to find popular Azure networking modules
	knownModules := []struct {
		namespace string
		name      string
		provider  string
	}{
		{"Azure", "vnet", "azurerm"},
		{"Azure", "network", "azurerm"},
		{"terraform-azurerm-modules", "terraform-azurerm-vnet", "azurerm"},
	}

	var module *registry.ModuleDetails
	var moduleErr error

	for _, km := range knownModules {
		d.logger.Debugf("Checking module: %s/%s/%s", km.namespace, km.name, km.provider)

		module, moduleErr = d.client.Modules.GetLatest(ctx, km.namespace, km.name, km.provider)
		if moduleErr == nil {
			fmt.Printf("✓ Found module: %s/%s/%s\n", km.namespace, km.name, km.provider)
			break
		}

		if registry.IsNotFound(moduleErr) {
			fmt.Printf("✗ Module not found: %s/%s/%s\n", km.namespace, km.name, km.provider)
		} else {
			fmt.Printf("✗ Error: %v\n", moduleErr)
		}
	}

	if module == nil {
		// Fallback: search for any Azure VNet module
		fmt.Println("\nSearching for any Azure VNet module...")
		results, err := d.client.Modules.SearchWithRelevance(ctx, "azure vnet", 0)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(results) > 0 {
			// Get the first verified module or just the first one
			for _, result := range results {
				if result.Verified && strings.Contains(strings.ToLower(result.Name), "vnet") {
					module, moduleErr = d.client.Modules.GetByID(ctx, result.ID)
					if moduleErr == nil {
						break
					}
				}
			}

			// If no verified module found, use the first one
			if module == nil && len(results) > 0 {
				module, moduleErr = d.client.Modules.GetByID(ctx, results[0].ID)
			}
		}
	}

	if module == nil {
		return fmt.Errorf("could not find any Azure VNet module")
	}

	// Display module information
	d.displayModuleDetails(module)

	return nil
}

func (d *AzureVNetDemo) displayModuleDetails(module *registry.ModuleDetails) {
	fmt.Printf("\nModule: %s\n", module.ID)
	fmt.Printf("Source: %s\n", module.Source)
	fmt.Printf("Version: %s\n", module.Version)
	fmt.Printf("Downloads: %d\n", module.Downloads)
	fmt.Printf("Verified: %v\n", module.Verified)

	if module.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", module.Description)
	}

	// Display basic usage
	fmt.Println("\nBasic Usage:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf(`module "vnet" {
  source  = "%s"
  version = "%s"

  # Add your configuration here
  # See module inputs below for required and optional variables
}
`, module.Source, module.Version)

	// Display inputs
	if len(module.Root.Inputs) > 0 {
		fmt.Println("\nModule Inputs:")
		d.displayModuleInputs(module.Root.Inputs)
	}

	// Display outputs
	if len(module.Root.Outputs) > 0 {
		fmt.Println("\nModule Outputs:")
		d.displayModuleOutputs(module.Root.Outputs)
	}
}
