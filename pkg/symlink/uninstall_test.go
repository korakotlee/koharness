package symlink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/spf13/afero"
)

func TestUninstallEngine_DiscoverAndExecute(t *testing.T) {
	tmpDir := t.TempDir()
	fs := afero.NewOsFs()

	homeDir := filepath.Join(tmpDir, "user")
	repoDir := filepath.Join(homeDir, ".koharness", "repo")
	geminiSkills := filepath.Join(homeDir, ".gemini", "config", "skills")
	claudeWorkflows := filepath.Join(homeDir, ".claude", "workflows")

	// Setup repository files
	repoSkillFile := filepath.Join(repoDir, "skills", "ipd", "SKILL.md")
	repoWorkflowFile := filepath.Join(repoDir, "workflows", "review.md")

	if err := fs.MkdirAll(filepath.Dir(repoSkillFile), 0755); err != nil {
		t.Fatalf("failed setup repo skill dir: %v", err)
	}
	if err := fs.MkdirAll(filepath.Dir(repoWorkflowFile), 0755); err != nil {
		t.Fatalf("failed setup repo workflow dir: %v", err)
	}

	if err := afero.WriteFile(fs, repoSkillFile, []byte("skill payload"), 0644); err != nil {
		t.Fatalf("failed writing repo skill file: %v", err)
	}
	if err := afero.WriteFile(fs, repoWorkflowFile, []byte("workflow payload"), 0644); err != nil {
		t.Fatalf("failed writing repo workflow file: %v", err)
	}

	// Setup target directories and symlinks
	if err := fs.MkdirAll(geminiSkills, 0755); err != nil {
		t.Fatalf("failed setup gemini skills dir: %v", err)
	}
	if err := fs.MkdirAll(claudeWorkflows, 0755); err != nil {
		t.Fatalf("failed setup claude workflows dir: %v", err)
	}

	targetSkillDir := filepath.Join(geminiSkills, "ipd")
	targetWorkflowFile := filepath.Join(claudeWorkflows, "review.md")

	repoSkillDir := filepath.Join(repoDir, "skills", "ipd")
	if linker, ok := fs.(afero.Symlinker); ok {
		if err := linker.SymlinkIfPossible(repoSkillDir, targetSkillDir); err != nil {
			t.Fatalf("failed creating symlink for skill dir: %v", err)
		}
		if err := linker.SymlinkIfPossible(repoWorkflowFile, targetWorkflowFile); err != nil {
			t.Fatalf("failed creating symlink for workflow file: %v", err)
		}
	} else {
		t.Skip("OS filesystem does not support symlinks in test environment")
	}

	det, err := harness.NewDetector(harness.WithFs(fs), harness.WithHomeDir(homeDir))
	if err != nil {
		t.Fatalf("failed creating detector: %v", err)
	}

	engine, err := NewUninstallEngine(UninstallConfig{
		Fs:          fs,
		HomeDir:     homeDir,
		Detector:    det,
		RepoPath:    repoDir,
		PurgeConfig: true,
	})
	if err != nil {
		t.Fatalf("failed creating uninstall engine: %v", err)
	}

	discovered, err := engine.DiscoverSymlinks()
	if err != nil {
		t.Fatalf("DiscoverSymlinks failed: %v", err)
	}

	if len(discovered) != 2 {
		t.Fatalf("expected 2 discovered symlinks, got %d", len(discovered))
	}

	result, err := engine.Execute(discovered)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.RepoRemoved {
		t.Errorf("expected RepoRemoved to be true")
	}

	if !result.ConfigPurged {
		t.Errorf("expected ConfigPurged to be true")
	}

	// Verify restored targetSkillDir is now a real standalone directory containing SKILL.md
	skillStat, err := os.Lstat(targetSkillDir)
	if err != nil {
		t.Fatalf("failed to stat restored targetSkillDir: %v", err)
	}
	if skillStat.Mode()&os.ModeSymlink != 0 {
		t.Errorf("targetSkillDir is still a symlink after uninstall")
	}

	restoredSkillContent, err := afero.ReadFile(fs, filepath.Join(targetSkillDir, "SKILL.md"))
	if err != nil || string(restoredSkillContent) != "skill payload" {
		t.Errorf("restored skill file content mismatch: got %q", string(restoredSkillContent))
	}

	// Verify restored targetWorkflowFile is now a real standalone file
	workflowStat, err := os.Lstat(targetWorkflowFile)
	if err != nil {
		t.Fatalf("failed to stat restored targetWorkflowFile: %v", err)
	}
	if workflowStat.Mode()&os.ModeSymlink != 0 {
		t.Errorf("targetWorkflowFile is still a symlink after uninstall")
	}

	restoredWorkflowContent, err := afero.ReadFile(fs, targetWorkflowFile)
	if err != nil || string(restoredWorkflowContent) != "workflow payload" {
		t.Errorf("restored workflow file content mismatch: got %q", string(restoredWorkflowContent))
	}

	// Verify repository and config directory removal
	repoExists, _ := afero.Exists(fs, repoDir)
	if repoExists {
		t.Errorf("expected repository directory to be deleted")
	}

	configDirExists, _ := afero.Exists(fs, filepath.Join(homeDir, ".koharness"))
	if configDirExists {
		t.Errorf("expected ~/.koharness config directory to be purged")
	}
}

