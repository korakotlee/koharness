package harness_test

import (
	"path/filepath"
	"testing"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/spf13/afero"
)

func TestDetector_EmptyEnvironment(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/testuser"

	detector, err := harness.NewDetector(
		harness.WithFs(fs),
		harness.WithHomeDir(homeDir),
	)
	if err != nil {
		t.Fatalf("failed to create detector: %v", err)
	}

	installed := detector.DetectInstalled()
	if len(installed) != 0 {
		t.Errorf("expected 0 installed harnesses in empty environment, got %d", len(installed))
	}

	statuses := detector.DetectAll()
	if len(statuses) != 3 {
		t.Fatalf("expected 3 statuses, got %d", len(statuses))
	}

	for _, s := range statuses {
		if s.Installed {
			t.Errorf("expected harness %s to be uninstalled", s.ID)
		}
		if len(s.PathsFound) != 0 {
			t.Errorf("expected 0 found paths for %s, got %d", s.ID, len(s.PathsFound))
		}
		if len(s.PathsMissing) == 0 {
			t.Errorf("expected missing paths listed for %s", s.ID)
		}
	}
}

func TestDetector_PartialEnvironment(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/testuser"

	// Create Antigravity skills dir only
	antigravitySkills := filepath.Join(homeDir, ".gemini", "config", "skills")
	if err := fs.MkdirAll(antigravitySkills, 0755); err != nil {
		t.Fatalf("failed to create mock directory: %v", err)
	}

	// Create Claude config file only
	claudeConfigFile := filepath.Join(homeDir, ".claude.json")
	if err := afero.WriteFile(fs, claudeConfigFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create mock config file: %v", err)
	}

	detector, err := harness.NewDetector(
		harness.WithFs(fs),
		harness.WithHomeDir(homeDir),
	)
	if err != nil {
		t.Fatalf("failed to create detector: %v", err)
	}

	installed := detector.DetectInstalled()
	if len(installed) != 2 {
		t.Errorf("expected 2 installed harnesses (Antigravity, Claude), got %d", len(installed))
	}

	// Verify Antigravity status
	adapter, ok := detector.GetAdapter(harness.HarnessAntigravity)
	if !ok || !adapter.IsInstalled() {
		t.Errorf("expected Antigravity to be detected as installed")
	}

	// Verify Claude status
	adapter, ok = detector.GetAdapter(harness.HarnessClaude)
	if !ok || !adapter.IsInstalled() {
		t.Errorf("expected Claude Code to be detected as installed")
	}

	// Verify Codex status
	adapter, ok = detector.GetAdapter(harness.HarnessCodex)
	if !ok || adapter.IsInstalled() {
		t.Errorf("expected Codex to be detected as uninstalled")
	}
}

func TestDetector_FullyInstalledEnvironment(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/testuser"

	// Setup Antigravity
	_ = fs.MkdirAll(filepath.Join(homeDir, ".gemini", "config", "skills"), 0755)
	_ = fs.MkdirAll(filepath.Join(homeDir, ".gemini", "config", "global_workflows"), 0755)
	_ = fs.MkdirAll(filepath.Join(homeDir, ".gemini", "antigravity-ide", "mcp"), 0755)

	// Setup Claude Code
	_ = afero.WriteFile(fs, filepath.Join(homeDir, ".claude.json"), []byte("{}"), 0644)
	_ = fs.MkdirAll(filepath.Join(homeDir, ".claude", "skills"), 0755)

	// Setup OpenAI Codex
	_ = fs.MkdirAll(filepath.Join(homeDir, ".codex", "skills"), 0755)
	_ = fs.MkdirAll(filepath.Join(homeDir, ".codex", "mcp"), 0755)

	detector, err := harness.NewDetector(
		harness.WithFs(fs),
		harness.WithHomeDir(homeDir),
	)
	if err != nil {
		t.Fatalf("failed to create detector: %v", err)
	}

	installed := detector.DetectInstalled()
	if len(installed) != 3 {
		t.Errorf("expected 3 installed harnesses, got %d", len(installed))
	}

	adapters := detector.GetAdapters()
	if len(adapters) != 3 {
		t.Errorf("expected 3 total adapters registered, got %d", len(adapters))
	}
}

