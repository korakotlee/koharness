// Package sync implements the core orchestration logic for updating local capability repositories and refreshing client AI harnesses.
package sync

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/credentials"
	"github.com/korakotlee/koharness/pkg/git"
	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/mcp"
	"github.com/korakotlee/koharness/pkg/symlink"
	"github.com/korakotlee/koharness/pkg/tui"
	"github.com/spf13/afero"
)

// ErrDirtyRepo indicates that the target repository has uncommitted local changes and sync was aborted.
var ErrDirtyRepo = errors.New("uncommitted local changes detected in repository")

// SyncOptions configures execution parameters for repository synchronization.
type SyncOptions struct {
	// RepoPath specifies the target repository path. If empty, the global configuration path is resolved.
	RepoPath string
	// HomeDir specifies the user home directory. If empty, os.UserHomeDir() is used.
	HomeDir string
	// NonInteractive disables interactive TUI warning prompts.
	NonInteractive bool
	// Fs is the file system abstraction. Defaults to afero.NewOsFs().
	Fs afero.Fs
	// Out specifies the writer for standard informational messages.
	Out io.Writer
	// ErrOut specifies the writer for warning and error messages.
	ErrOut io.Writer
}

// SyncResult details the outcome of a repository synchronization operation.
type SyncResult struct {
	// RepoPath is the cleaned path of the synchronized repository.
	RepoPath string
	// Branch is the active Git branch name.
	Branch string
	// LinkedCount is the number of capability symlinks updated or verified.
	LinkedCount int
	// MCPMerged indicates whether MCP configurations were re-merged.
	MCPMerged bool
}

// SyncEngine coordinates Git status inspection, interactive TUI warning screens, clean rebase pulling, and client harness re-linking.
type SyncEngine struct {
	opts SyncOptions
}

// NewSyncEngine initializes a SyncEngine with the given options.
func NewSyncEngine(opts SyncOptions) (*SyncEngine, error) {
	if opts.HomeDir == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		opts.HomeDir = h
	}
	if opts.Fs == nil {
		opts.Fs = afero.NewOsFs()
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.ErrOut == nil {
		opts.ErrOut = os.Stderr
	}

	if opts.RepoPath == "" {
		p, err := harness.GetRepoPath("")
		if err != nil {
			return nil, fmt.Errorf("failed resolving repository path: %w", err)
		}
		opts.RepoPath = p
	} else {
		opts.RepoPath = harness.ExpandTilde(opts.RepoPath, opts.HomeDir)
	}
	opts.RepoPath = filepath.Clean(opts.RepoPath)

	return &SyncEngine{opts: opts}, nil
}

// Run executes the sync workflow.
func (e *SyncEngine) Run(ctx context.Context) (*SyncResult, error) {
	// 1. Verify repo existence
	exists, err := afero.DirExists(e.opts.Fs, e.opts.RepoPath)
	if err != nil || !exists {
		return nil, fmt.Errorf("repository directory does not exist: %s", e.opts.RepoPath)
	}

	// 2. Inspect Git status
	status, err := git.InspectStatus(ctx, e.opts.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("git status check failed: %w", err)
	}

	// 3. Handle dirty repository state
	if status.IsDirty {
		if !e.opts.NonInteractive {
			// Run interactive TUI warning screen
			if err := tui.RunSyncWarningView(e.opts.RepoPath, status.DirtyFiles); err != nil {
				fmt.Fprintln(e.opts.ErrOut, tui.RenderSyncWarningText(e.opts.RepoPath, status.DirtyFiles))
			}
		} else {
			fmt.Fprintln(e.opts.ErrOut, tui.RenderSyncWarningText(e.opts.RepoPath, status.DirtyFiles))
		}
		return nil, ErrDirtyRepo
	}

	// 4. Perform git pull --rebase on clean repository
	fmt.Fprintln(e.opts.Out, tui.BadgeInfo("SYNCING REPOSITORY"), fmt.Sprintf("Pulling latest upstream updates into %s...", e.opts.RepoPath))
	if err := git.PullRebase(ctx, e.opts.RepoPath); err != nil {
		return nil, fmt.Errorf("failed pulling updates: %w", err)
	}

	// 5. Re-link capabilities and refresh MCP configurations
	linkedCount, mcpMerged, err := e.refreshHarnesses()
	if err != nil {
		return nil, fmt.Errorf("failed refreshing client harnesses: %w", err)
	}

	result := &SyncResult{
		RepoPath:    e.opts.RepoPath,
		Branch:      status.Branch,
		LinkedCount: linkedCount,
		MCPMerged:   mcpMerged,
	}

	return result, nil
}

