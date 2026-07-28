package cmd_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/korakotlee/koharness/cmd"
)

func TestInitCmd_LocalRepoCloning(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	// Create sample source repository directory
	srcRepo := filepath.Join(tempDir, "src-repo")
	if err := os.MkdirAll(filepath.Join(srcRepo, "skills", "my-skill"), 0755); err != nil {
		t.Fatalf("failed creating src repo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcRepo, "skills", "my-skill", "SKILL.md"), []byte("sample skill"), 0644); err != nil {
		t.Fatalf("failed writing skill file: %v", err)
	}

	exec.Command("git", "-C", srcRepo, "init").Run()
	exec.Command("git", "-C", srcRepo, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", srcRepo, "config", "user.name", "test").Run()
	exec.Command("git", "-C", srcRepo, "add", ".").Run()
	exec.Command("git", "-C", srcRepo, "commit", "-m", "init").Run()

	targetRepo := filepath.Join(tempDir, "target-repo")

	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)

	cmd.InitCmd.Flags().Set("path", "")
	cmd.InitCmd.Flags().Set("force", "false")
	cmd.InitCmd.Flags().Set("non-interactive", "true")

	cmd.RootCmd.SetArgs([]string{"init", srcRepo, targetRepo, "--non-interactive"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Repository initialized") {
		t.Errorf("expected success message in output, got: %s", output)
	}
	if !strings.Contains(output, "koharness discover") {
		t.Errorf("expected guidance referencing koharness discover, got: %s", output)
	}

	// Verify target repo contains cloned skill
	clonedSkill := filepath.Join(targetRepo, "skills", "my-skill", "SKILL.md")
	if _, err := os.Stat(clonedSkill); os.IsNotExist(err) {
		t.Errorf("expected cloned skill at %s", clonedSkill)
	}
}

func TestInitCmd_CollisionError(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	targetRepo := filepath.Join(tempDir, "existing-repo")
	if err := os.MkdirAll(targetRepo, 0755); err != nil {
		t.Fatalf("failed creating target repo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetRepo, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed writing sample file: %v", err)
	}

	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)

	cmd.InitCmd.Flags().Set("path", "")
	cmd.InitCmd.Flags().Set("force", "false")
	cmd.InitCmd.Flags().Set("non-interactive", "true")

	cmd.RootCmd.SetArgs([]string{"init", "some-url", targetRepo})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when initializing over non-empty directory without force flag, got nil")
	}

	output := buf.String()
	if !strings.Contains(output, "REPOSITORY ALREADY EXISTS") {
		t.Errorf("expected collision badge in stderr/stdout, got: %s", output)
	}
}
