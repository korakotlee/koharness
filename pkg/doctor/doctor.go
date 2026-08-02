package doctor

import (
	"fmt"
	"os"
	"path/filepath"
)

// DoctorOptions provides operational parameters for executing environment diagnostics.
type DoctorOptions struct {
	// HomeDir overrides the user home directory location for testing or custom environments.
	HomeDir string
	// RepoRoot specifies the target repository path for inspecting memory layout and link integrity.
	RepoRoot string
	// SymlinkTargetDirs overrides specific directories to scan for symlinks.
	SymlinkTargetDirs []string
}

// DoctorResult aggregates all diagnostic check results.
type DoctorResult struct {
	SymlinkDiagnostics []SymlinkDiagnostic
	HarnessStatuses    []HarnessHealthStatus
	Memory             *MemoryDiagnostic
}

// HasBrokenSymlinks returns true if any inspected symlinks are broken.
func (r *DoctorResult) HasBrokenSymlinks() bool {
	for _, diag := range r.SymlinkDiagnostics {
		if diag.IsBroken {
			return true
		}
	}
	return false
}

// BrokenSymlinkCount returns the total number of broken symlinks detected.
func (r *DoctorResult) BrokenSymlinkCount() int {
	count := 0
	for _, diag := range r.SymlinkDiagnostics {
		if diag.IsBroken {
			count++
		}
	}
	return count
}

// Run executes environment health inspections including symlink target integrity, client harness configuration access, and 3-layer agent memory health.
func Run(opts DoctorOptions) (*DoctorResult, error) {
	homeDir := opts.HomeDir
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed resolving user home directory: %w", err)
		}
	}

	symlinkDirs := opts.SymlinkTargetDirs
	if len(symlinkDirs) == 0 {
		// Default to checking standard client harness config directories under homeDir
		symlinkDirs = []string{
			filepath.Join(homeDir, ".gemini"),
			filepath.Join(homeDir, ".claude"),
			filepath.Join(homeDir, ".codex"),
			filepath.Join(homeDir, ".koharness"),
		}
	}

	symlinkDiags, err := InspectSymlinks(symlinkDirs)
	if err != nil {
		return nil, fmt.Errorf("symlink health check failed: %w", err)
	}

	harnessStatuses, err := InspectHarnesses(homeDir)
	if err != nil {
		return nil, fmt.Errorf("harness diagnostic check failed: %w", err)
	}

	var memDiag *MemoryDiagnostic
	if opts.RepoRoot != "" {
		m, mErr := InspectMemory(opts.RepoRoot)
		if mErr == nil {
			memDiag = m
		}
	}

	// Filter out non-symlink diagnostic entries if any
	var actualSymlinkDiags []SymlinkDiagnostic
	for _, diag := range symlinkDiags {
		if diag.IsSymlink {
			actualSymlinkDiags = append(actualSymlinkDiags, diag)
		}
	}

	return &DoctorResult{
		SymlinkDiagnostics: actualSymlinkDiags,
		HarnessStatuses:    harnessStatuses,
		Memory:             memDiag,
	}, nil
}
