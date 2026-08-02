package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/doctor"
	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/harvester"
	"github.com/korakotlee/koharness/pkg/lint"
	"github.com/spf13/afero"
)

func TestIntegration_MemoryArchitectureFlow(t *testing.T) {
	tmpHome := t.TempDir()
	repoPath := filepath.Join(tmpHome, ".koharness", "repo")

	// 1. Scaffold dotfiles repo with memory architecture using harvester.Creator
	creator := harvester.NewCreator(harvester.CreatorOptions{
		HomeDir:  tmpHome,
		RepoPath: repoPath,
		InitGit:  false,
	})

	if err := creator.ScaffoldRepo(); err != nil {
		t.Fatalf("ScaffoldRepo failed: %v", err)
	}

	// Verify memory components created
	memoryAgents := filepath.Join(repoPath, "memory", "AGENTS.md")
	if _, err := os.Stat(memoryAgents); os.IsNotExist(err) {
		t.Fatalf("expected memory/AGENTS.md to be scaffolded")
	}

	// 2. Perform sync and verify client harness memory block injection
	if err := harness.SyncMemoryNavigationRules(afero.NewOsFs(), tmpHome, repoPath); err != nil {
		t.Fatalf("SyncMemoryNavigationRules failed: %v", err)
	}

	geminiAgents := filepath.Join(tmpHome, ".gemini", "AGENTS.md")
	if data, err := os.ReadFile(geminiAgents); err != nil || len(data) == 0 {
		t.Errorf("expected client harness memory navigation block at %s", geminiAgents)
	}

	// 3. Run static linter on memory layout
	lintRes, err := lint.Run(lint.LintOptions{RepoRoot: repoPath})
	if err != nil {
		t.Fatalf("lint.Run failed: %v", err)
	}
	if lintRes.HasErrors() {
		t.Errorf("expected clean lint on freshly scaffolded memory repo, got: %v", lintRes.Issues)
	}

	// 4. Run doctor diagnostics
	docRes, err := doctor.Run(doctor.DoctorOptions{
		HomeDir:  tmpHome,
		RepoRoot: repoPath,
	})
	if err != nil {
		t.Fatalf("doctor.Run failed: %v", err)
	}
	if docRes.Memory == nil || !docRes.Memory.IsHealthy() {
		t.Errorf("expected healthy doctor memory diagnostic result")
	}
}
