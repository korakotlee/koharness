package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/korakotlee/koharness/cmd"
)

func TestUninstallCmd_Execution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	repoDir := filepath.Join(tempDir, ".koharness", "repo")
	geminiSkills := filepath.Join(tempDir, ".gemini", "config", "skills")

	repoSkillFile := filepath.Join(repoDir, "skills", "test-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(repoSkillFile), 0755); err != nil {
		t.Fatalf("failed setup repo dir: %v", err)
	}
	if err := os.WriteFile(repoSkillFile, []byte("skill data"), 0644); err != nil {
		t.Fatalf("failed writing repo skill file: %v", err)
	}

	if err := os.MkdirAll(geminiSkills, 0755); err != nil {
		t.Fatalf("failed setup gemini dir: %v", err)
	}

	targetSkillDir := filepath.Join(geminiSkills, "test-skill")
	repoSkillDir := filepath.Join(repoDir, "skills", "test-skill")
	if err := os.Symlink(repoSkillDir, targetSkillDir); err != nil {
		t.Skipf("symlink creation failed: %v", err)
	}

	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)

	cmd.RootCmd.SetArgs([]string{"uninstall", "--force", "--purge-config"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("uninstall command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Uninstallation execution completed") {
		t.Errorf("expected completion message, got: %s", output)
	}

	// Verify targetSkillDir is now a real standalone directory containing SKILL.md
	st, err := os.Lstat(targetSkillDir)
	if err != nil {
		t.Fatalf("failed to stat target skill dir: %v", err)
	}
	if st.Mode()&os.ModeSymlink != 0 {
		t.Errorf("expected target skill dir to be a standalone directory, not a symlink")
	}

	content, err := os.ReadFile(filepath.Join(targetSkillDir, "SKILL.md"))
	if err != nil || string(content) != "skill data" {
		t.Errorf("restored file content mismatch: got %q", string(content))
	}

	// Verify repoDir is removed
	if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
		t.Errorf("expected repository directory to be deleted")
	}
}

func TestUninstallCmd_DryRun(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	repoDir := filepath.Join(tempDir, ".koharness", "repo")
	geminiSkills := filepath.Join(tempDir, ".gemini", "config", "skills")

	repoSkillFile := filepath.Join(repoDir, "skills", "dry-skill", "SKILL.md")
	_ = os.MkdirAll(filepath.Dir(repoSkillFile), 0755)
	_ = os.WriteFile(repoSkillFile, []byte("dry data"), 0644)

	_ = os.MkdirAll(geminiSkills, 0755)
	targetSkillDir := filepath.Join(geminiSkills, "dry-skill")
	repoSkillDir := filepath.Join(repoDir, "skills", "dry-skill")
	if err := os.Symlink(repoSkillDir, targetSkillDir); err != nil {
		t.Skipf("symlink creation failed: %v", err)
	}

	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)

	cmd.RootCmd.SetArgs([]string{"uninstall", "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("uninstall --dry-run failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN header in output, got: %s", output)
	}

	// Target should remain a symlink during dry-run
	st, err := os.Lstat(targetSkillDir)
	if err != nil || st.Mode()&os.ModeSymlink == 0 {
		t.Errorf("target should remain a symlink during dry-run")
	}
}
