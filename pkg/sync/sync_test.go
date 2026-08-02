package sync

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestSyncEngine_DirtyRepoGuard(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git binary not found in PATH")
	}

	tmpHome := t.TempDir()
	repoPath := filepath.Join(tmpHome, ".koharness", "repo")

	// Init git repo
	_ = exec.Command(gitBin, "init", repoPath).Run()
	_ = exec.Command(gitBin, "-C", repoPath, "config", "user.name", "Test").Run()
	_ = exec.Command(gitBin, "-C", repoPath, "config", "user.email", "test@test.com").Run()

	fs := afero.NewOsFs()
	_ = afero.WriteFile(fs, filepath.Join(repoPath, "dirty.txt"), []byte("uncommitted"), 0644)

	var errBuf bytes.Buffer
	engine, err := NewSyncEngine(SyncOptions{
		RepoPath:       repoPath,
		HomeDir:        tmpHome,
		NonInteractive: true,
		Fs:             fs,
		ErrOut:         &errBuf,
	})
	if err != nil {
		t.Fatalf("failed creating SyncEngine: %v", err)
	}

	ctx := context.Background()
	res, err := engine.Run(ctx)
	if err != ErrDirtyRepo {
		t.Errorf("expected ErrDirtyRepo, got %v (res=%v)", err, res)
	}
	if !bytes.Contains(errBuf.Bytes(), []byte("UNCOMMITTED LOCAL CHANGES DETECTED")) {
		t.Errorf("expected warning header in stderr, got %s", errBuf.String())
	}
}
