package harvester

import (
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/spf13/afero"
)

func TestLocalConfig_LoadAndSave(t *testing.T) {
	fs := afero.NewMemMapFs()
	repoPath := "/repo"
	_ = fs.MkdirAll(repoPath, 0755)

	cfg, err := LoadLocalConfig(fs, repoPath)
	if err != nil {
		t.Fatalf("expected no error loading empty config, got %v", err)
	}

	if len(cfg.IgnoredCapabilities) != 0 {
		t.Fatalf("expected 0 items, got %d", len(cfg.IgnoredCapabilities))
	}

	cfg.AddIgnore(harness.CapabilitySkill, "test-skill", harness.HarnessAntigravity)
	if !cfg.IsIgnored(harness.CapabilitySkill, "test-skill", harness.HarnessAntigravity) {
		t.Fatalf("expected test-skill to be marked as ignored")
	}

	err = SaveLocalConfig(fs, repoPath, cfg)
	if err != nil {
		t.Fatalf("failed to save local config: %v", err)
	}

	// Verify .gitignore updated
	gitignoreData, err := afero.ReadFile(fs, filepath.Join(repoPath, ".gitignore"))
	if err != nil {
		t.Fatalf("expected .gitignore to exist, got %v", err)
	}
	if string(gitignoreData) != ".koharness.local.yaml\n" {
		t.Fatalf("unexpected .gitignore content: %s", string(gitignoreData))
	}

	// Load back
	loaded, err := LoadLocalConfig(fs, repoPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if !loaded.IsIgnored(harness.CapabilitySkill, "test-skill", harness.HarnessAntigravity) {
		t.Fatalf("expected test-skill to be ignored after reload")
	}
}

func TestEnsureGitignore_Idempotent(t *testing.T) {
	fs := afero.NewMemMapFs()
	repoPath := "/repo"
	_ = fs.MkdirAll(repoPath, 0755)

	err := EnsureGitignore(fs, repoPath, ".koharness.local.yaml")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	err = EnsureGitignore(fs, repoPath, ".koharness.local.yaml")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	data, _ := afero.ReadFile(fs, filepath.Join(repoPath, ".gitignore"))
	if string(data) != ".koharness.local.yaml\n" {
		t.Fatalf("expected single entry in .gitignore, got: %s", string(data))
	}
}