func TestUninstallEngine_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	fs := afero.NewOsFs()

	homeDir := filepath.Join(tmpDir, "user")
	repoDir := filepath.Join(homeDir, ".koharness", "repo")
	geminiSkills := filepath.Join(homeDir, ".gemini", "config", "skills")

	repoSkillFile := filepath.Join(repoDir, "skills", "test", "SKILL.md")
	_ = fs.MkdirAll(filepath.Dir(repoSkillFile), 0755)
	_ = afero.WriteFile(fs, repoSkillFile, []byte("test"), 0644)

	_ = fs.MkdirAll(geminiSkills, 0755)
	targetSkillDir := filepath.Join(geminiSkills, "test")
	repoSkillDir := filepath.Join(repoDir, "skills", "test")

	if linker, ok := fs.(afero.Symlinker); ok {
		_ = linker.SymlinkIfPossible(repoSkillDir, targetSkillDir)
	}

	det, _ := harness.NewDetector(harness.WithFs(fs), harness.WithHomeDir(homeDir))

	engine, err := NewUninstallEngine(UninstallConfig{
		Fs:       fs,
		HomeDir:  homeDir,
		Detector: det,
		RepoPath: repoDir,
		DryRun:   true,
	})
	if err != nil {
		t.Fatalf("failed creating uninstall engine: %v", err)
	}

	discovered, err := engine.DiscoverSymlinks()
	if err != nil || len(discovered) != 1 {
		t.Fatalf("expected 1 discovered symlink in dry-run, got %d (err: %v)", len(discovered), err)
	}

	result, err := engine.Execute(discovered)
	if err != nil {
		t.Fatalf("Execute dry-run failed: %v", err)
	}

	if len(result.RestoredAssets) != 1 {
		t.Errorf("expected 1 asset in dry-run result")
	}

	// Verify targetSkillDir remains a symlink in dry-run
	skillStat, err := os.Lstat(targetSkillDir)
	if err != nil || skillStat.Mode()&os.ModeSymlink == 0 {
		t.Errorf("targetSkillDir should remain a symlink during dry-run")
	}

	// Verify repoDir still exists in dry-run
	repoExists, _ := afero.Exists(fs, repoDir)
	if !repoExists {
		t.Errorf("repository directory should not be removed during dry-run")
	}
}

