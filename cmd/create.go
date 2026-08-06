// Package cmd implements CLI commands for the KoHarness application.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/harvester"
	"github.com/korakotlee/koharness/pkg/tui"
	"github.com/spf13/cobra"
)

var (
	nonInteractive bool
	gitInit        bool
)

// createCmd represents the `koharness create [repo-path]` subcommand.
var createCmd = &cobra.Command{
	Use:   "create [repo-path]",
	Short: "Harvest local AI capabilities and initialize a new dotfiles repository",
	Long:  `Scan local client harnesses (~/.gemini, ~/.claude.json, ~/.codex), display an interactive TUI to pick capabilities, bootstrap a new dotfiles repo, back up original assets, and establish symlinks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		repoPath := filepath.Join(homeDir, ".koharness", "repo")
		if len(args) > 0 && args[0] != "" {
			repoPath = args[0]
		}

		// Pre-check: Abort if target repository path already exists and is non-empty
		if info, err := os.Stat(repoPath); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(repoPath)
			if len(entries) > 0 {
				fmt.Println(tui.BadgeError("REPOSITORY ALREADY EXISTS"), lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Target repository path %s already exists.", repoPath)))
				fmt.Println(tui.StyleMuted.Render("To manage an existing repository, run `koharness sync` or specify a new target path: `koharness create [repo-path]`."))
				return nil
			}
		}

		scanner, err := harvester.NewScanner(harvester.WithHomeDir(homeDir))
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}

		caps, err := scanner.ScanAll()
		if err != nil {
			return fmt.Errorf("failed scanning capabilities: %w", err)
		}

		if len(caps) == 0 {
			fmt.Println(tui.BadgeInfo("NO CAPABILITIES FOUND"), "No unmanaged skills, prompts, or MCP servers discovered.")
		}

		selectedCaps := caps
		if !nonInteractive && len(caps) > 0 {
			items, confirmed, err := tui.RunHarvestView(caps)
			if err != nil {
				return fmt.Errorf("harvest TUI error: %w", err)
			}
			if !confirmed {
				fmt.Println(tui.BadgeWarn("CANCELED"), "Repository creation canceled by user.")
				return nil
			}
			selectedCaps = items
		}

		creator := harvester.NewCreator(harvester.CreatorOptions{
			HomeDir:  homeDir,
			RepoPath: repoPath,
			InitGit:  gitInit,
		})

		detector, _ := harness.NewDetector(harness.WithHomeDir(homeDir))
		var origHarnesses []string
		if detector != nil {
			for _, adapter := range detector.DetectInstalled() {
				origHarnesses = append(origHarnesses, string(adapter.ID()))
			}
		}

		if err := creator.HarvestCapabilities(selectedCaps); err != nil {
			return fmt.Errorf("failed harvesting capabilities: %w", err)
		}

		cfg, _ := harness.LoadGlobalConfig(harness.PathOptions{HomeDir: homeDir})
		if cfg == nil {
			cfg = &harness.GlobalConfig{}
		}
		cfg.RepoPath = repoPath
		if len(cfg.OriginalHarnesses) == 0 {
			cfg.OriginalHarnesses = origHarnesses
		}

		_ = harness.SaveGlobalConfig(cfg, harness.PathOptions{HomeDir: homeDir})

		fmt.Println(tui.BadgeSuccess("SUCCESS"), lipgloss.NewStyle().Bold(true).Render("Repository initialized and capabilities linked cleanly!"))
		fmt.Printf("Repository Path: %s\n", repoPath)
		return nil
	},
}

func init() {
	createCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "skip TUI checklist and harvest all discovered capabilities")
	createCmd.Flags().BoolVar(&gitInit, "git", true, "automatically run git init in the target repository")
	RootCmd.AddCommand(createCmd)
}
