package mcp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// IssueType categorizes configuration safety warnings.
type IssueType string

const (
	// IssueNonPortablePath indicates a hardcoded absolute user home path.
	IssueNonPortablePath IssueType = "NonPortablePath"
	// IssuePossibleSecret indicates an un-expanded plain text API key or secret token.
	IssuePossibleSecret IssueType = "PossibleSecret"
)

// ValidationIssue describes a discovered portability or security issue in an MCP config.
type ValidationIssue struct {
	Type    IssueType `json:"type"`
	KeyPath string    `json:"key_path"`
	Value   string    `json:"value"`
	Message string    `json:"message"`
}

var (
	nonPortablePattern = regexp.MustCompile(`^(/Users/|/home/|[A-Za-z]:\\Users\\)`)
	secretPattern      = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|auth|bearer)`)
	secretValuePattern = regexp.MustCompile(`^(sk-[A-Za-z0-9]{20,}|ghp_[A-Za-z0-9]{20,}|eyJ[A-Za-z0-9_-]{20,})`)
)

// MaskSecrets replaces sensitive tokens or secret values with masked placeholders (e.g., "sk-***" or "***").
func MaskSecrets(s string) string {
	if s == "" {
		return ""
	}
	masked := secretValuePattern.ReplaceAllStringFunc(s, func(match string) string {
		if strings.HasPrefix(match, "sk-") {
			return "sk-***"
		}
		if strings.HasPrefix(match, "ghp_") {
			return "ghp_***"
		}
		if strings.HasPrefix(match, "eyJ") {
			return "eyJ***"
		}
		return "***"
	})
	return masked
}

// ValidateConfig inspects JSON configuration data for non-portable paths and hardcoded secrets.
func ValidateConfig(data []byte) ([]ValidationIssue, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var val interface{}
	if err := json.Unmarshal(data, &val); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON for validation: %w", err)
	}

	var issues []ValidationIssue
	inspectValue(val, "", &issues)
	return issues, nil
}

func inspectValue(val interface{}, currentPath string, issues *[]ValidationIssue) {
	switch v := val.(type) {
	case string:
		// Check for non-portable absolute user paths
		if nonPortablePattern.MatchString(v) {
			*issues = append(*issues, ValidationIssue{
				Type:    IssueNonPortablePath,
				KeyPath: currentPath,
				Value:   v,
				Message: fmt.Sprintf("Hardcoded user home path '%s' at %s is non-portable across machines. Use ${HOME} or move to mcp.local.json.", v, currentPath),
			})
		}
		// Check for hardcoded secret values (e.g. API tokens) if key name suggests secret or value looks like token
		if secretValuePattern.MatchString(v) {
			masked := MaskSecrets(v)
			*issues = append(*issues, ValidationIssue{
				Type:    IssuePossibleSecret,
				KeyPath: currentPath,
				Value:   masked,
				Message: fmt.Sprintf("Possible raw API secret key detected at %s (%s). Use environment variables or move to mcp.local.json.", currentPath, masked),
			})
		}
	case map[string]interface{}:
		for k, child := range v {
			nextPath := k
			if currentPath != "" {
				nextPath = currentPath + "." + k
			}
			// Check if key name indicates secret and child is non-empty literal string
			if strVal, ok := child.(string); ok && strVal != "" {
				if secretPattern.MatchString(k) && !strings.HasPrefix(strVal, "${") {
					masked := MaskSecrets(strVal)
					if masked == strVal {
						masked = "***"
					}
					*issues = append(*issues, ValidationIssue{
						Type:    IssuePossibleSecret,
						KeyPath: nextPath,
						Value:   masked,
						Message: fmt.Sprintf("Key '%s' contains plain text secret value. Use environment variables or mcp.local.json.", nextPath),
					})
				}
			}
			inspectValue(child, nextPath, issues)
		}
	case []interface{}:
		for i, child := range v {
			nextPath := fmt.Sprintf("%s[%d]", currentPath, i)
			inspectValue(child, nextPath, issues)
		}
	}
}
