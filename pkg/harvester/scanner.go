// Package harvester provides discovery, scanning, and harvesting capabilities
// for local AI client harnesses (Google Antigravity, Claude Code, OpenAI Codex).
package harvester

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/mcp"
	"github.com/spf13/afero"
)

// HarvestState represents the 3 selection states for a discovered capability.
type HarvestState int

const (
	StateImport HarvestState = iota
	StateSkip
	StateIgnore
)

// DiscoveredCapability represents a standalone capability discovered during workstation scanning.
type DiscoveredCapability struct {
	// HarnessID identifies which client harness owns this asset.
	HarnessID harness.HarnessID
	// Type specifies the category of the discovered asset (skill, workflow, prompt, mcp).
	Type harness.CapabilityType
	// Name specifies the unique identifier or file/folder name of the capability.
	Name string
	// SourcePath specifies the absolute filesystem path where the asset was found.
	SourcePath string
	// Selected indicates whether the user opted to harvest this item into the repository.
	Selected bool
	// Ignored indicates whether the user opted to ignore this item locally on this machine.
	Ignored bool
	// IsSecret indicates whether the capability contains sensitive keys/tokens to keep as local override.
	IsSecret bool
}

// GetState returns the current 3-state HarvestState of the capability.
func (c *DiscoveredCapability) GetState() HarvestState {
	if c.Ignored {
		return StateIgnore
	}
	if c.Selected {
		return StateImport
	}
	return StateSkip
}

// SetState updates Selected and Ignored fields based on the provided HarvestState.
func (c *DiscoveredCapability) SetState(state HarvestState) {
	switch state {
	case StateImport:
		c.Selected = true
		c.Ignored = false
	case StateSkip:
		c.Selected = false
		c.Ignored = false
	case StateIgnore:
		c.Selected = false
		c.Ignored = true
	}
}

// ToggleImportSkip toggles between StateImport and StateSkip.
func (c *DiscoveredCapability) ToggleImportSkip() {
	if c.GetState() == StateImport {
		c.SetState(StateSkip)
	} else {
		c.SetState(StateImport)
	}
}

// ToggleIgnore toggles between StateIgnore and StateSkip.
func (c *DiscoveredCapability) ToggleIgnore() {
	if c.GetState() == StateIgnore {
		c.SetState(StateSkip)
	} else {
		c.SetState(StateIgnore)
	}
}


// ScannerOption configures operational parameters for Scanner instances.
type ScannerOption func(*Scanner)

// WithFs sets a custom afero.Fs filesystem abstraction for the scanner.
func WithFs(fs afero.Fs) ScannerOption {
	return func(s *Scanner) {
		s.fs = fs
	}
}

// WithHomeDir sets a custom home directory override for the scanner.
func WithHomeDir(dir string) ScannerOption {
	return func(s *Scanner) {
		s.homeDir = dir
	}
}

// Scanner manages discovery and extraction of unmanaged capabilities across local harnesses.
type Scanner struct {
	fs       afero.Fs
	homeDir  string
	detector *harness.Detector
}

// NewScanner initializes a new Scanner using default OS filesystem or provided options.
func NewScanner(opts ...ScannerOption) (*Scanner, error) {
	s := &Scanner{
		fs: afero.NewOsFs(),
	}

	home, err := os.UserHomeDir()
	if err == nil {
		s.homeDir = home
	}

	for _, opt := range opts {
		opt(s)
	}

	det, err := harness.NewDetector(
		harness.WithFs(s.fs),
		harness.WithHomeDir(s.homeDir),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create detector for scanner: %w", err)
	}
	s.detector = det

	return s, nil
}

// ScanAll evaluates all registered client harness paths and returns discovered standalone capabilities.
func (s *Scanner) ScanAll() ([]DiscoveredCapability, error) {
	var results []DiscoveredCapability

	for _, adapter := range s.detector.GetAdapters() {
		caps, err := s.ScanAdapter(adapter)
		if err != nil {
			return nil, err
		}
		results = append(results, caps...)
	}

	return results, nil
}

