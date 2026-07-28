package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/korakotlee/koharness/cmd"
	"github.com/korakotlee/koharness/pkg/harness"
)

func TestRepoCmd_PrintPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "koharness-repo-test-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)

	cmd.RepoCmd.Flags().Set("path", "")
	cmd.RepoCmd.Flags().Set("print-path", "false")
	cmd.RepoCmd.Flags().Set("shell-init", "false")

	cmd.RootCmd.SetArgs([]string{"repo", "-p", "-d", tempDir})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	expected := filepath.Clean(tempDir)
	if !strings.Contains(output, expected) {
		t.Errorf("got output %q, expected to contain %q", output, expected)
	}
}

func TestRepoCmd_NonExistentRepo(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)

	cmd.RepoCmd.Flags().Set("path", "")
	cmd.RepoCmd.Flags().Set("print-path", "false")
	cmd.RepoCmd.Flags().Set("shell-init", "false")

	nonExistentPath := "/path/that/does/not/exist/9999"
	cmd.RootCmd.SetArgs([]string{"repo", "-d", nonExistentPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent repo path, got nil")
	}

	if !strings.Contains(err.Error(), "repository directory not found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRepoCmd_ShellInit(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)

	cmd.RepoCmd.Flags().Set("path", "")
	cmd.RepoCmd.Flags().Set("print-path", "false")
	cmd.RepoCmd.Flags().Set("shell-init", "false")

	cmd.RootCmd.SetArgs([]string{"repo", "--shell-init"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error running --shell-init: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "khcd() {") {
		t.Errorf("expected shell init snippet output, got: %s", output)
	}
}

func TestRepoCmd_EnvVarOverride(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "koharness-env-repo-test-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv(harness.EnvKoharnessRepo, tempDir)
	defer os.Unsetenv(harness.EnvKoharnessRepo)

	buf := new(bytes.Buffer)
	cmd.RootCmd.SetOut(buf)
	cmd.RootCmd.SetErr(buf)

	cmd.RepoCmd.Flags().Set("path", "")
	cmd.RepoCmd.Flags().Set("print-path", "false")
	cmd.RepoCmd.Flags().Set("shell-init", "false")

	cmd.RootCmd.SetArgs([]string{"repo", "-p"})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	expected := filepath.Clean(tempDir)
	if !strings.Contains(output, expected) {
		t.Errorf("got output %q, expected to contain %q", output, expected)
	}
}
