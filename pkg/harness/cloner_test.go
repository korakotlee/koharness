package harness_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/spf13/afero"
)

func TestCloneRepository_CollisionHandling(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "repo")

	// Create non-empty target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		t.Fatalf("failed creating target directory: %v", err)
	}
	sampleFile := filepath.Join(targetPath, "existing.txt")
	if err := os.WriteFile(sampleFile, []byte("existing"), 0644); err != nil {
		t.Fatalf("failed writing sample file: %v", err)
	}

	// 1. Without force flag, should return ErrTargetDirectoryExists
	err := harness.CloneRepository("dummy-url", targetPath, false)
	if err == nil || err != harness.ErrTargetDirectoryExists {
		t.Errorf("expected ErrTargetDirectoryExists, got: %v", err)
	}

	// Verify existing file is untouched
	if _, err := os.Stat(sampleFile); os.IsNotExist(err) {
		t.Errorf("expected sampleFile to exist")
	}
}

func TestCloneRepository_EmptyDirectory(t *testing.T) {
	memFs := afero.NewMemMapFs()
	targetPath := "/home/test/repo"
	if err := memFs.MkdirAll(targetPath, 0755); err != nil {
		t.Fatalf("failed creating dir: %v", err)
	}

	opts := harness.ClonerOptions{Fs: memFs}

	// Target is empty directory, should proceed (will fail at git command since dummy-url, but validation passes)
	err := harness.CloneRepository("dummy-url", targetPath, false, opts)
	if err == nil {
		t.Fatalf("expected error from git clone execution, got nil")
	}
	// Verify error is NOT ErrTargetDirectoryExists
	if err == harness.ErrTargetDirectoryExists {
		t.Errorf("unexpected ErrTargetDirectoryExists for empty directory")
	}
}