// ScanForRepo scans all client harnesses and filters out capabilities that are ignored in .koharness.local.yaml or already present/symlinked in repoPath.
func (s *Scanner) ScanForRepo(repoPath string) ([]DiscoveredCapability, error) {
	all, err := s.ScanAll()
	if err != nil {
		return nil, err
	}

	localCfg, err := LoadLocalConfig(s.fs, repoPath)
	if err != nil {
		return nil, err
	}

	var filtered []DiscoveredCapability
	for _, item := range all {
		if localCfg.IsIgnored(item.Type, item.Name, item.HarnessID) {
			continue
		}

		if s.isAlreadyTrackedOrSymlinked(item, repoPath) {
			continue
		}

		filtered = append(filtered, item)
	}

	return filtered, nil
}

func (s *Scanner) isAlreadyTrackedOrSymlinked(item DiscoveredCapability, repoPath string) bool {
	if lstater, ok := s.fs.(afero.Lstater); ok {
		info, _, err := lstater.LstatIfPossible(item.SourcePath)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			return true
		}
	} else {
		if info, err := os.Lstat(item.SourcePath); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return true
		}
	}

	var targetSubDir string
	switch item.Type {
	case harness.CapabilitySkill:
		targetSubDir = "skills"
	case harness.CapabilityWorkflow, harness.CapabilityPrompt:
		targetSubDir = "prompts"
	case harness.CapabilityMCP:
		targetSubDir = "mcp"
	default:
		targetSubDir = "skills"
	}

	targetRepoPath := filepath.Join(repoPath, targetSubDir, item.Name)
	exists, _ := afero.Exists(s.fs, targetRepoPath)
	return exists
}


// ScanAdapter discovers capabilities for a specific client harness adapter.
func (s *Scanner) ScanAdapter(adapter harness.HarnessAdapter) ([]DiscoveredCapability, error) {
	var items []DiscoveredCapability
	paths := adapter.GetConfigPaths()

	// Scan Skills directory
	if paths.SkillsDir != "" {
		skills, err := s.scanDirectory(adapter.ID(), harness.CapabilitySkill, paths.SkillsDir)
		if err == nil {
			items = append(items, skills...)
		}
	}

	// Scan Workflows directory
	if paths.WorkflowsDir != "" {
		workflows, err := s.scanDirectory(adapter.ID(), harness.CapabilityWorkflow, paths.WorkflowsDir)
		if err == nil {
			items = append(items, workflows...)
		}
	}

	// Scan MCP directory or config file
	if paths.MCPDir != "" {
		mcpItems, err := s.scanMCP(adapter.ID(), paths.MCPDir)
		if err == nil {
			items = append(items, mcpItems...)
		}
	}

	return items, nil
}

func (s *Scanner) scanDirectory(harnessID harness.HarnessID, capType harness.CapabilityType, dir string) ([]DiscoveredCapability, error) {
	exists, err := afero.Exists(s.fs, dir)
	if err != nil || !exists {
		return nil, nil
	}

	entries, err := afero.ReadDir(s.fs, dir)
	if err != nil {
		return nil, err
	}

	var results []DiscoveredCapability
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		results = append(results, DiscoveredCapability{
			HarnessID:  harnessID,
			Type:       capType,
			Name:       name,
			SourcePath: filepath.Join(dir, name),
			Selected:   true,
			IsSecret:   false,
		})
	}
	return results, nil
}

func (s *Scanner) scanMCP(harnessID harness.HarnessID, mcpPath string) ([]DiscoveredCapability, error) {
	exists, err := afero.Exists(s.fs, mcpPath)
	if err != nil || !exists {
		return nil, nil
	}

	info, err := s.fs.Stat(mcpPath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return s.scanDirectory(harnessID, harness.CapabilityMCP, mcpPath)
	}

	// If it's a JSON config file (e.g., ~/.claude.json), parse mcpServers
	data, err := afero.ReadFile(s.fs, mcpPath)
	if err != nil {
		return nil, err
	}

	var cfg struct {
		MCPServers map[string]interface{} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, nil
	}

	var results []DiscoveredCapability
	for serverName, serverDef := range cfg.MCPServers {
		isSecret := false
		if serverBytes, err := json.Marshal(serverDef); err == nil {
			issues, _ := mcp.ValidateConfig(serverBytes)
			if len(issues) > 0 {
				isSecret = true
			}
		}
		results = append(results, DiscoveredCapability{
			HarnessID:  harnessID,
			Type:       harness.CapabilityMCP,
			Name:       serverName,
			SourcePath: mcpPath,
			Selected:   true,
			IsSecret:   isSecret,
		})
	}
	return results, nil
}
