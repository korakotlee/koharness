package cmd

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSyncCmd_DirtyRepo(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found in PATH")
	}

	tmpDir := t.TempDir()

	// Init git repo
	_ = exec.Command(gitBin, "init", tmpDir).Run()
	_ = exec.Command(gitBin, "-C", tmpDir, "config", "user.name", "Test").Run()
	_ = exec.Command(gitBin, "-C", tmpDir, "config", "user.email", "test@test.com").Run()

	// Create dirty file
	dirtyFile := filepath.Join(tmpDir, "uncommitted.txt")
	_ = exec.Command(gitBin, "-C", tmpDir, "add", ".").Run()
	_ = exec.Command("bash", "-c", "echo 'dirty' > "+dirtyFile).Run()

	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	RootCmd.SetOut(buf)
	RootCmd.SetErr(errBuf)
	RootCmd.SetArgs([]string{"sync", tmpDir, "--non-interactive"})

	err = RootCmd.Execute()
	if err == nil {
		t.Fatalf("expected sync command to fail on dirty repository")
	}
}