func (e *SyncEngine) refreshHarnesses() (int, bool, error) {
	detector, err := harness.NewDetector(harness.WithHomeDir(e.opts.HomeDir))
	if err != nil {
		return 0, false, err
	}
	adapters := detector.GetAdapters()

	linker := symlink.NewLinkerEngine(symlink.LinkerConfig{
		Fs:      e.opts.Fs,
		HomeDir: e.opts.HomeDir,
	})

	linkedCount := 0
	mcpMerged := false

	categories := []string{"skills", "prompts", "workflows", "mcp"}
	for _, cat := range categories {
		catPath := filepath.Join(e.opts.RepoPath, cat)
		entries, err := afero.ReadDir(e.opts.Fs, catPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			assetRepoPath := filepath.Join(catPath, entry.Name())

			for _, adapter := range adapters {
				paths := adapter.GetConfigPaths()
				targetDir := ""
				switch cat {
				case "skills":
					targetDir = paths.SkillsDir
				case "prompts", "workflows":
					targetDir = paths.WorkflowsDir
				case "mcp":
					targetDir = paths.MCPDir
				}

				if targetDir == "" {
					continue
				}

				targetAssetPath := filepath.Join(targetDir, entry.Name())
				if cat == "mcp" && strings.HasSuffix(entry.Name(), ".json") {
					// Perform deep MCP config merge if target exists
					if targetExists, _ := afero.Exists(e.opts.Fs, targetAssetPath); targetExists {
						baseBytes, bErr := afero.ReadFile(e.opts.Fs, targetAssetPath)
						overrideBytes, oErr := afero.ReadFile(e.opts.Fs, assetRepoPath)
						if bErr == nil && oErr == nil {
							mergedBytes, mErr := mcp.MergeJSON(baseBytes, overrideBytes)
							if mErr == nil {
								resolvers := []credentials.CredentialResolver{credentials.NewProtonPassResolver()}
								warningHandler := func(warning string) {
									fmt.Fprintln(e.opts.ErrOut, tui.BadgeWarn("CREDENTIAL WARNING"), warning)
								}
								expandedBytes, expErr := mcp.ExpandJSONTokens(mergedBytes, mcp.EnvOptions{
									HomeDir:        e.opts.HomeDir,
									Resolvers:      resolvers,
									WarningHandler: warningHandler,
								})
								if expErr == nil {
									mergedBytes = expandedBytes
								}
								_ = afero.WriteFile(e.opts.Fs, targetAssetPath, mergedBytes, 0644)
								mcpMerged = true
							}
						}
					}
				}

				_, err := linker.RepairSymlink(assetRepoPath, targetAssetPath)
				if err == nil {
					linkedCount++
				}
			}
		}
	}

	// 5. Refresh client harness AGENTS.md memory navigation rules
	_ = harness.SyncMemoryNavigationRules(e.opts.Fs, e.opts.HomeDir, e.opts.RepoPath)

	return linkedCount, mcpMerged, nil
}

// PrintSyncSuccess outputs formatted completion details.
func PrintSyncSuccess(out io.Writer, res *SyncResult) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, tui.BadgeSuccess("SYNC COMPLETE"), lipgloss.NewStyle().Bold(true).Render("Repository updated and client AI harnesses refreshed cleanly!"))
	fmt.Fprintln(out, fmt.Sprintf("  Repository Path: %s", res.RepoPath))
	if res.Branch != "" {
		fmt.Fprintln(out, fmt.Sprintf("  Active Branch:   %s", res.Branch))
	}
	fmt.Fprintln(out, fmt.Sprintf("  Updated Links:   %d capability symlink(s)", res.LinkedCount))
	if res.MCPMerged {
		fmt.Fprintln(out, "  MCP Configs:     Merged successfully")
	}
	fmt.Fprintln(out)
}
