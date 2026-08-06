// Package cmd implements CLI subcommands for the KoHarness application.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/harvester"
	"github.com/korakotlee/koharness/pkg/symlink"
	"github.com/korakotlee/koharness/pkg/tui"
	"github.com/spf13/cobra"
)

var (
	initPathOverride   string
	initForce          bool
	initNonInteractive bool
)

// InitCmd represents the `koharness init` subcommand.
var InitCmd = &cobra.Command{
	Use:   "init <repo-url-or-path> [target-dir]",
	Short: "Clone an existing dotfiles repository and link capabilities into local client harnesses",
	Long: `Clone an existing company or personal dotfiles repository into ~/.koharness/repo (or specified target path), inspect local client AI harnesses (~/.gemini, ~/.claude.json, ~/.codex), display an interactive TUI setup dashboard, back up conflicting files, and establish atomic symlinks.`,
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		targetPath := initPathOverride
		if targetPath == "" && len(args) > 1 && args[1] != "" {
			targetPath = args[1]
		}
		if targetPath == "" {
			targetPath, err = harness.GetRepoPath("")
			if err != nil {
				return fmt.Errorf("failed resolving target repository path: %w", err)
			}
		} else {
			targetPath = harness.ExpandTilde(targetPath, homeDir)
		}
		targetPath = filepath.Clean(targetPath)

		out := cmd.OutOrStdout()
		fmt.Fprintln(out, tui.BadgeInfo("CLONING REPOSITORY"), fmt.Sprintf("Cloning %s into %s...", repoURL, targetPath))

		if err := harness.CloneRepository(repoURL, targetPath, initForce, harness.ClonerOptions{Out: out}); err != nil {
			if err == harness.ErrTargetDirectoryExists {
				errOut := cmd.OutOrStderr()
				fmt.Fprintln(errOut, tui.BadgeError("REPOSITORY ALREADY EXISTS"), lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Target path %s exists and is non-empty.", targetPath)))
				fmt.Fprintln(errOut, tui.StyleMuted.Render("Run `koharness init <url> --force` to backup existing directory and re-initialize."))
				return fmt.Errorf("target repository directory exists: %s", targetPath)
			}
			return fmt.Errorf("failed cloning repository: %w", err)
		}

		items, err := harvester.ScanRepoCapabilities(targetPath, homeDir)
		if err != nil {
			return fmt.Errorf("failed scanning repository capabilities: %w", err)
		}

		if len(items) == 0 {
			fmt.Fprintln(out, tui.BadgeInfo("NO CAPABILITIES DISCOVERED"), "Repository cloned successfully, but no skills, prompts, or MCP servers found.")
		}

		selectedItems := items
		if !initNonInteractive && len(items) > 0 {
			viewItems, confirmed, err := tui.RunInitView(items)
			if err != nil {
				return fmt.Errorf("init TUI error: %w", err)
			}
			if !confirmed {
				fmt.Fprintln(out, tui.BadgeWarn("CANCELED"), "Repository setup canceled by user.")
				return nil
			}
			selectedItems = viewItems
		}

		detector, _ := harness.NewDetector(harness.WithHomeDir(homeDir))
		var origHarnesses []string
		if detector != nil {
			for _, adapter := range detector.DetectInstalled() {
				origHarnesses = append(origHarnesses, string(adapter.ID()))
			}
		}

		linker := symlink.NewLinkerEngine(symlink.LinkerConfig{HomeDir: homeDir})
		linkedCount := 0
		for _, item := range selectedItems {
			if !item.Selected {
				continue
			}
			_, err := linker.CreateSymlink(item.RepoPath, item.TargetPath)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), tui.BadgeError("SYMLINK ERROR"), fmt.Sprintf("Failed linking %s to %s: %v", item.RepoPath, item.TargetPath, err))
			} else {
				linkedCount++
			}
		}

		cfg, _ := harness.LoadGlobalConfig(harness.PathOptions{HomeDir: homeDir})
		if cfg == nil {
			cfg = &harness.GlobalConfig{}
		}
		cfg.RepoPath = targetPath
		if len(cfg.OriginalHarnesses) == 0 {
			cfg.OriginalHarnesses = origHarnesses
		}

		if err := harness.SaveGlobalConfig(cfg, harness.PathOptions{HomeDir: homeDir}); err != nil {
			fmt.Fprintln(cmd.OutOrStderr(), tui.BadgeWarn("CONFIG WARNING"), fmt.Sprintf("Failed writing global config: %v", err))
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, tui.BadgeSuccess("SUCCESS"), lipgloss.NewStyle().Bold(true).Render("Repository initialized and capabilities linked cleanly!"))
		fmt.Fprintln(out, fmt.Sprintf("  Repository Path: %s", targetPath))
		fmt.Fprintln(out, fmt.Sprintf("  Linked Assets:   %d capability symlink(s)", linkedCount))
		fmt.Fprintln(out)
		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("Next Steps:"))
		fmt.Fprintln(out, fmt.Sprintf("  1. cd %s", targetPath))
		fmt.Fprintln(out, "  2. koharness repo -e")
		fmt.Fprintln(out, "  3. Future unmanaged capability discovery: Run `koharness discover` to scan local client harnesses.")

		return nil
	},
}

func init() {
	InitCmd.Flags().StringVarP(&initPathOverride, "path", "d", "", "override repository directory target path")
	InitCmd.Flags().BoolVarP(&initForce, "force", "f", false, "force overwrite/backup of existing repository directory if present")
	InitCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "skip TUI checklist and link all capabilities automatically")

	RootCmd.AddCommand(InitCmd)
}
