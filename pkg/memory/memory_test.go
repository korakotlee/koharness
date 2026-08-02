package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectMemoryLayout(t *testing.T) {
	tmpDir := t.TempDir()

	// Initially missing memory/
	status, err := InspectMemoryLayout(tmpDir)
	if err != nil {
		t.Fatalf("InspectMemoryLayout failed: %v", err)
	}
	if status.IsComplete() {
		t.Errorf("expected IsComplete() to be false for empty dir")
	}

	// Create full layout
	memDir := filepath.Join(tmpDir, "memory")
	os.MkdirAll(filepath.Join(memDir, "raw"), 0755)
	os.MkdirAll(filepath.Join(memDir, "wiki"), 0755)
	os.WriteFile(filepath.Join(memDir, "AGENTS.md"), []byte("# Steering"), 0644)
	os.WriteFile(filepath.Join(memDir, "wiki", "INDEX.md"), []byte("# Index"), 0644)

	status, err = InspectMemoryLayout(tmpDir)
	if err != nil {
		t.Fatalf("InspectMemoryLayout failed: %v", err)
	}
	if !status.IsComplete() {
		t.Errorf("expected IsComplete() to be true when all files exist, missing: %v", status.MissingItems)
	}
}

func TestValidateMarkdownLinks(t *testing.T) {
	tmpDir := t.TempDir()

	targetFile := filepath.Join(tmpDir, "target.md")
	os.WriteFile(targetFile, []byte("# Target"), 0644)

	indexContent := "# Index\n- [Valid](target.md)\n- [Broken](missing.md)\n- [Web](https://example.com)\n"
	indexFile := filepath.Join(tmpDir, "INDEX.md")
	os.WriteFile(indexFile, []byte(indexContent), 0644)

	broken, err := ValidateMarkdownLinks(indexFile)
	if err != nil {
		t.Fatalf("ValidateMarkdownLinks failed: %v", err)
	}

	if len(broken) != 1 {
		t.Fatalf("expected 1 broken link, got %d", len(broken))
	}
	if broken[0].Target != "missing.md" {
		t.Errorf("expected target 'missing.md', got '%s'", broken[0].Target)
	}
}

func TestValidateTriggerPaths(t *testing.T) {
	tmpDir := t.TempDir()
	memDir := filepath.Join(tmpDir, "memory")
	wikiDir := filepath.Join(memDir, "wiki")
	os.MkdirAll(wikiDir, 0755)

	os.WriteFile(filepath.Join(wikiDir, "architecture.md"), []byte("# Architecture"), 0644)

	agentsContent := `# Memory Steering
- Check 'wiki/architecture.md' before answering tech stack questions.
- Check 'wiki/nonexistent.md' for history.
- Check 'wiki/*.md' for general topics.
`
	agentsFile := filepath.Join(memDir, "AGENTS.md")
	os.WriteFile(agentsFile, []byte(agentsContent), 0644)

	invalid, err := ValidateTriggerPaths(tmpDir, agentsFile)
	if err != nil {
		t.Fatalf("ValidateTriggerPaths failed: %v", err)
	}

	if len(invalid) != 1 {
		t.Fatalf("expected 1 invalid trigger, got %d", len(invalid))
	}
	if invalid[0].TargetPath != "wiki/nonexistent.md" {
		t.Errorf("expected invalid target 'wiki/nonexistent.md', got '%s'", invalid[0].TargetPath)
	}
}
