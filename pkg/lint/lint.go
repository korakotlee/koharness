package lint

import (
	"fmt"
	"os"
)

// LintOptions defines options for running repository quality assurance checks.
type LintOptions struct {
	// RepoRoot is the root directory of the repository to lint.
	RepoRoot string
	// Subdirs specifies subdirectories for syntax checking (default: ["mcp", "harnesses"]).
	Subdirs []string
}

// LintResult contains all issues identified during a linting run.
type LintResult struct {
	Issues []LintIssue
}

// HasErrors returns true if any issues were detected.
func (r *LintResult) HasErrors() bool {
	return len(r.Issues) > 0
}

// Run executes all repository lint checks and returns a summary result.
func Run(opts LintOptions) (*LintResult, error) {
	root := opts.RepoRoot
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("unable to determine current working directory: %w", err)
		}
	}

	subdirs := opts.Subdirs
	if len(subdirs) == 0 {
		subdirs = []string{"mcp", "harnesses"}
	}

	var allIssues []LintIssue

	// 1. JSON & YAML Syntax
	syntaxIssues, err := ValidateSyntax(root, subdirs)
	if err != nil {
		return nil, fmt.Errorf("syntax lint failed: %w", err)
	}
	allIssues = append(allIssues, syntaxIssues...)

	// 2. Script Permissions
	permIssues, err := ValidateScriptPermissions(root)
	if err != nil {
		return nil, fmt.Errorf("permissions lint failed: %w", err)
	}
	allIssues = append(allIssues, permIssues...)

	// 3. Skill Frontmatter
	frontmatterIssues, err := ValidateSkillFrontmatter(root)
	if err != nil {
		return nil, fmt.Errorf("frontmatter lint failed: %w", err)
	}
	allIssues = append(allIssues, frontmatterIssues...)

	return &LintResult{Issues: allIssues}, nil
}
