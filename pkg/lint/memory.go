package lint

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/korakotlee/koharness/pkg/memory"
)

// ValidateMemory checks the integrity of the 3-layer agent memory architecture,
// verifying Markdown link validity and AGENTS.md trigger target existence.
func ValidateMemory(repoRoot string) ([]LintIssue, error) {
	memDir := filepath.Join(repoRoot, "memory")
	if _, err := os.Stat(memDir); os.IsNotExist(err) {
		// Memory layer is optional for standard dotfiles repositories
		return nil, nil
	}

	var issues []LintIssue

	// 1. Memory layout integrity check
	status, err := memory.InspectMemoryLayout(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed inspecting memory layout: %w", err)
	}

	for _, item := range status.MissingItems {
		issues = append(issues, LintIssue{
			Path:     item,
			Category: "Agent Memory",
			Message:  fmt.Sprintf("missing 3-layer memory component: %s", item),
		})
	}

	// 2. Validate wiki Markdown links
	wikiDir := filepath.Join(memDir, "wiki")
	if info, err := os.Stat(wikiDir); err == nil && info.IsDir() {
		_ = filepath.Walk(wikiDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() || filepath.Ext(path) != ".md" {
				return nil
			}

			relPath, _ := filepath.Rel(repoRoot, path)
			broken, bErr := memory.ValidateMarkdownLinks(path)
			if bErr == nil {
				for _, b := range broken {
					issues = append(issues, LintIssue{
						Path:     relPath,
						Category: "Agent Memory Link",
						Message:  fmt.Sprintf("line %d: broken link [%s](%s) target '%s' not found", b.LineNumber, b.Text, b.Target, b.ResolvedPath),
					})
				}
			}
			return nil
		})
	}

	// 3. Validate AGENTS.md trigger paths
	agentsPath := filepath.Join(memDir, "AGENTS.md")
	if info, err := os.Stat(agentsPath); err == nil && !info.IsDir() {
		invalid, tErr := memory.ValidateTriggerPaths(repoRoot, agentsPath)
		if tErr == nil {
			relAgents, _ := filepath.Rel(repoRoot, agentsPath)
			for _, inv := range invalid {
				issues = append(issues, LintIssue{
					Path:     relAgents,
					Category: "Agent Memory Trigger",
					Message:  fmt.Sprintf("line %d: referenced trigger target '%s' does not exist", inv.LineNumber, inv.TargetPath),
				})
			}
		}
	}

	return issues, nil
}
