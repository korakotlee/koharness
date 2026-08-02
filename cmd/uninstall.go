package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/symlink"
	"github.com/korakotlee/koharness/pkg/tui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	uninstallDryRun      bool
	uninstallForce       bool
	uninstallPath        string
	uninstallPurgeConfig bool
)

// uninstallCmd represents the `koharness uninstall` subcommand.
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall koharness, restore symlinks to physical files, and remove target repository",
	Long:  `Inspect local AI client harnesses (~/.gemini, ~/.claude, ~/.codex), detect symlinks pointing to the managed koharness repository, atomically convert them back to standalone physical files and directories, and remove the target repository clone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		fs := afero.NewOsFs()
		repoPath, err := harness.GetRepoPath(uninstallPath, harness.PathOptions{Fs: fs, HomeDir: homeDir})
		if err != nil {
			return fmt.Errorf("failed to resolve repository path: %w", err)
		}

		engine, err := symlink.NewUninstallEngine(symlink.UninstallConfig{
			Fs:          fs,
			HomeDir:     homeDir,
			RepoPath:    repoPath,
			DryRun:      uninstallDryRun,
			PurgeConfig: uninstallPurgeConfig,
		})
		if err != nil {
			return fmt.Errorf("failed creating uninstall engine: %w", err)
		}

		discovered, err := engine.DiscoverSymlinks()
		if err != nil {
			return fmt.Errorf("failed discovering symlinks: %w", err)
		}

		if !uninstallForce && !uninstallDryRun {
			fmt.Fprintln(out, tui.BadgeWarn("WARNING"), lipgloss.NewStyle().Bold(true).Render("This will replace all symlinks pointing into your target repository with standalone physical files and remove the repository directory."))
			fmt.Fprintf(out, "Repository Path: %s\n", repoPath)
			fmt.Fprintf(out, "Symlinks to restore: %d\n", len(discovered))
			fmt.Fprint(out, "Are you sure you want to proceed? [y/N]: ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Fprintln(out, tui.BadgeWarn("CANCELED"), "Uninstallation canceled.")
				return nil
			}
		}

		result, err := engine.Execute(discovered)
		if err != nil {
			return fmt.Errorf("uninstallation failed: %w", err)
		}

		if uninstallDryRun {
			fmt.Fprintln(out, tui.BadgeInfo("DRY RUN"), lipgloss.NewStyle().Bold(true).Render("Preview of uninstallation actions:"))
		} else {
			fmt.Fprintln(out, tui.BadgeSuccess("SUCCESS"), lipgloss.NewStyle().Bold(true).Render("Uninstallation execution completed."))
		}

		for _, asset := range result.RestoredAssets {
			assetKind := "file"
			if asset.IsDir {
				assetKind = "directory"
			}
			fmt.Fprintf(out, "  ✔ Restored %s (converted symlink -> %s)\n", asset.TargetPath, assetKind)
		}

		fmt.Fprintf(out, "\n  ✔ Replaced %d symlink(s) with standalone physical assets.\n", len(result.RestoredAssets))
		if result.RepoRemoved {
			fmt.Fprintf(out, "  ✔ Removed repository directory: %s\n", repoPath)
		}
		if result.ConfigPurged {
			fmt.Fprintln(out, "  ✔ Purged global configuration directory.")
		}

		fmt.Fprintln(out, "\n"+tui.BadgeSuccess("COMPLETED")+" Koharness uninstalled. Dotfiles successfully restored to standalone files.")
		return nil
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallDryRun, "dry-run", false, "display all symlinks that will be converted and paths to be removed without altering the filesystem")
	uninstallCmd.Flags().BoolVarP(&uninstallForce, "force", "f", false, "bypass interactive confirmation prompt")
	uninstallCmd.Flags().StringVarP(&uninstallPath, "path", "d", "", "override target repository path")
	uninstallCmd.Flags().BoolVar(&uninstallPurgeConfig, "purge-config", false, "purge global ~/.koharness directory after repository removal")

	RootCmd.AddCommand(uninstallCmd)
}
