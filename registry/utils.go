package registry

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Semantic version regex pattern
	semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9\-\.]+))?(?:\+([a-zA-Z0-9\-\.]+))?$`)

	// Valid namespace/name pattern
	validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_]*$`)

	// Valid provider name pattern (lowercase with hyphens)
	validProviderPattern = regexp.MustCompile(`^[a-z][a-z0-9\-]*$`)
)

// ValidateProviderVersion validates a provider version string
func ValidateProviderVersion(version string) error {
	if version == "" || version == "latest" {
		return nil
	}

	if !semverRegex.MatchString(version) {
		return fmt.Errorf("invalid semantic version format: %s", version)
	}

	return nil
}

// ValidateProviderDataType validates a provider data type
func ValidateProviderDataType(dataType string) error {
	validTypes := []string{"resources", "data-sources", "functions", "guides", "overview"}

	for _, valid := range validTypes {
		if dataType == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid provider data type: %s, must be one of: %s",
		dataType, strings.Join(validTypes, ", "))
}

// IsV2DataType returns true if the data type requires v2 API
func IsV2DataType(dataType string) bool {
	v2Types := []string{"guides", "functions", "overview"}

	for _, v2Type := range v2Types {
		if dataType == v2Type {
			return true
		}
	}

	return false
}

// ExtractProviderInfo extracts namespace, name, and version from a provider URI
func ExtractProviderInfo(uri string) (namespace, name, version string, err error) {
	if uri == "" {
		err = fmt.Errorf("provider URI cannot be empty")
		return
	}

	// Remove any protocol prefix
	uri = strings.TrimPrefix(uri, "registry://")
	uri = strings.TrimPrefix(uri, "providers/")
	uri = strings.TrimSpace(uri)

	parts := strings.Split(uri, "/")

	if len(parts) < 2 {
		err = fmt.Errorf("invalid provider URI format: %s, expected at least namespace/name", uri)
		return
	}

	namespace = parts[0]
	if namespace == "" {
		err = fmt.Errorf("namespace cannot be empty in URI: %s", uri)
		return
	}

	// Handle different URI formats
	switch len(parts) {
	case 2:
		// Format: namespace/name
		name = parts[1]
	case 3:
		// Format: namespace/name/version
		name = parts[1]
		version = parts[2]
	default:
		// Format: namespace/name/version or namespace/providers/name/versions/version
		if parts[1] == "providers" || parts[1] == "name" {
			if len(parts) > 2 {
				name = parts[2]
			}
			if len(parts) > 4 && (parts[3] == "versions" || parts[3] == "version") {
				version = parts[4]
			}
		} else {
			name = parts[1]
			if len(parts) > 2 {
				version = parts[2]
			}
		}
	}

	if name == "" {
		err = fmt.Errorf("name cannot be empty in URI: %s", uri)
		return
	}

	// Validate extracted values
	if !validNamePattern.MatchString(namespace) {
		err = fmt.Errorf("invalid namespace format in URI: %s", namespace)
		return
	}

	if !validProviderPattern.MatchString(name) {
		err = fmt.Errorf("invalid provider name format in URI: %s", name)
		return
	}

	if version != "" && version != "latest" {
		if err = ValidateProviderVersion(version); err != nil {
			return
		}
	}

	return
}

// ParseModuleID parses a module ID into its components
func ParseModuleID(moduleID string) (namespace, name, provider, version string, err error) {
	if moduleID == "" {
		err = fmt.Errorf("module ID cannot be empty")
		return
	}

	parts := strings.Split(moduleID, "/")

	if len(parts) != 4 {
		err = fmt.Errorf("invalid module ID format: %s, expected namespace/name/provider/version", moduleID)
		return
	}

	namespace = strings.TrimSpace(parts[0])
	name = strings.TrimSpace(parts[1])
	provider = strings.TrimSpace(parts[2])
	version = strings.TrimSpace(parts[3])

	// Validate components
	if namespace == "" || name == "" || provider == "" || version == "" {
		err = fmt.Errorf("module ID components cannot be empty: %s", moduleID)
		return
	}

	if !validNamePattern.MatchString(namespace) {
		err = fmt.Errorf("invalid namespace format: %s", namespace)
		return
	}

	if !validNamePattern.MatchString(name) {
		err = fmt.Errorf("invalid module name format: %s", name)
		return
	}

	if !validProviderPattern.MatchString(provider) {
		err = fmt.Errorf("invalid provider format: %s", provider)
		return
	}

	if err = ValidateProviderVersion(version); err != nil {
		return
	}

	return
}

// ParsePolicyID parses a policy ID into its components
func ParsePolicyID(policyID string) (namespace, name, version string, err error) {
	if policyID == "" {
		err = fmt.Errorf("policy ID cannot be empty")
		return
	}

	// Remove "policies/" prefix if present
	policyID = strings.TrimPrefix(policyID, "policies/")
	policyID = strings.TrimSpace(policyID)

	parts := strings.Split(policyID, "/")

	if len(parts) != 3 {
		err = fmt.Errorf("invalid policy ID format: %s, expected namespace/name/version", policyID)
		return
	}

	namespace = strings.TrimSpace(parts[0])
	name = strings.TrimSpace(parts[1])
	version = strings.TrimSpace(parts[2])

	// Validate components
	if namespace == "" || name == "" || version == "" {
		err = fmt.Errorf("policy ID components cannot be empty: %s", policyID)
		return
	}

	if !validNamePattern.MatchString(namespace) {
		err = fmt.Errorf("invalid namespace format: %s", namespace)
		return
	}

	if !validNamePattern.MatchString(name) {
		err = fmt.Errorf("invalid policy name format: %s", name)
		return
	}

	if err = ValidateProviderVersion(version); err != nil {
		return
	}

	return
}

