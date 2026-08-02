package lint_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/korakotlee/koharness/pkg/lint"
)

func TestValidateSyntax(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Setup valid and invalid JSON/YAML files
	mcpDir := filepath.Join(tmpDir, "mcp")
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		t.Fatalf("failed creating mcp dir: %v", err)
	}

	validJSON := filepath.Join(mcpDir, "valid.json")
	if err := os.WriteFile(validJSON, []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatalf("failed writing valid json: %v", err)
	}

	invalidJSON := filepath.Join(mcpDir, "invalid.json")
	if err := os.WriteFile(invalidJSON, []byte(`{"name": test}`), 0644); err != nil {
		t.Fatalf("failed writing invalid json: %v", err)
	}

	validYAML := filepath.Join(mcpDir, "valid.yaml")
	if err := os.WriteFile(validYAML, []byte("key: value\n"), 0644); err != nil {
		t.Fatalf("failed writing valid yaml: %v", err)
	}

	invalidYAML := filepath.Join(mcpDir, "invalid.yaml")
	if err := os.WriteFile(invalidYAML, []byte("key: [unclosed list\n"), 0644); err != nil {
		t.Fatalf("failed writing invalid yaml: %v", err)
	}

	issues, err := lint.ValidateSyntax(tmpDir, []string{"mcp"})
	if err != nil {
		t.Fatalf("unexpected error running ValidateSyntax: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("expected 2 syntax issues, got %d", len(issues))
	}
}

func TestValidateScriptPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping POSIX permission test on Windows")
	}

	tmpDir := t.TempDir()

	scriptDir := filepath.Join(tmpDir, "skills", "my-skill", "scripts")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		t.Fatalf("failed creating scripts dir: %v", err)
	}

	execScript := filepath.Join(scriptDir, "run.sh")
	if err := os.WriteFile(execScript, []byte("#!/bin/sh\necho ok"), 0755); err != nil {
		t.Fatalf("failed writing executable script: %v", err)
	}

	nonExecScript := filepath.Join(scriptDir, "helper.py")
	if err := os.WriteFile(nonExecScript, []byte("print('hi')"), 0644); err != nil {
		t.Fatalf("failed writing non-executable script: %v", err)
	}

	issues, err := lint.ValidateScriptPermissions(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error running ValidateScriptPermissions: %v", err)
	}

	if len(issues) != 1 {
		t.Errorf("expected 1 permission issue, got %d", len(issues))
	}
}

func TestValidateSkillFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	validSkillDir := filepath.Join(tmpDir, "skills", "valid-skill")
	if err := os.MkdirAll(validSkillDir, 0755); err != nil {
		t.Fatalf("failed creating valid skill dir: %v", err)
	}
	validSkillFile := filepath.Join(validSkillDir, "SKILL.md")
	validContent := "---\nname: valid-skill\ndescription: A valid skill\n---\n# Body\n"
	if err := os.WriteFile(validSkillFile, []byte(validContent), 0644); err != nil {
		t.Fatalf("failed writing valid skill: %v", err)
	}

	invalidSkillDir := filepath.Join(tmpDir, "skills", "invalid-skill")
	if err := os.MkdirAll(invalidSkillDir, 0755); err != nil {
		t.Fatalf("failed creating invalid skill dir: %v", err)
	}
	invalidSkillFile := filepath.Join(invalidSkillDir, "SKILL.md")
	invalidContent := "---\nname: invalid-skill\n---\n# Missing description\n"
	if err := os.WriteFile(invalidSkillFile, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed writing invalid skill: %v", err)
	}

	issues, err := lint.ValidateSkillFrontmatter(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error running ValidateSkillFrontmatter: %v", err)
	}

	if len(issues) != 1 {
		t.Errorf("expected 1 frontmatter issue, got %d", len(issues))
	}
}

func TestLintRun(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := lint.Run(lint.LintOptions{RepoRoot: tmpDir})
	if err != nil {
		t.Fatalf("unexpected error running lint.Run: %v", err)
	}
	if result.HasErrors() {
		t.Errorf("expected clean run on empty repo, got issues: %v", result.Issues)
	}
}
