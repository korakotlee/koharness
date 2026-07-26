package harness

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// AntigravityAdapter implements HarnessAdapter for Google Antigravity.
type AntigravityAdapter struct {
	fs      afero.Fs
	homeDir string
}

// NewAntigravityAdapter creates a new AntigravityAdapter instance with the given filesystem and home directory.
func NewAntigravityAdapter(fs afero.Fs, homeDir string) *AntigravityAdapter {
	return &AntigravityAdapter{
		fs:      fs,
		homeDir: homeDir,
	}
}

// ID returns the unique identifier for Google Antigravity.
func (a *AntigravityAdapter) ID() HarnessID {
	return HarnessAntigravity
}

// Name returns the display title for Google Antigravity.
func (a *AntigravityAdapter) Name() string {
	return "Google Antigravity"
}

// GetConfigPaths returns the absolute configuration and capabilities directories for Antigravity.
func (a *AntigravityAdapter) GetConfigPaths() HarnessPaths {
	base := filepath.Join(a.homeDir, ".gemini")
	return HarnessPaths{
		ConfigDir:    base,
		ConfigFile:   "",
		SkillsDir:    filepath.Join(base, "config", "skills"),
		WorkflowsDir: filepath.Join(base, "config", "global_workflows"),
		MCPDir:       filepath.Join(base, "antigravity-ide", "mcp"),
	}
}

// IsInstalled returns true if any of the Antigravity configuration directory footprints exist.
func (a *AntigravityAdapter) IsInstalled() bool {
	paths := a.GetConfigPaths()
	checkPaths := []string{paths.SkillsDir, paths.WorkflowsDir, paths.MCPDir}

	for _, p := range checkPaths {
		exists, err := afero.Exists(a.fs, p)
		if err == nil && exists {
			return true
		}
	}
	return false
}

// GetStatus returns the complete installation status breakdown for Google Antigravity.
func (a *AntigravityAdapter) GetStatus() HarnessStatus {
	paths := a.GetConfigPaths()
	checkPaths := []string{paths.SkillsDir, paths.WorkflowsDir, paths.MCPDir}

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

// LinkCapability establishes linkage for a capability asset into Antigravity target directories.
func (a *AntigravityAdapter) LinkCapability(cap Capability) error {
	paths := a.GetConfigPaths()
	var targetDir string

	switch cap.Type {
	case CapabilitySkill:
		targetDir = paths.SkillsDir
	case CapabilityWorkflow:
		targetDir = paths.WorkflowsDir
	case CapabilityMCP:
		targetDir = paths.MCPDir
	default:
		return fmt.Errorf("unsupported capability type %q for Antigravity", cap.Type)
	}

	if targetDir == "" {
		return fmt.Errorf("target path for capability type %q is not defined for Antigravity", cap.Type)
	}

	return a.fs.MkdirAll(targetDir, 0755)
}
