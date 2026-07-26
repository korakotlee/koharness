package harness

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// ClaudeAdapter implements HarnessAdapter for Claude Code.
type ClaudeAdapter struct {
	fs      afero.Fs
	homeDir string
}

// NewClaudeAdapter creates a new ClaudeAdapter instance with the given filesystem and home directory.
func NewClaudeAdapter(fs afero.Fs, homeDir string) *ClaudeAdapter {
	return &ClaudeAdapter{
		fs:      fs,
		homeDir: homeDir,
	}
}

// ID returns the unique identifier for Claude Code.
func (a *ClaudeAdapter) ID() HarnessID {
	return HarnessClaude
}

// Name returns the display title for Claude Code.
func (a *ClaudeAdapter) Name() string {
	return "Claude Code"
}

// GetConfigPaths returns the absolute configuration and capabilities paths for Claude Code.
func (a *ClaudeAdapter) GetConfigPaths() HarnessPaths {
	baseDir := filepath.Join(a.homeDir, ".claude")
	configFile := filepath.Join(a.homeDir, ".claude.json")
	return HarnessPaths{
		ConfigDir:    baseDir,
		ConfigFile:   configFile,
		SkillsDir:    filepath.Join(baseDir, "skills"),
		WorkflowsDir: "",
		MCPDir:       configFile,
	}
}

// IsInstalled returns true if either ~/.claude.json or ~/.claude/ directory exists.
func (a *ClaudeAdapter) IsInstalled() bool {
	paths := a.GetConfigPaths()
	checkPaths := []string{paths.ConfigFile, paths.ConfigDir}

	for _, p := range checkPaths {
		exists, err := afero.Exists(a.fs, p)
		if err == nil && exists {
			return true
		}
	}
	return false
}

// GetStatus returns the complete installation status breakdown for Claude Code.
func (a *ClaudeAdapter) GetStatus() HarnessStatus {
	paths := a.GetConfigPaths()
	checkPaths := []string{paths.ConfigFile, paths.ConfigDir, paths.SkillsDir}

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

// LinkCapability establishes linkage for a capability asset into Claude Code configuration paths.
func (a *ClaudeAdapter) LinkCapability(cap Capability) error {
	paths := a.GetConfigPaths()

	switch cap.Type {
	case CapabilitySkill:
		if paths.SkillsDir == "" {
			return fmt.Errorf("skills directory not defined for Claude Code")
		}
		return a.fs.MkdirAll(paths.SkillsDir, 0755)
	case CapabilityMCP:
		if paths.ConfigFile == "" {
			return fmt.Errorf("config file path not defined for Claude Code")
		}
		// Ensure parent directory of config file exists if needed
		dir := filepath.Dir(paths.ConfigFile)
		return a.fs.MkdirAll(dir, 0755)
	default:
		return fmt.Errorf("unsupported capability type %q for Claude Code", cap.Type)
	}
}
