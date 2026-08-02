package doctor_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/korakotlee/koharness/pkg/doctor"
)

func TestInspectSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink test on Windows platform")
	}

	tmpDir := t.TempDir()

	targetFile := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed writing target file: %v", err)
	}

	validLink := filepath.Join(tmpDir, "valid_link")
	if err := os.Symlink(targetFile, validLink); err != nil {
		t.Fatalf("failed creating valid symlink: %v", err)
	}

	brokenLink := filepath.Join(tmpDir, "broken_link")
	nonExistentTarget := filepath.Join(tmpDir, "non_existent.txt")
	if err := os.Symlink(nonExistentTarget, brokenLink); err != nil {
		t.Fatalf("failed creating broken symlink: %v", err)
	}

	diags, err := doctor.InspectSymlinks([]string{validLink, brokenLink})
	if err != nil {
		t.Fatalf("unexpected error running InspectSymlinks: %v", err)
	}

	if len(diags) != 2 {
		t.Fatalf("expected 2 symlink diagnostics, got %d", len(diags))
	}

	brokenCount := 0
	for _, d := range diags {
		if d.IsBroken {
			brokenCount++
		}
	}

	if brokenCount != 1 {
		t.Errorf("expected 1 broken symlink, got %d", brokenCount)
	}
}

func TestInspectHarnesses(t *testing.T) {
	tmpHome := t.TempDir()

	geminiSkillsDir := filepath.Join(tmpHome, ".gemini", "config", "skills")
	if err := os.MkdirAll(geminiSkillsDir, 0755); err != nil {
		t.Fatalf("failed creating .gemini skills dir: %v", err)
	}

	statuses, err := doctor.InspectHarnesses(tmpHome)
	if err != nil {
		t.Fatalf("unexpected error running InspectHarnesses: %v", err)
	}

	if len(statuses) != 3 {
		t.Errorf("expected 3 harness statuses, got %d", len(statuses))
	}

	foundAntigravity := false
	for _, s := range statuses {
		if s.ID == "antigravity" {
			foundAntigravity = true
			if !s.Installed {
				t.Errorf("expected antigravity to be reported as installed")
			}
			if !s.ConfigDirAccessible {
				t.Errorf("expected antigravity config dir to be accessible")
			}
		}
	}

	if !foundAntigravity {
		t.Errorf("antigravity status not found")
	}
}

func TestDoctorRun(t *testing.T) {
	tmpHome := t.TempDir()

	result, err := doctor.Run(doctor.DoctorOptions{
		HomeDir:           tmpHome,
		SymlinkTargetDirs: []string{tmpHome},
	})
	if err != nil {
		t.Fatalf("unexpected error running doctor.Run: %v", err)
	}

	if result.HasBrokenSymlinks() {
		t.Errorf("expected 0 broken symlinks on clean temp dir, got %d", result.BrokenSymlinkCount())
	}
}
