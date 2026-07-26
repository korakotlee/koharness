package harvester_test

import (
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/harvester"
	"github.com/spf13/afero"
)

func TestCreator_ScaffoldRepoAndHarvest(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/testuser"
	repoPath := filepath.Join(homeDir, ".koharness", "repo")
	backupRoot := filepath.Join(homeDir, ".koharness", "backups")

	// Create dummy original skill
	origSkillDir := filepath.Join(homeDir, ".gemini", "config", "skills", "test-skill")
	_ = fs.MkdirAll(origSkillDir, 0755)
	_ = afero.WriteFile(fs, filepath.Join(origSkillDir, "SKILL.md"), []byte("skill instructions"), 0644)

	items := []harvester.DiscoveredCapability{
		{
			HarnessID:  harness.HarnessAntigravity,
			Type:       harness.CapabilitySkill,
			Name:       "test-skill",
			SourcePath: origSkillDir,
			Selected:   true,
			IsSecret:   false,
		},
	}

	creator := harvester.NewCreator(harvester.CreatorOptions{
		Fs:         fs,
		HomeDir:    homeDir,
		RepoPath:   repoPath,
		BackupRoot: backupRoot,
		InitGit:    false,
	})

	if err := creator.HarvestCapabilities(items); err != nil {
		t.Fatalf("unexpected error during HarvestCapabilities: %v", err)
	}

	// Verify repo scaffolding exists
	checkRepoPaths := []string{
		filepath.Join(repoPath, "skills"),
		filepath.Join(repoPath, "prompts"),
		filepath.Join(repoPath, "mcp"),
		filepath.Join(repoPath, ".gitignore"),
		filepath.Join(repoPath, ".koharness.yaml"),
		filepath.Join(repoPath, "skills", "test-skill", "SKILL.md"),
	}

	for _, p := range checkRepoPaths {
		exists, err := afero.Exists(fs, p)
		if err != nil || !exists {
			t.Errorf("expected path %s to exist in repository", p)
		}
	}

	// Verify backup was created
	backupEntries, err := afero.ReadDir(fs, backupRoot)
	if err != nil || len(backupEntries) == 0 {
		t.Errorf("expected backup directory session to exist under %s", backupRoot)
	}
}

func TestCreator_IsRepoExisting(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/testuser"
	repoPath := filepath.Join(homeDir, ".koharness", "repo")

	creator := harvester.NewCreator(harvester.CreatorOptions{
		Fs:       fs,
		HomeDir:  homeDir,
		RepoPath: repoPath,
	})

	exists, err := creator.IsRepoExisting()
	if err != nil {
		t.Fatalf("unexpected error checking IsRepoExisting: %v", err)
	}
	if exists {
		t.Errorf("expected IsRepoExisting to return false for non-existent repo")
	}

	_ = creator.ScaffoldRepo()
	exists, err = creator.IsRepoExisting()
	if err != nil {
		t.Fatalf("unexpected error checking IsRepoExisting after scaffold: %v", err)
	}
	if !exists {
		t.Errorf("expected IsRepoExisting to return true after scaffold")
	}
}
