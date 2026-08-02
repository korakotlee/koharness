package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harvester"
)

func TestDiscoverCmd_RepoNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentRepo := filepath.Join(tmpDir, "does-not-exist")

	RootCmd.SetArgs([]string{"discover", nonExistentRepo})
	err := RootCmd.Execute()
	if err == nil {
		t.Fatalf("expected error when repo path does not exist")
	}
}

func TestDiscoverCmd_NonInteractiveDiscover(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	repoDir := filepath.Join(tmpDir, "repo")

	_ = os.MkdirAll(homeDir, 0755)
	t.Setenv("HOME", homeDir)

	// Scaffold repo
	creator := harvester.NewCreator(harvester.CreatorOptions{
		HomeDir:  homeDir,
		RepoPath: repoDir,
		InitGit:  false,
	})
	if err := creator.ScaffoldRepo(); err != nil {
		t.Fatalf("failed to scaffold repo: %v", err)
	}

	// Create mock skill in ~/.gemini/config/skills/mock-skill
	skillDir := filepath.Join(homeDir, ".gemini", "config", "skills", "mock-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: mock-skill"), 0644)

	// Execute discover --non-interactive
	RootCmd.SetArgs([]string{"discover", repoDir, "--non-interactive"})
	err := RootCmd.Execute()
	if err != nil {
		t.Fatalf("expected discover command to succeed, got %v", err)
	}

	// Verify imported asset in repo
	importedSkill := filepath.Join(repoDir, "skills", "mock-skill", "SKILL.md")
	if _, err := os.Stat(importedSkill); os.IsNotExist(err) {
		t.Fatalf("expected imported skill file at %s", importedSkill)
	}

	// Re-run discover, should discover 0 new unmanaged capabilities
	RootCmd.SetArgs([]string{"discover", repoDir, "--non-interactive"})
	err = RootCmd.Execute()
	if err != nil {
		t.Fatalf("expected second discover run to succeed, got %v", err)
	}
}
