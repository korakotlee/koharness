// Package lint provides quality assurance validators for repository assets,
// verifying JSON/YAML syntax, script executable permissions, and skill metadata schemas.
package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LintIssue represents a single linting error or warning found during repository inspection.
type LintIssue struct {
	// Path is the relative or absolute filepath where the issue was detected.
	Path string
	// Category describes the type of lint check (e.g., "JSON Syntax", "Executable Bit", "Skill Metadata").
	Category string
	// Message contains a detailed description of the error or non-compliance.
	Message string
}

// (Issue formatting helper)
func (i LintIssue) String() string {
	return fmt.Sprintf("[%s] %s: %s", i.Category, i.Path, i.Message)
}

// ValidateSyntax inspects all JSON and YAML files inside the specified subdirectories
// (such as "mcp" and "harnesses") within repoRoot, reporting syntax error issues.
func ValidateSyntax(repoRoot string, subdirs []string) ([]LintIssue, error) {
	var issues []LintIssue

	for _, dir := range subdirs {
		targetDir := filepath.Join(repoRoot, dir)
		info, err := os.Stat(targetDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed inspecting directory %s: %w", targetDir, err)
		}
		if !info.IsDir() {
			continue
		}

		err = filepath.Walk(targetDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if info.IsDir() {
				return nil
			}

			relPath, relErr := filepath.Rel(repoRoot, path)
			if relErr != nil {
				relPath = path
			}

			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".json":
				if issue := validateJSONFile(path, relPath); issue != nil {
					issues = append(issues, *issue)
				}
			case ".yaml", ".yml":
				if issue := validateYAMLFile(path, relPath); issue != nil {
					issues = append(issues, *issue)
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking directory %s: %w", targetDir, err)
		}
	}

	return issues, nil
}

func validateJSONFile(fullPath, displayPath string) *LintIssue {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return &LintIssue{
			Path:     displayPath,
			Category: "JSON Syntax",
			Message:  fmt.Sprintf("unable to read file: %v", err),
		}
	}

	var raw json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return &LintIssue{
			Path:     displayPath,
			Category: "JSON Syntax",
			Message:  fmt.Sprintf("invalid JSON syntax: %v", err),
		}
	}
	return nil
}

func validateYAMLFile(fullPath, displayPath string) *LintIssue {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return &LintIssue{
			Path:     displayPath,
			Category: "YAML Syntax",
			Message:  fmt.Sprintf("unable to read file: %v", err),
		}
	}

	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return &LintIssue{
			Path:     displayPath,
			Category: "YAML Syntax",
			Message:  fmt.Sprintf("invalid YAML syntax: %v", err),
		}
	}
	return nil
}
