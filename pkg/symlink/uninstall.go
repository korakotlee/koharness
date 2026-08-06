// Package symlink provides engines for atomic symlink creation, backup management,
// broken link detection, and uninstallation/offboarding restoration.
package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/korakotlee/koharness/pkg/fsutil"
	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/spf13/afero"
)

// RestoredAsset details a single symlink target slated for restoration to a physical standalone asset or cleanup.
type RestoredAsset struct {
	// TargetPath is the absolute path to the active symlink in the user's harness directory.
	TargetPath string
	// SourcePath is the absolute path to the target source file or directory inside the repository.
	SourcePath string
	// IsDir indicates whether the asset represents a directory payload.
	IsDir bool
	// HarnessID identifies the client harness (e.g. "antigravity", "claude", "codex").
	HarnessID string
	// WasOriginallyInstalled indicates whether this harness was present before koharness setup.
	WasOriginallyInstalled bool
}

// UninstallResult details the outcome of an uninstallation execution.
type UninstallResult struct {
	// RestoredAssets lists all symlink targets converted to standalone physical files or directories.
	RestoredAssets []RestoredAsset
	// CleanedUpAssets lists all symlink targets cleaned up because the harness was not originally installed.
	CleanedUpAssets []RestoredAsset
	// RepoRemoved indicates whether the managed repository directory was deleted.
	RepoRemoved bool
	// ConfigPurged indicates whether global koharness configuration was removed.
	ConfigPurged bool
}

// UninstallConfig configures operational parameters for UninstallEngine.
type UninstallConfig struct {
	// Fs specifies the filesystem abstraction (defaults to afero.NewOsFs()).
	Fs afero.Fs
	// HomeDir specifies the user home directory.
	HomeDir string
	// Detector manages discovery of client harness installation locations.
	Detector *harness.Detector
	// DryRun previews uninstallation actions without mutating the filesystem.
	DryRun bool
	// RepoPath specifies the target repository path to un-link and remove.
	RepoPath string
	// PurgeConfig specifies whether to remove global ~/.koharness configuration directory.
	PurgeConfig bool
}

// UninstallEngine inspects managed AI harness directories, locates active symlinks
// pointing into the koharness repository, atomically restores physical standalone assets,
// and safely removes the local repository clone.
type UninstallEngine struct {
	fs                afero.Fs
	homeDir           string
	detector          *harness.Detector
	dryRun            bool
	repoPath          string
	purgeConfig       bool
	originalHarnesses map[string]bool
	recordedHarnesses bool
}

// NewUninstallEngine constructs a new UninstallEngine instance with the provided configuration options.
func NewUninstallEngine(cfg UninstallConfig) (*UninstallEngine, error) {
	fs := cfg.Fs
	if fs == nil {
		fs = afero.NewOsFs()
	}

	homeDir := cfg.HomeDir
	if homeDir == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve user home directory: %w", err)
		}
		homeDir = h
	}

	repoPath := cfg.RepoPath
	if repoPath == "" {
		r, err := harness.GetRepoPath("", harness.PathOptions{Fs: fs, HomeDir: homeDir})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve repo path: %w", err)
		}
		repoPath = r
	}
	repoPath = harness.ExpandTilde(repoPath, homeDir)

	det := cfg.Detector
	if det == nil {
		d, err := harness.NewDetector(harness.WithFs(fs), harness.WithHomeDir(homeDir))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize detector: %w", err)
		}
		det = d
	}

	origMap := make(map[string]bool)
	recorded := false
	globalCfg, err := harness.LoadGlobalConfig(harness.PathOptions{Fs: fs, HomeDir: homeDir})
	if err == nil && globalCfg != nil && len(globalCfg.OriginalHarnesses) > 0 {
		recorded = true
		for _, h := range globalCfg.OriginalHarnesses {
			origMap[h] = true
		}
	}

	return &UninstallEngine{
		fs:                fs,
		homeDir:           homeDir,
		detector:          det,
		dryRun:            cfg.DryRun,
		repoPath:          filepath.Clean(repoPath),
		purgeConfig:       cfg.PurgeConfig,
		originalHarnesses: origMap,
		recordedHarnesses: recorded,
	}, nil
}

