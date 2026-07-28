package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// DefaultRepoDir is the relative directory under user home where dotfiles repositories are placed by default.
const DefaultRepoDir = ".koharness/repo"

// ConfigDir defines the directory under user home where koharness global config and state reside.
const ConfigDir = ".koharness"

// ConfigFileName defines the default configuration file name under ~/.koharness/
const ConfigFileName = "config.yaml"

// EnvKoharnessRepo defines the environment variable used to override the repository path.
const EnvKoharnessRepo = "KOHARNESS_REPO"

// PathOptions specifies optional configuration for resolving repository paths,
// allowing filesystem injection for testing.
type PathOptions struct {
	// Fs is the filesystem abstraction used for path expansion and checks.
	Fs afero.Fs
	// HomeDir overrides the user home directory if specified.
	HomeDir string
	// GetEnv is a function used to fetch environment variable values.
	GetEnv func(string) string
}

// GlobalConfig stores global CLI settings saved in ~/.koharness/config.yaml.
type GlobalConfig struct {
	RepoPath string `yaml:"repo_path,omitempty"`
}

// GetGlobalConfigPath returns the absolute path to ~/.koharness/config.yaml.
func GetGlobalConfigPath(homeDir string) string {
	return filepath.Join(homeDir, ConfigDir, ConfigFileName)
}

// LoadGlobalConfig loads the global configuration from ~/.koharness/config.yaml.
func LoadGlobalConfig(opts ...PathOptions) (*GlobalConfig, error) {
	var opt PathOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	homeDir := opt.HomeDir
	if homeDir == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve home directory: %w", err)
		}
		homeDir = h
	}

	configPath := GetGlobalConfigPath(homeDir)
	var data []byte
	var err error
	if opt.Fs != nil {
		data, err = afero.ReadFile(opt.Fs, configPath)
	} else {
		data, err = os.ReadFile(configPath)
	}

	if err != nil {
		if os.IsNotExist(err) {
			return &GlobalConfig{}, nil
		}
		return nil, err
	}

	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse global config %s: %w", configPath, err)
	}
	return &cfg, nil
}

// SaveGlobalConfig saves the global configuration into ~/.koharness/config.yaml.
func SaveGlobalConfig(cfg *GlobalConfig, opts ...PathOptions) error {
	var opt PathOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	homeDir := opt.HomeDir
	if homeDir == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to resolve home directory: %w", err)
		}
		homeDir = h
	}

	dir := filepath.Join(homeDir, ConfigDir)
	configPath := filepath.Join(dir, ConfigFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed marshaling global config: %w", err)
	}

	if opt.Fs != nil {
		if err := opt.Fs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed creating config dir: %w", err)
		}
		return afero.WriteFile(opt.Fs, configPath, data, 0644)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed creating config dir: %w", err)
	}
	return os.WriteFile(configPath, data, 0644)
}

// GetRepoPath resolves the absolute filesystem path for the user's local dotfiles repository.
//
// Resolution precedence:
// 1. Provided customPath parameter (if non-empty)
// 2. KOHARNESS_REPO environment variable (if non-empty)
// 3. repo_path saved in ~/.koharness/config.yaml (if non-empty)
// 4. Fallback default: ~/.koharness/repo
//
// Tilde prefix (~) is automatically expanded to the user's home directory.
func GetRepoPath(customPath string, opts ...PathOptions) (string, error) {
	var opt PathOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	getEnv := opt.GetEnv
	if getEnv == nil {
		getEnv = os.Getenv
	}

	homeDir := opt.HomeDir
	if homeDir == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to resolve user home directory: %w", err)
		}
		homeDir = h
	}

	target := customPath
	if target == "" {
		target = getEnv(EnvKoharnessRepo)
	}
	if target == "" {
		cfg, err := LoadGlobalConfig(opts...)
		if err == nil && cfg != nil && cfg.RepoPath != "" {
			target = cfg.RepoPath
		}
	}
	if target == "" {
		target = filepath.Join(homeDir, DefaultRepoDir)
	}

	expanded := ExpandTilde(target, homeDir)
	return filepath.Clean(expanded), nil
}

// ExpandTilde replaces a leading `~` or `~/` with the user's absolute home directory.
func ExpandTilde(path string, homeDir string) string {
	if path == "~" {
		return homeDir
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