func TestUninstallEngine_RecordedOriginalHarnesses(t *testing.T) {
	tmpDir := t.TempDir()
	fs := afero.NewOsFs()

	homeDir := filepath.Join(tmpDir, "user")
	repoDir := filepath.Join(homeDir, ".koharness", "repo")
	geminiSkills := filepath.Join(homeDir, ".gemini", "config", "skills")
	claudeWorkflows := filepath.Join(homeDir, ".claude", "workflows")

	// Save global config with ONLY antigravity as original harness
	cfg := &harness.GlobalConfig{
		RepoPath:          repoDir,
		OriginalHarnesses: []string{"antigravity"},
	}
	if err := harness.SaveGlobalConfig(cfg, harness.PathOptions{Fs: fs, HomeDir: homeDir}); err != nil {
		t.Fatalf("failed saving global config: %v", err)
	}

	// Setup repository payload files
	repoSkillFile := filepath.Join(repoDir, "skills", "ipd", "SKILL.md")
	repoWorkflowFile := filepath.Join(repoDir, "workflows", "review.md")
	_ = fs.MkdirAll(filepath.Dir(repoSkillFile), 0755)
	_ = fs.MkdirAll(filepath.Dir(repoWorkflowFile), 0755)
	_ = afero.WriteFile(fs, repoSkillFile, []byte("skill payload"), 0644)
	_ = afero.WriteFile(fs, repoWorkflowFile, []byte("workflow payload"), 0644)

	// Setup target directories and symlinks
	_ = fs.MkdirAll(geminiSkills, 0755)
	_ = fs.MkdirAll(claudeWorkflows, 0755)

	targetSkillDir := filepath.Join(geminiSkills, "ipd")
	targetWorkflowFile := filepath.Join(claudeWorkflows, "review.md")

	if linker, ok := fs.(afero.Symlinker); ok {
		_ = linker.SymlinkIfPossible(filepath.Join(repoDir, "skills", "ipd"), targetSkillDir)
		_ = linker.SymlinkIfPossible(repoWorkflowFile, targetWorkflowFile)
	} else {
		t.Skip("OS filesystem does not support symlinks")
	}

	det, err := harness.NewDetector(harness.WithFs(fs), harness.WithHomeDir(homeDir))
	if err != nil {
		t.Fatalf("failed creating detector: %v", err)
	}

	engine, err := NewUninstallEngine(UninstallConfig{
		Fs:       fs,
		HomeDir:  homeDir,
		Detector: det,
		RepoPath: repoDir,
	})
	if err != nil {
		t.Fatalf("failed creating uninstall engine: %v", err)
	}

	discovered, err := engine.DiscoverSymlinks()
	if err != nil || len(discovered) != 2 {
		t.Fatalf("expected 2 discovered symlinks, got %d (err: %v)", len(discovered), err)
	}

	var geminiAsset, claudeAsset *RestoredAsset
	for i := range discovered {
		if discovered[i].HarnessID == "antigravity" {
			geminiAsset = &discovered[i]
		} else if discovered[i].HarnessID == "claude" {
			claudeAsset = &discovered[i]
		}
	}

	if geminiAsset == nil || !geminiAsset.WasOriginallyInstalled {
		t.Errorf("expected geminiAsset to have WasOriginallyInstalled = true")
	}
	if claudeAsset == nil || claudeAsset.WasOriginallyInstalled {
		t.Errorf("expected claudeAsset to have WasOriginallyInstalled = false")
	}

	result, err := engine.Execute(discovered)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(result.RestoredAssets) != 1 || result.RestoredAssets[0].HarnessID != "antigravity" {
		t.Errorf("expected 1 restored asset for antigravity, got %d", len(result.RestoredAssets))
	}
	if len(result.CleanedUpAssets) != 1 || result.CleanedUpAssets[0].HarnessID != "claude" {
		t.Errorf("expected 1 cleaned up asset for claude, got %d", len(result.CleanedUpAssets))
	}

	// Verify gemini skill was restored to standalone directory
	skillStat, err := os.Lstat(targetSkillDir)
	if err != nil || skillStat.Mode()&os.ModeSymlink != 0 {
		t.Errorf("targetSkillDir should be standalone directory, not symlink")
	}

	// Verify claude directory was completely cleaned up (removed)
	claudeExists, _ := afero.Exists(fs, filepath.Join(homeDir, ".claude"))
	if claudeExists {
		t.Errorf("expected uninstalled harness directory ~/.claude to be completely removed")
	}
}
