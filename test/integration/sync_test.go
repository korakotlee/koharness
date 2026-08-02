package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/sync"
	"github.com/spf13/afero"
)

func setupTestGitRepo(t *testing.T) (remotePath string, localPath string) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found in PATH")
	}

	tmpDir := t.TempDir()
	remotePath = filepath.Join(tmpDir, "remote.git")
	localPath = filepath.Join(tmpDir, "local-repo")

	// Init bare remote repo
	cmd := exec.Command(gitBin, "init", "--bare", remotePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed init bare remote repo: %v (out: %s)", err, out)
	}

	// Init temporary clone repo to make initial commit
	seedPath := filepath.Join(tmpDir, "seed")
	_ = exec.Command(gitBin, "init", seedPath).Run()
	_ = exec.Command(gitBin, "-C", seedPath, "config", "user.name", "Test").Run()
	_ = exec.Command(gitBin, "-C", seedPath, "config", "user.email", "test@test.com").Run()

	// Initial commit with a skill
	skillDir := filepath.Join(seedPath, "skills", "test-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test Skill"), 0644)

	_ = exec.Command(gitBin, "-C", seedPath, "add", ".").Run()
	_ = exec.Command(gitBin, "-C", seedPath, "commit", "-m", "initial commit").Run()
	_ = exec.Command(gitBin, "-C", seedPath, "remote", "add", "origin", remotePath).Run()
	_ = exec.Command(gitBin, "-C", seedPath, "branch", "-M", "main").Run()
	_ = exec.Command(gitBin, "-C", seedPath, "push", "-u", "origin", "main").Run()

	// Clone to localPath
	_ = exec.Command(gitBin, "clone", remotePath, localPath).Run()
	_ = exec.Command(gitBin, "-C", localPath, "config", "user.name", "Test").Run()
	_ = exec.Command(gitBin, "-C", localPath, "config", "user.email", "test@test.com").Run()

	return remotePath, localPath
}

func TestIntegration_SyncCleanRepository(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found")
	}

	remotePath, localPath := setupTestGitRepo(t)
	tmpHome := t.TempDir()

	// Create a new commit on remote
	extraSeed := t.TempDir()
	_ = exec.Command(gitBin, "clone", remotePath, extraSeed).Run()
	_ = exec.Command(gitBin, "-C", extraSeed, "config", "user.name", "Test").Run()
	_ = exec.Command(gitBin, "-C", extraSeed, "config", "user.email", "test@test.com").Run()

	newSkillDir := filepath.Join(extraSeed, "skills", "new-remote-skill")
	_ = os.MkdirAll(newSkillDir, 0755)
	_ = os.WriteFile(filepath.Join(newSkillDir, "SKILL.md"), []byte("# New Remote Skill"), 0644)

	_ = exec.Command(gitBin, "-C", extraSeed, "add", ".").Run()
	_ = exec.Command(gitBin, "-C", extraSeed, "commit", "-m", "add remote skill").Run()
	_ = exec.Command(gitBin, "-C", extraSeed, "push", "origin", "main").Run()

	// Run sync on localPath
	engine, err := sync.NewSyncEngine(sync.SyncOptions{
		RepoPath:       localPath,
		HomeDir:        tmpHome,
		NonInteractive: true,
		Fs:             afero.NewOsFs(),
	})
	if err != nil {
		t.Fatalf("failed creating SyncEngine: %v", err)
	}

	ctx := context.Background()
	res, err := engine.Run(ctx)
	if err != nil {
		t.Fatalf("sync engine failed: %v", err)
	}

	if res.RepoPath != localPath {
		t.Errorf("RepoPath = %s, want %s", res.RepoPath, localPath)
	}

	// Verify newly pulled skill exists in localPath
	pulledSkill := filepath.Join(localPath, "skills", "new-remote-skill", "SKILL.md")
	if _, err := os.Stat(pulledSkill); os.IsNotExist(err) {
		t.Errorf("expected pulled skill file to exist at %s", pulledSkill)
	}
}

func TestIntegration_SyncDirtyRepositoryGuard(t *testing.T) {
	_, localPath := setupTestGitRepo(t)
	tmpHome := t.TempDir()

	// Create uncommitted change in localPath
	dirtyFile := filepath.Join(localPath, "uncommitted.txt")
	_ = os.WriteFile(dirtyFile, []byte("local work"), 0644)

	engine, err := sync.NewSyncEngine(sync.SyncOptions{
		RepoPath:       localPath,
		HomeDir:        tmpHome,
		NonInteractive: true,
		Fs:             afero.NewOsFs(),
	})
	if err != nil {
		t.Fatalf("failed creating SyncEngine: %v", err)
	}

	ctx := context.Background()
	_, err = engine.Run(ctx)
	if err != sync.ErrDirtyRepo {
		t.Errorf("expected ErrDirtyRepo when syncing dirty repository, got %v", err)
	}
}
