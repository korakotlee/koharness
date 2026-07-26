package symlink

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestLinkerEngine_CreateSymlink_NewTarget(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/mock/user"
	engine := NewLinkerEngine(LinkerConfig{
		Fs:      fs,
		HomeDir: homeDir,
	})

	sourcePath := filepath.Join(homeDir, "source-skill", "SKILL.md")
	targetPath := filepath.Join(homeDir, ".gemini", "config", "skills", "source-skill", "SKILL.md")

	if err := afero.WriteFile(fs, sourcePath, []byte("skill payload"), 0644); err != nil {
		t.Fatalf("failed to setup source path: %v", err)
	}

	tx, err := engine.CreateSymlink(sourcePath, targetPath)
	if err != nil {
		t.Fatalf("CreateSymlink failed: %v", err)
	}

	if len(tx.Operations) == 0 || tx.Operations[0].Status != OpStatusCommitted {
		t.Errorf("expected operation status committed, got %v", tx.Operations)
	}

	exists, err := afero.Exists(fs, targetPath)
	if err != nil || !exists {
		t.Errorf("target symlink path does not exist: %s", targetPath)
	}
}

func TestLinkerEngine_CreateSymlink_OverwritesExistingTargetWithBackup(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/mock/user"
	bm := NewBackupManager(fs, homeDir, "")
	engine := NewLinkerEngine(LinkerConfig{
		Fs:            fs,
		HomeDir:       homeDir,
		BackupManager: bm,
	})

	sourcePath := filepath.Join(homeDir, "new-version", "config.json")
	targetPath := filepath.Join(homeDir, ".claude.json")

	oldContent := []byte(`{"version": 1}`)
	newContent := []byte(`{"version": 2}`)

	_ = afero.WriteFile(fs, targetPath, oldContent, 0644)
	_ = afero.WriteFile(fs, sourcePath, newContent, 0644)

	tx, err := engine.CreateSymlink(sourcePath, targetPath)
	if err != nil {
		t.Fatalf("CreateSymlink over existing file failed: %v", err)
	}

	op := tx.Operations[0]
	if op.BackupRecord == nil {
		t.Fatalf("expected BackupRecord to be populated for existing file overwrite")
	}

	backupContent, err := afero.ReadFile(fs, op.BackupRecord.BackupPath)
	if err != nil || string(backupContent) != string(oldContent) {
		t.Errorf("backup content mismatch: got %q, want %q", string(backupContent), string(oldContent))
	}
}

func TestLinkerEngine_DryRun(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/mock/user"
	engine := NewLinkerEngine(LinkerConfig{
		Fs:      fs,
		HomeDir: homeDir,
		DryRun:  true,
	})

	sourcePath := filepath.Join(homeDir, "source.txt")
	targetPath := filepath.Join(homeDir, "target.txt")

	_ = afero.WriteFile(fs, sourcePath, []byte("dry run source"), 0644)

	tx, err := engine.CreateSymlink(sourcePath, targetPath)
	if err != nil {
		t.Fatalf("CreateSymlink in dry-run mode failed: %v", err)
	}

	if !tx.DryRun {
		t.Errorf("expected transaction DryRun to be true")
	}

	// In dry-run mode, target file must NOT be written to filesystem
	exists, _ := afero.Exists(fs, targetPath)
	if exists {
		t.Errorf("target path should not exist on disk in dry-run mode")
	}
}

func TestLinkerEngine_BrokenSymlinkRepair(t *testing.T) {
	tmpDir := t.TempDir()
	fs := afero.NewOsFs()
	engine := NewLinkerEngine(LinkerConfig{
		Fs:      fs,
		HomeDir: tmpDir,
	})

	oldSource := filepath.Join(tmpDir, "old-source.txt")
	newSource := filepath.Join(tmpDir, "new-source.txt")
	targetPath := filepath.Join(tmpDir, "target.txt")

	// Create valid initial symlink then remove source to break it
	_ = afero.WriteFile(fs, oldSource, []byte("old source"), 0644)
	_ = afero.WriteFile(fs, newSource, []byte("new source"), 0644)
	_, _ = engine.CreateSymlink(oldSource, targetPath)

	// Remove old source to break link
	_ = fs.Remove(oldSource)

	isBroken, err := engine.IsBrokenSymlink(targetPath)
	if err != nil || !isBroken {
		t.Fatalf("expected IsBrokenSymlink to return true for dangling symlink, got %v (err: %v)", isBroken, err)
	}

	// Repair broken symlink
	_, err = engine.RepairSymlink(newSource, targetPath)
	if err != nil {
		t.Fatalf("RepairSymlink failed: %v", err)
	}

	isBroken, _ = engine.IsBrokenSymlink(targetPath)
	if isBroken {
		t.Errorf("expected symlink to be repaired, but IsBrokenSymlink still returned true")
	}
}

func TestTransaction_Rollback(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/mock/user"
	bm := NewBackupManager(fs, homeDir, "")

	targetPath := filepath.Join(homeDir, "config.yaml")
	originalContent := []byte("original: true")
	_ = afero.WriteFile(fs, targetPath, originalContent, 0644)

	rec, err := bm.Backup(targetPath)
	if err != nil {
		t.Fatalf("setup backup failed: %v", err)
	}

	// Simulate modified target
	_ = afero.WriteFile(fs, targetPath, []byte("corrupted: true"), 0644)

	tx := NewTransaction(fs, false)
	op := &OperationLog{
		TargetPath:   targetPath,
		BackupRecord: rec,
		Status:       OpStatusExecuted,
	}
	tx.Operations = append(tx.Operations, op)

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	restoredContent, err := afero.ReadFile(fs, targetPath)
	if err != nil || string(restoredContent) != string(originalContent) {
		t.Errorf("rollback failed to restore original content: got %q, want %q", string(restoredContent), string(originalContent))
	}
}