func TestAdapters_GetConfigPaths(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/testuser"

	detector, err := harness.NewDetector(
		harness.WithFs(fs),
		harness.WithHomeDir(homeDir),
	)
	if err != nil {
		t.Fatalf("failed to create detector: %v", err)
	}

	// Antigravity check
	anti, _ := detector.GetAdapter(harness.HarnessAntigravity)
	if anti.Name() != "Google Antigravity" {
		t.Errorf("unexpected name %q", anti.Name())
	}
	antiPaths := anti.GetConfigPaths()
	if antiPaths.SkillsDir != filepath.Join(homeDir, ".gemini", "config", "skills") {
		t.Errorf("unexpected skills dir %q", antiPaths.SkillsDir)
	}

	// Claude check
	claude, _ := detector.GetAdapter(harness.HarnessClaude)
	if claude.Name() != "Claude Code" {
		t.Errorf("unexpected name %q", claude.Name())
	}
	claudePaths := claude.GetConfigPaths()
	if claudePaths.ConfigFile != filepath.Join(homeDir, ".claude.json") {
		t.Errorf("unexpected config file %q", claudePaths.ConfigFile)
	}

	// Codex check
	codex, _ := detector.GetAdapter(harness.HarnessCodex)
	if codex.Name() != "OpenAI Codex" {
		t.Errorf("unexpected name %q", codex.Name())
	}
	codexPaths := codex.GetConfigPaths()
	if codexPaths.MCPDir != filepath.Join(homeDir, ".codex", "mcp") {
		t.Errorf("unexpected MCP dir %q", codexPaths.MCPDir)
	}
}

func TestAdapters_LinkCapability(t *testing.T) {
	fs := afero.NewMemMapFs()
	homeDir := "/home/testuser"

	detector, err := harness.NewDetector(
		harness.WithFs(fs),
		harness.WithHomeDir(homeDir),
	)
	if err != nil {
		t.Fatalf("failed to create detector: %v", err)
	}

	// Test Antigravity LinkCapability
	anti, _ := detector.GetAdapter(harness.HarnessAntigravity)
	err = anti.LinkCapability(harness.Capability{Type: harness.CapabilitySkill, Name: "test-skill"})
	if err != nil {
		t.Errorf("failed linking skill to Antigravity: %v", err)
	}
	err = anti.LinkCapability(harness.Capability{Type: harness.CapabilityWorkflow, Name: "test-wf"})
	if err != nil {
		t.Errorf("failed linking workflow to Antigravity: %v", err)
	}
	err = anti.LinkCapability(harness.Capability{Type: harness.CapabilityMCP, Name: "test-mcp"})
	if err != nil {
		t.Errorf("failed linking mcp to Antigravity: %v", err)
	}
	err = anti.LinkCapability(harness.Capability{Type: harness.CapabilityType("unsupported"), Name: "invalid"})
	if err == nil {
		t.Errorf("expected error linking unsupported capability to Antigravity")
	}

	// Test Claude LinkCapability
	claude, _ := detector.GetAdapter(harness.HarnessClaude)
	err = claude.LinkCapability(harness.Capability{Type: harness.CapabilitySkill, Name: "test-skill"})
	if err != nil {
		t.Errorf("failed linking skill to Claude: %v", err)
	}
	err = claude.LinkCapability(harness.Capability{Type: harness.CapabilityMCP, Name: "test-mcp"})
	if err != nil {
		t.Errorf("failed linking mcp to Claude: %v", err)
	}
	err = claude.LinkCapability(harness.Capability{Type: harness.CapabilityWorkflow, Name: "test-wf"})
	if err == nil {
		t.Errorf("expected error linking unsupported workflow to Claude")
	}

	// Test Codex LinkCapability
	codex, _ := detector.GetAdapter(harness.HarnessCodex)
	err = codex.LinkCapability(harness.Capability{Type: harness.CapabilitySkill, Name: "test-skill"})
	if err != nil {
		t.Errorf("failed linking skill to Codex: %v", err)
	}
	err = codex.LinkCapability(harness.Capability{Type: harness.CapabilityMCP, Name: "test-mcp"})
	if err != nil {
		t.Errorf("failed linking mcp to Codex: %v", err)
	}
	err = codex.LinkCapability(harness.Capability{Type: harness.CapabilityWorkflow, Name: "test-wf"})
	if err == nil {
		t.Errorf("expected error linking unsupported workflow to Codex")
	}
}

func TestDetector_DefaultOptions(t *testing.T) {
	detector, err := harness.NewDetector()
	if err != nil {
		t.Fatalf("failed creating detector with default options: %v", err)
	}

	// Non-existent dummy ID check
	_, ok := detector.GetAdapter(harness.HarnessID("unknown"))
	if ok {
		t.Errorf("expected unknown harness ID to return false")
	}
}
