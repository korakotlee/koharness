// Package harvester provides discovery, scanning, and harvesting capabilities
// for local AI client harnesses (Google Antigravity, Claude Code, OpenAI Codex).
package harvester

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// LocalConfigFile defines the filename for machine-local ignore rules.
const LocalConfigFile = ".koharness.local.yaml"

// IgnoredItem defines a capability item marked as ignored on this machine.
type IgnoredItem struct {
	Type      harness.CapabilityType `yaml:"type"`
	Name      string                 `yaml:"name"`
	HarnessID harness.HarnessID      `yaml:"harness_id,omitempty"`
}

// LocalConfig represents the machine-specific configuration stored in .koharness.local.yaml.
type LocalConfig struct {
	IgnoredCapabilities []IgnoredItem `yaml:"ignored_capabilities"`
}

// LoadLocalConfig reads .koharness.local.yaml from the given repo directory.
func LoadLocalConfig(fs afero.Fs, repoPath string) (*LocalConfig, error) {
	configPath := filepath.Join(repoPath, LocalConfigFile)
	exists, err := afero.Exists(fs, configPath)
	if err != nil || !exists {
		return &LocalConfig{}, nil
	}

	data, err := afero.ReadFile(fs, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", LocalConfigFile, err)
	}

	var cfg LocalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", LocalConfigFile, err)
	}

	return &cfg, nil
}

// SaveLocalConfig writes local configuration to .koharness.local.yaml in repoPath and ensures .gitignore tracks it.
func SaveLocalConfig(fs afero.Fs, repoPath string, cfg *LocalConfig) error {
	configPath := filepath.Join(repoPath, LocalConfigFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal local config: %w", err)
	}

	if err := afero.WriteFile(fs, configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", LocalConfigFile, err)
	}

	return EnsureGitignore(fs, repoPath, LocalConfigFile)
}

// EnsureGitignore ensures targetEntry is listed in .gitignore under repoPath.
func EnsureGitignore(fs afero.Fs, repoPath string, targetEntry string) error {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	exists, err := afero.Exists(fs, gitignorePath)
	if err != nil {
		return err
	}

	var content string
	if exists {
		data, err := afero.ReadFile(fs, gitignorePath)
		if err != nil {
			return err
		}
		content = string(data)
		for _, line := range strings.Split(content, "\n") {
			if strings.TrimSpace(line) == targetEntry {
				return nil
			}
		}
	}

	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += targetEntry + "\n"

	return afero.WriteFile(fs, gitignorePath, []byte(content), 0644)
}

// IsIgnored checks whether a capability matches any ignored item in LocalConfig.
func (cfg *LocalConfig) IsIgnored(capType harness.CapabilityType, name string, harnessID harness.HarnessID) bool {
	for _, item := range cfg.IgnoredCapabilities {
		if item.Type == capType && item.Name == name {
			if item.HarnessID == "" || item.HarnessID == harnessID {
				return true
			}
		}
	}
	return false
}

// AddIgnore adds a new capability to ignored capabilities if not already present.
func (cfg *LocalConfig) AddIgnore(capType harness.CapabilityType, name string, harnessID harness.HarnessID) {
	if cfg.IsIgnored(capType, name, harnessID) {
		return
	}
	cfg.IgnoredCapabilities = append(cfg.IgnoredCapabilities, IgnoredItem{
		Type:      capType,
		Name:      name,
		HarnessID: harnessID,
	})
}
