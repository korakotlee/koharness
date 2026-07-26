package harness

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// CodexAdapter implements HarnessAdapter for OpenAI Codex.
type CodexAdapter struct {
	fs      afero.Fs
	homeDir string
}

// NewCodexAdapter creates a new CodexAdapter instance with the given filesystem and home directory.
func NewCodexAdapter(fs afero.Fs, homeDir string) *CodexAdapter {
	return &CodexAdapter{
		fs:      fs,
		homeDir: homeDir,
	}
}

// ID returns the unique identifier for OpenAI Codex.
func (a *CodexAdapter) ID() HarnessID {
	return HarnessCodex
}

// Name returns the display title for OpenAI Codex.
func (a *CodexAdapter) Name() string {
	return "OpenAI Codex"
}

// GetConfigPaths returns the absolute configuration and capabilities paths for OpenAI Codex.
func (a *CodexAdapter) GetConfigPaths() HarnessPaths {
	base := filepath.Join(a.homeDir, ".codex")
	return HarnessPaths{
		ConfigDir:    base,
		ConfigFile:   "",
		SkillsDir:    filepath.Join(base, "skills"),
		WorkflowsDir: "",
		MCPDir:       filepath.Join(base, "mcp"),
	}
}

// IsInstalled returns true if the ~/.codex/ directory exists.
func (a *CodexAdapter) IsInstalled() bool {
	paths := a.GetConfigPaths()
	exists, err := afero.Exists(a.fs, paths.ConfigDir)
	return err == nil && exists
}

// GetStatus returns the complete installation status breakdown for OpenAI Codex.
func (a *CodexAdapter) GetStatus() HarnessStatus {
	paths := a.GetConfigPaths()
	checkPaths := []string{paths.ConfigDir, paths.SkillsDir, paths.MCPDir}

	var found []string
	var missing []string

	for _, p := range checkPaths {
		exists, err := afero.Exists(a.fs, p)
		if err == nil && exists {
			found = append(found, p)
		} else {
			missing = append(missing, p)
		}
	}

	return HarnessStatus{
		ID:           a.ID(),
		Name:         a.Name(),
		Installed:    len(found) > 0,
		PathsFound:   found,
		PathsMissing: missing,
	}
}

// LinkCapability establishes linkage for a capability asset into OpenAI Codex configuration paths.
func (a *CodexAdapter) LinkCapability(cap Capability) error {
	paths := a.GetConfigPaths()
	var targetDir string

	switch cap.Type {
	case CapabilitySkill:
		targetDir = paths.SkillsDir
	case CapabilityMCP:
		targetDir = paths.MCPDir
	default:
		return fmt.Errorf("unsupported capability type %q for OpenAI Codex", cap.Type)
	}

	if targetDir == "" {
		return fmt.Errorf("target path for capability type %q is not defined for OpenAI Codex", cap.Type)
	}

	return a.fs.MkdirAll(targetDir, 0755)
}
