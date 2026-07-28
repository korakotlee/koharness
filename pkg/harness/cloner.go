// Package harness provides cross-harness capability detection, adapter management, repository path resolution, and cloning utilities.
package harness

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// ErrTargetDirectoryExists is returned when target repository directory exists and is non-empty without force flag.
var ErrTargetDirectoryExists = errors.New("target repository directory exists and is non-empty")

// ClonerOptions specifies options for repository cloning.
type ClonerOptions struct {
	// Fs is the filesystem abstraction used for path checks.
	Fs afero.Fs
	// Out is an optional writer for git clone output logs.
	Out io.Writer
}

// CloneRepository clones a remote Git repo or local repository path into targetPath.
//
// If targetPath exists and is non-empty:
//   - If force is false, returns ErrTargetDirectoryExists.
//   - If force is true, renames existing targetPath to targetPath.bak.<timestamp> before cloning.
func CloneRepository(repoURL string, targetPath string, force bool, opts ...ClonerOptions) error {
	var opt ClonerOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	fs := opt.Fs
	if fs == nil {
		fs = afero.NewOsFs()
	}

	targetPath = filepath.Clean(targetPath)

	exists, err := afero.Exists(fs, targetPath)
	if err != nil {
		return fmt.Errorf("failed checking target path: %w", err)
	}

	if exists {
		empty, err := isDirEmpty(fs, targetPath)
		if err != nil {
			return fmt.Errorf("failed checking target path contents: %w", err)
		}

		if !empty {
			if !force {
				return ErrTargetDirectoryExists
			}

			timestamp := time.Now().Format("20060102-150405")
			backupPath := fmt.Sprintf("%s.bak.%s", targetPath, timestamp)
			if err := os.Rename(targetPath, backupPath); err != nil {
				return fmt.Errorf("failed creating repository directory backup %s: %w", backupPath, err)
			}
		}
	}

	parentDir := filepath.Dir(targetPath)
	if err := fs.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed creating parent directory %s: %w", parentDir, err)
	}

	cmd := exec.Command("git", "clone", repoURL, targetPath)
	if opt.Out != nil {
		cmd.Stdout = opt.Out
		cmd.Stderr = opt.Out
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func isDirEmpty(fs afero.Fs, path string) (bool, error) {
	f, err := fs.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	names, err := f.Readdirnames(1)
	if err != nil {
		if errors.Is(err, io.EOF) || err.Error() == "EOF" {
			return true, nil
		}
		return false, err
	}
	return len(names) == 0, nil
}
