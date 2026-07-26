package harvester_test

import (
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harvester"
	"github.com/spf13/afero"
)

func TestScanner_ScanAll(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/testuser"

	// Create mock Antigravity skills and workflows
	agSkillsDir := filepath.Join(homeDir, ".gemini", "config", "skills")
	agWorkflowsDir := filepath.Join(homeDir, ".gemini", "config", "global_workflows")
	_ = fs.MkdirAll(filepath.Join(agSkillsDir, "a11y"), 0755)
	_ = fs.MkdirAll(filepath.Join(agSkillsDir, "code-review"), 0755)
	_ = fs.MkdirAll(agWorkflowsDir, 0755)
	_ = afero.WriteFile(fs, filepath.Join(agWorkflowsDir, "commit.md"), []byte("commit workflow"), 0644)

	// Create mock Claude JSON with mcpServers
	claudeJson := filepath.Join(homeDir, ".claude.json")
	claudeContent := `{
		"mcpServers": {
			"github": {"command": "npx"},
			"postgres": {"command": "docker"}
		}
	}`
	_ = afero.WriteFile(fs, claudeJson, []byte(claudeContent), 0644)

	scanner, err := harvester.NewScanner(
		harvester.WithFs(fs),
		harvester.WithHomeDir(homeDir),
	)
	if err != nil {
		t.Fatalf("unexpected error creating scanner: %v", err)
	}

	caps, err := scanner.ScanAll()
	if err != nil {
		t.Fatalf("unexpected error during ScanAll: %v", err)
	}

	if len(caps) < 4 {
		t.Fatalf("expected at least 4 capabilities, found %d", len(caps))
	}

	foundMap := make(map[string]bool)
	for _, c := range caps {
		foundMap[c.Name] = true
	}

	expectedNames := []string{"a11y", "code-review", "commit.md", "github", "postgres"}
	for _, name := range expectedNames {
		if !foundMap[name] {
			t.Errorf("expected capability %q not found in scan results", name)
		}
	}
}

func TestScanner_EmptyHomeDir(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/emptyuser"

	scanner, err := harvester.NewScanner(
		harvester.WithFs(fs),
		harvester.WithHomeDir(homeDir),
	)
	if err != nil {
		t.Fatalf("unexpected error creating scanner: %v", err)
	}

	caps, err := scanner.ScanAll()
	if err != nil {
		t.Fatalf("unexpected error during ScanAll: %v", err)
	}

	if len(caps) != 0 {
		t.Errorf("expected 0 capabilities for empty home dir, got %d", len(caps))
	}
}
