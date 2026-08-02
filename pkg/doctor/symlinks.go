// Package doctor provides diagnostic engines to inspect developer environment health,
// verifying symlink target integrity and client AI harness config access across local systems.
package doctor

import (
	"fmt"
	"os"
	"path/filepath"
)

// SymlinkDiagnostic details the health status of a single symlink on the workstation.
type SymlinkDiagnostic struct {
	// LinkPath is the location of the symlink file.
	LinkPath string
	// TargetPath is the path the symlink points to.
	TargetPath string
	// IsSymlink is true if LinkPath is indeed a symbolic link.
	IsSymlink bool
	// IsBroken is true if the symlink target does not exist on disk.
	IsBroken bool
	// Error contains any error encountered while reading or resolving the symlink.
	Error string
}

// InspectSymlinks recursively walks the target directories and evaluates the health of all symbolic links found.
func InspectSymlinks(dirs []string) ([]SymlinkDiagnostic, error) {
	var diagnostics []SymlinkDiagnostic

	for _, dir := range dirs {
		info, err := os.Lstat(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			diagnostics = append(diagnostics, SymlinkDiagnostic{
				LinkPath: dir,
				Error:    fmt.Sprintf("failed accessing directory: %v", err),
			})
			continue
		}

		if !info.IsDir() {
			diag := checkSingleSymlink(dir)
			if diag != nil {
				diagnostics = append(diagnostics, *diag)
			}
			continue
		}

		err = filepath.Walk(dir, func(path string, fileInfo os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			diag := checkSingleSymlink(path)
			if diag != nil {
				diagnostics = append(diagnostics, *diag)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error inspecting symlinks in %s: %w", dir, err)
		}
	}

	return diagnostics, nil
}

func checkSingleSymlink(path string) *SymlinkDiagnostic {
	info, err := os.Lstat(path)
	if err != nil {
		return nil
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}

	target, err := os.Readlink(path)
	if err != nil {
		return &SymlinkDiagnostic{
			LinkPath:  path,
			IsSymlink: true,
			Error:     fmt.Sprintf("failed reading symlink target: %v", err),
		}
	}

	// Resolve relative target paths against symlink location
	resolvedTarget := target
	if !filepath.IsAbs(target) {
		resolvedTarget = filepath.Join(filepath.Dir(path), target)
	}

	_, statErr := os.Stat(resolvedTarget)
	isBroken := os.IsNotExist(statErr)

	return &SymlinkDiagnostic{
		LinkPath:   path,
		TargetPath: target,
		IsSymlink:  true,
		IsBroken:   isBroken,
	}
}
