package symlink

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestBackupManager_EnsureSession(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/mock/user"
	bm := NewBackupManager(fs, homeDir, "")

	sessionDir, err := bm.EnsureSession()
	if err != nil {
		t.Fatalf("EnsureSession failed: %v", err)
	}

	if !strings.HasPrefix(sessionDir, filepath.Join(homeDir, ".koharness", "backups")) {
		t.Errorf("unexpected session directory path: %s", sessionDir)
	}

	exists, err := afero.DirExists(fs, sessionDir)
	if err != nil || !exists {
		t.Errorf("session directory was not created on filesystem: %s", sessionDir)
	}

	// Secondary call should return the same session directory
	sessionDir2, err := bm.EnsureSession()
	if err != nil || sessionDir2 != sessionDir {
		t.Errorf("EnsureSession did not return idempotent session dir: got %s, want %s", sessionDir2, sessionDir)
	}
}

func TestBackupManager_BackupSingleFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/mock/user"
	bm := NewBackupManager(fs, homeDir, "")

	originalPath := filepath.Join(homeDir, ".gemini", "config", "skills", "weather", "SKILL.md")
	content := []byte("name: Weather Skill\ndescription: Forecast skill")
	if err := afero.WriteFile(fs, originalPath, content, 0644); err != nil {
		t.Fatalf("failed to setup test file: %v", err)
	}

	rec, err := bm.Backup(originalPath)
	if err != nil {
		t.Fatalf("Backup failed for file %s: %v", originalPath, err)
	}

	if rec.OriginalPath != originalPath {
		t.Errorf("expected original path %s, got %s", originalPath, rec.OriginalPath)
	}

	backupContent, err := afero.ReadFile(fs, rec.BackupPath)
	if err != nil {
		t.Fatalf("failed reading backup file content: %v", err)
	}

	if string(backupContent) != string(content) {
		t.Errorf("backup content mismatch: got %q, want %q", string(backupContent), string(content))
	}
}

func TestBackupManager_BackupDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/mock/user"
	bm := NewBackupManager(fs, homeDir, "")

	dirPath := filepath.Join(homeDir, ".claude", "commands")
	file1 := filepath.Join(dirPath, "build.sh")
	file2 := filepath.Join(dirPath, "sub", "deploy.sh")

	_ = afero.WriteFile(fs, file1, []byte("echo build"), 0755)
	_ = afero.WriteFile(fs, file2, []byte("echo deploy"), 0755)

	rec, err := bm.Backup(dirPath)
	if err != nil {
		t.Fatalf("Backup failed for directory %s: %v", dirPath, err)
	}

	if !rec.IsDirectory {
		t.Errorf("expected IsDirectory true, got false")
	}

	bFile1, err := afero.ReadFile(fs, filepath.Join(rec.BackupPath, "build.sh"))
	if err != nil || string(bFile1) != "echo build" {
		t.Errorf("backup dir missing or invalid child build.sh: %v", err)
	}

	bFile2, err := afero.ReadFile(fs, filepath.Join(rec.BackupPath, "sub", "deploy.sh"))
	if err != nil || string(bFile2) != "echo deploy" {
		t.Errorf("backup dir missing or invalid nested deploy.sh: %v", err)
	}
}

func TestBackupManager_NonExistentTarget(t *testing.T) {
	fs := afero.NewMemMapFs()
	bm := NewBackupManager(fs, "/mock/user", "")

	_, err := bm.Backup("/mock/user/missing.txt")
	if err == nil {
		t.Errorf("expected error when backing up non-existent target, got nil")
	}
}
