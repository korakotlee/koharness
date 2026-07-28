package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// DefaultRepoDir is the relative directory under user home where dotfiles repositories are placed by default.
const DefaultRepoDir = ".koharness/repo"

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

// GetRepoPath resolves the absolute filesystem path for the user's local dotfiles repository.
//
// Resolution precedence:
// 1. Provided customPath parameter (if non-empty)
// 2. KOHARNESS_REPO environment variable (if non-empty)
// 3. Fallback default: ~/.koharness/repo
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