// ExtractContentDescription extracts a description from markdown content
func ExtractContentDescription(content string, maxLength int) string {
	if content == "" {
		return ""
	}

	if maxLength <= 0 {
		maxLength = 200 // Default max length
	}

	// Try to extract description from frontmatter
	if idx := strings.Index(content, "description: |-"); idx != -1 {
		start := idx + len("description: |-")
		end := strings.Index(content[start:], "\n---")
		if end == -1 {
			end = len(content[start:])
		}

		desc := strings.TrimSpace(content[start : start+end])
		desc = strings.ReplaceAll(desc, "\n", " ")
		desc = strings.ReplaceAll(desc, "  ", " ") // Remove double spaces

		return truncateString(desc, maxLength)
	}

	// Fallback: use first paragraph
	lines := strings.Split(content, "\n")
	var desc strings.Builder
	inCodeBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// Stop at empty line (paragraph break)
		if line == "" && desc.Len() > 0 {
			break
		}

		// Skip headers and metadata
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") {
			if desc.Len() > 0 {
				desc.WriteString(" ")
			}
			desc.WriteString(line)
		}
	}

	return truncateString(desc.String(), maxLength)
}

// ExtractReadmeSection extracts the first section from a README
func ExtractReadmeSection(readme string) string {
	if readme == "" {
		return ""
	}

	var builder strings.Builder
	headerFound := false
	lines := strings.Split(readme, "\n")
	headerRegex := regexp.MustCompile(`^#+\s`)
	inCodeBlock := false

	for _, line := range lines {
		// Track code blocks
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
		}

		if !inCodeBlock && headerRegex.MatchString(line) {
			if headerFound {
				// Stop at the second header
				break
			}
			headerFound = true
		}

		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return strings.TrimSuffix(builder.String(), "\n")
}

// NormalizeVersion removes the 'v' prefix from version strings if present
func NormalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

// CompareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Normalize versions
	v1 = NormalizeVersion(v1)
	v2 = NormalizeVersion(v2)

	// Parse versions
	v1Parts := parseSemanticVersion(v1)
	v2Parts := parseSemanticVersion(v2)

	// Compare major, minor, patch
	for i := 0; i < 3; i++ {
		if v1Parts[i] < v2Parts[i] {
			return -1
		}
		if v1Parts[i] > v2Parts[i] {
			return 1
		}
	}

	// Compare pre-release versions
	v1Pre := extractPreRelease(v1)
	v2Pre := extractPreRelease(v2)

	// No pre-release version is greater than a pre-release version
	if v1Pre == "" && v2Pre != "" {
		return 1
	}
	if v1Pre != "" && v2Pre == "" {
		return -1
	}

	// Compare pre-release versions lexically
	if v1Pre < v2Pre {
		return -1
	}
	if v1Pre > v2Pre {
		return 1
	}

	return 0
}

// parseSemanticVersion parses a semantic version string into major, minor, patch
func parseSemanticVersion(version string) [3]int {
	result := [3]int{0, 0, 0}

	matches := semverRegex.FindStringSubmatch(version)
	if len(matches) >= 4 {
		result[0], _ = strconv.Atoi(matches[1])
		result[1], _ = strconv.Atoi(matches[2])
		result[2], _ = strconv.Atoi(matches[3])
	}

	return result
}

// extractPreRelease extracts the pre-release part of a version
func extractPreRelease(version string) string {
	matches := semverRegex.FindStringSubmatch(version)
	if len(matches) >= 5 {
		return matches[4]
	}
	return ""
}

// truncateString truncates a string to the specified length, adding ellipsis if needed
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	// Try to break at a word boundary
	truncated := s[:maxLength]
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLength*3/4 { // If space is in the last quarter
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	return fmt.Sprintf("%.0f days", d.Hours()/24)
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(s string) string {
	// Remove null bytes and other control characters
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, s)

	return strings.TrimSpace(sanitized)
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(s string) bool {
	if s == "" {
		return false
	}

	// Must start with http:// or https://
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return false
	}

	// Basic URL validation
	parts := strings.SplitN(s, "://", 2)
	if len(parts) != 2 {
		return false
	}

	// Check for valid characters in the domain part
	domain := parts[1]
	if domain == "" {
		return false
	}

	// Split by first slash to get domain
	domainParts := strings.SplitN(domain, "/", 2)
	domain = domainParts[0]

	// Domain must contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// ExtractTerraformExamples extracts Terraform code examples from content
func ExtractTerraformExamples(content string) []string {
	examples := []string{}

	// Look for code blocks
	codeBlockRegex := regexp.MustCompile("(?s)```(?:hcl|terraform)?\\s*\n(.*?)```")
	matches := codeBlockRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			code := strings.TrimSpace(match[1])
			// Only include if it's a meaningful example (has resource or module blocks)
			if strings.Contains(code, "resource") || strings.Contains(code, "module") {
				examples = append(examples, code)
			}
		}
	}

	return examples
}