// DiscoverSymlinks traverses all detected harness directories and returns all active symlinks
// that point to targets inside the managed repository.
func (ue *UninstallEngine) DiscoverSymlinks() ([]RestoredAsset, error) {
	var discovered []RestoredAsset
	seen := make(map[string]bool)

	adapters := ue.detector.GetAdapters()
	for _, adapter := range adapters {
		paths := adapter.GetConfigPaths()
		searchDirs := []string{
			paths.SkillsDir,
			paths.WorkflowsDir,
			paths.ConfigDir,
		}

		for _, searchDir := range searchDirs {
			if searchDir == "" {
				continue
			}
			exists, err := afero.Exists(ue.fs, searchDir)
			if err != nil || !exists {
				continue
			}

			err = afero.Walk(ue.fs, searchDir, func(path string, info os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return nil // Continue walking despite individual file errors
				}

				cleanPath := filepath.Clean(path)
				if seen[cleanPath] {
					return nil
				}

				isSym, _, err := isSymlinkPath(ue.fs, path)
				if err != nil || !isSym {
					return nil
				}

				targetLink, err := readSymlinkPath(ue.fs, path)
				if err != nil {
					return nil
				}

				absTarget := targetLink
				if !filepath.IsAbs(targetLink) {
					absTarget = filepath.Join(filepath.Dir(path), targetLink)
				}
				absTarget = filepath.Clean(absTarget)

				rel, relErr := filepath.Rel(ue.repoPath, absTarget)
				if relErr == nil && !strings.HasPrefix(rel, "..") && rel != "." {
					seen[cleanPath] = true
					isDir := false
					srcInfo, statErr := ue.fs.Stat(absTarget)
					if statErr == nil && srcInfo.IsDir() {
						isDir = true
					}

					harnessID := string(adapter.ID())
					wasOriginallyInstalled := true
					if ue.recordedHarnesses {
						wasOriginallyInstalled = ue.originalHarnesses[harnessID]
					}

					discovered = append(discovered, RestoredAsset{
						TargetPath:             cleanPath,
						SourcePath:             absTarget,
						IsDir:                  isDir,
						HarnessID:              harnessID,
						WasOriginallyInstalled: wasOriginallyInstalled,
					})

					if info.IsDir() {
						return filepath.SkipDir
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed scanning harness directory %s: %w", searchDir, err)
			}
		}
	}

	return discovered, nil
}

// Execute converts the provided symlink assets into standalone physical files/directories for originally
// installed harnesses, removes uninstalled harness symlinks and empty directories, removes the managed
// repository directory, and optionally purges global configuration.
func (ue *UninstallEngine) Execute(assets []RestoredAsset) (*UninstallResult, error) {
	result := &UninstallResult{
		RestoredAssets:  make([]RestoredAsset, 0, len(assets)),
		CleanedUpAssets: make([]RestoredAsset, 0, len(assets)),
	}

	for _, asset := range assets {
		if asset.WasOriginallyInstalled {
			if !ue.dryRun {
				// Stage physical restoration to temporary location
				tempPath := fmt.Sprintf("%s.tmp_restore.%s", asset.TargetPath, generateRandomID())

				srcExists, _ := afero.Exists(ue.fs, asset.SourcePath)
				if srcExists {
					if asset.IsDir {
						if err := fsutil.CopyDir(ue.fs, asset.SourcePath, tempPath); err != nil {
							return nil, fmt.Errorf("failed copying source directory %s: %w", asset.SourcePath, err)
						}
					} else {
						if err := fsutil.CopyFile(ue.fs, asset.SourcePath, tempPath); err != nil {
							return nil, fmt.Errorf("failed copying source file %s: %w", asset.SourcePath, err)
						}
					}

					// Remove active symlink
					if err := ue.fs.RemoveAll(asset.TargetPath); err != nil {
						_ = ue.fs.RemoveAll(tempPath)
						return nil, fmt.Errorf("failed removing symlink at %s: %w", asset.TargetPath, err)
					}

					// Rename temporary copy to original target path
					if err := ue.fs.Rename(tempPath, asset.TargetPath); err != nil {
						return nil, fmt.Errorf("failed placing restored asset at %s: %w", asset.TargetPath, err)
					}
				} else {
					// Target in repo does not exist (broken symlink); cleanly remove broken symlink
					_ = ue.fs.RemoveAll(asset.TargetPath)
				}
			}
			result.RestoredAssets = append(result.RestoredAssets, asset)
		} else {
			if !ue.dryRun {
				if err := ue.fs.RemoveAll(asset.TargetPath); err != nil {
					return nil, fmt.Errorf("failed removing symlink at %s: %w", asset.TargetPath, err)
				}
				ue.cleanupEmptyParentDirs(asset.TargetPath, asset.HarnessID)
			}
			result.CleanedUpAssets = append(result.CleanedUpAssets, asset)
		}
	}

	// Remove repo directory if present
	repoExists, _ := afero.Exists(ue.fs, ue.repoPath)
	if repoExists {
		if !ue.dryRun {
			if err := ue.fs.RemoveAll(ue.repoPath); err != nil {
				return nil, fmt.Errorf("failed to remove repository directory %s: %w", ue.repoPath, err)
			}
		}
		result.RepoRemoved = true
	}

	// Purge global config directory if requested
	if ue.purgeConfig {
		configDir := filepath.Join(ue.homeDir, harness.ConfigDir)
		cfgExists, _ := afero.Exists(ue.fs, configDir)
		if cfgExists {
			if !ue.dryRun {
				if err := ue.fs.RemoveAll(configDir); err != nil {
					return nil, fmt.Errorf("failed to purge config directory %s: %w", configDir, err)
				}
			}
			result.ConfigPurged = true
		}
	}

	return result, nil
}

func (ue *UninstallEngine) cleanupEmptyParentDirs(targetPath string, harnessID string) {
	stopDir := ""
	if adapter, ok := ue.detector.GetAdapter(harness.HarnessID(harnessID)); ok {
		stopDir = filepath.Clean(adapter.GetConfigPaths().ConfigDir)
	}

	dir := filepath.Dir(targetPath)
	for {
		cleanDir := filepath.Clean(dir)
		if cleanDir == "." || cleanDir == "/" || cleanDir == ue.homeDir {
			break
		}

		entries, err := afero.ReadDir(ue.fs, cleanDir)
		if err != nil || len(entries) > 0 {
			break
		}

		_ = ue.fs.Remove(cleanDir)

		if stopDir != "" && (cleanDir == stopDir || strings.HasPrefix(stopDir, cleanDir)) {
			break
		}
		dir = filepath.Dir(cleanDir)
	}
}

func isSymlinkPath(fs afero.Fs, path string) (bool, os.FileInfo, error) {
	if lstater, ok := fs.(afero.Lstater); ok {
		fi, _, err := lstater.LstatIfPossible(path)
		if err == nil {
			return fi.Mode()&os.ModeSymlink != 0, fi, nil
		}
	}
	fi, err := os.Lstat(path)
	if err != nil {
		return false, nil, err
	}
	return fi.Mode()&os.ModeSymlink != 0, fi, nil
}

func readSymlinkPath(fs afero.Fs, path string) (string, error) {
	if lr, ok := fs.(interface {
		ReadlinkIfPossible(string) (string, error)
	}); ok {
		target, err := lr.ReadlinkIfPossible(path)
		if err == nil {
			return target, nil
		}
	}
	return os.Readlink(path)
}
