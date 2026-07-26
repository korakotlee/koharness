package symlink_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/korakotlee/koharness/pkg/symlink"
)

func TestEndToEnd_HarnessSymlinkAndBackupSafety(t *testing.T) {
	tempHome := t.TempDir()
	fs := afero.NewOsFs()

	bm := symlink.NewBackupManager(fs, tempHome, "")
	engine := symlink.NewLinkerEngine(symlink.LinkerConfig{
		Fs:            fs,
		HomeDir:       tempHome,
		BackupManager: bm,
	})

	// Setup mock harness paths
	antigravitySkillDir := filepath.Join(tempHome, ".gemini", "config", "skills", "custom-skill")
	existingSkillFile := filepath.Join(antigravitySkillDir, "SKILL.md")
	oldContent := []byte("name: Old Skill Spec")

	if err := fs.MkdirAll(antigravitySkillDir, 0755); err != nil {
		t.Fatalf("failed setup antigravity dir: %v", err)
	}
	if err := afero.WriteFile(fs, existingSkillFile, oldContent, 0644); err != nil {
		t.Fatalf("failed setup existing skill file: %v", err)
	}

	// Setup new capability source
	repoSourceDir := filepath.Join(tempHome, "repo", "capabilities", "custom-skill")
	newSkillFile := filepath.Join(repoSourceDir, "SKILL.md")
	newContent := []byte("name: Updated Skill Spec v2")

	if err := fs.MkdirAll(repoSourceDir, 0755); err != nil {
		t.Fatalf("failed setup repo source dir: %v", err)
	}
	if err := afero.WriteFile(fs, newSkillFile, newContent, 0644); err != nil {
		t.Fatalf("failed setup new skill file: %v", err)
	}

	// Create atomic symlink over pre-existing file
	tx, err := engine.CreateSymlink(newSkillFile, existingSkillFile)
	if err != nil {
		t.Fatalf("CreateSymlink failed during E2E integration test: %v", err)
	}

	// 1. Verify backup created under ~/.koharness/backups/
	sessionDir := bm.SessionDir()
	if sessionDir == "" {
		t.Fatalf("expected non-empty session dir after backup")
	}

	backupFile := filepath.Join(sessionDir, ".gemini", "config", "skills", "custom-skill", "SKILL.md")
	bContent, err := afero.ReadFile(fs, backupFile)
	if err != nil || string(bContent) != string(oldContent) {
		t.Errorf("backup file mismatch: got %q, want %q (err: %v)", string(bContent), string(oldContent), err)
	}

	// 2. Verify target link resolves to new capability content
	readContent, err := afero.ReadFile(fs, existingSkillFile)
	if err != nil || string(readContent) != string(newContent) {
		t.Errorf("target path content mismatch: got %q, want %q (err: %v)", string(readContent), string(newContent), err)
	}

	// 3. Test Rollback
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	restoredContent, err := afero.ReadFile(fs, existingSkillFile)
	if err != nil || string(restoredContent) != string(oldContent) {
		t.Errorf("target path content after rollback mismatch: got %q, want %q (err: %v)", string(restoredContent), string(oldContent), err)
	}
}
