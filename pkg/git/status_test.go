package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParsePorcelain(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantDirty  bool
		wantBranch string
		wantFiles  int
	}{
		{
			name:       "clean branch",
			input:      "## main...origin/main\n",
			wantDirty:  false,
			wantBranch: "main",
			wantFiles:  0,
		},
		{
			name: "dirty branch with modifications and untracked files",
			input: `## feature/sync...origin/feature/sync
 M pkg/git/status.go
?? newfile.txt
D  deleted.txt
`,
			wantDirty:  true,
			wantBranch: "feature/sync",
			wantFiles:  3,
		},
		{
			name:       "detached HEAD",
			input:      "## HEAD (no branch)\n M README.md\n",
			wantDirty:  true,
			wantBranch: "HEAD",
			wantFiles:  1,
		},
		{
			name:       "renamed file",
			input:      "## main\nR  old.txt -> new.txt\n",
			wantDirty:  true,
			wantBranch: "main",
			wantFiles:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePorcelain(tt.input)
			if got.IsDirty != tt.wantDirty {
				t.Errorf("IsDirty = %v, want %v", got.IsDirty, tt.wantDirty)
			}
			if got.Branch != tt.wantBranch {
				t.Errorf("Branch = %q, want %q", got.Branch, tt.wantBranch)
			}
			if len(got.DirtyFiles) != tt.wantFiles {
				t.Errorf("len(DirtyFiles) = %d, want %d", len(got.DirtyFiles), tt.wantFiles)
			}
		})
	}
}

func TestInspectStatusAndPullRebaseIntegration(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found in PATH")
	}

	tmpDir := t.TempDir()

	// Initialize git repo in tmpDir
	cmdInit := exec.Command(gitBin, "init", tmpDir)
	if output, err := cmdInit.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v, output: %s", err, output)
	}

	// Configure user for test repo
	_ = exec.Command(gitBin, "-C", tmpDir, "config", "user.name", "Test User").Run()
	_ = exec.Command(gitBin, "-C", tmpDir, "config", "user.email", "test@example.com").Run()

	ctx := context.Background()

	// Test 1: Empty repo should show status (untracked file after creation)
	file1 := filepath.Join(tmpDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed writing test file: %v", err)
	}

	status, err := InspectStatus(ctx, tmpDir)
	if err != nil {
		t.Fatalf("InspectStatus failed: %v", err)
	}
	if !status.IsDirty {
		t.Errorf("expected repo to be dirty with untracked file")
	}
	if len(status.DirtyFiles) != 1 || status.DirtyFiles[0].Path != "file1.txt" {
		t.Errorf("unexpected dirty files: %+v", status.DirtyFiles)
	}

	// Commit file to make clean
	_ = exec.Command(gitBin, "-C", tmpDir, "add", ".").Run()
	_ = exec.Command(gitBin, "-C", tmpDir, "commit", "-m", "initial commit").Run()

	statusClean, err := InspectStatus(ctx, tmpDir)
	if err != nil {
		t.Fatalf("InspectStatus failed: %v", err)
	}
	if statusClean.IsDirty {
		t.Errorf("expected repo to be clean after commit")
	}
}
