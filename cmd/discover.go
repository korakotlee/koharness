// Package cmd implements CLI commands for the KoHarness application.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/harvester"
	"github.com/korakotlee/koharness/pkg/tui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	discoverNonInteractive bool
)

// discoverCmd represents the `koharness discover [repo-path]` subcommand.
var discoverCmd = &cobra.Command{
	Use:   "discover [repo-path]",
	Short: "Discover unmanaged local capabilities and import or ignore them",
	Long:  `Inspect local AI client harnesses (~/.gemini, ~/.claude.json, ~/.codex) for new unmanaged capabilities, present an interactive 3-state checklist (Import, Skip, Ignore), persist ignored items into .koharness.local.yaml, import selected capabilities, and display PR instructions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		repoPath := filepath.Join(homeDir, ".koharness", "repo")
		if len(args) > 0 && args[0] != "" {
			repoPath = args[0]
		}

		fs := afero.NewOsFs()

		// Validate repository exists
		exists, err := afero.Exists(fs, repoPath)
		if err != nil || !exists {
			fmt.Println(tui.BadgeError("REPOSITORY NOT FOUND"), lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Target repository path %s does not exist.", repoPath)))
			fmt.Println(tui.StyleMuted.Render("Please initialize a repository first using `koharness create [repo-path]`."))
			return fmt.Errorf("target repository path %s does not exist", repoPath)
		}

		scanner, err := harvester.NewScanner(
			harvester.WithFs(fs),
			harvester.WithHomeDir(homeDir),
		)
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}

		caps, err := scanner.ScanForRepo(repoPath)
		if err != nil {
			return fmt.Errorf("failed scanning capabilities for repository: %w", err)
		}

		if len(caps) == 0 {
			fmt.Println(tui.BadgeInfo("NO UNMANAGED CAPABILITIES"), "All local capabilities are already tracked or listed in .koharness.local.yaml.")
			return nil
		}

		processedCaps := caps
		if !discoverNonInteractive {
			items, confirmed, err := tui.RunHarvestView(caps)
			if err != nil {
				return fmt.Errorf("harvest TUI error: %w", err)
			}
			if !confirmed {
				fmt.Println(tui.BadgeWarn("CANCELED"), "Capability discovery canceled by user.")
				return nil
			}
			processedCaps = items
		}

		var importItems []harvester.DiscoveredCapability
		var ignoreItems []harvester.DiscoveredCapability

		for _, item := range processedCaps {
			switch item.GetState() {
			case harvester.StateImport:
				importItems = append(importItems, item)
			case harvester.StateIgnore:
				ignoreItems = append(ignoreItems, item)
			}
		}

		// Handle Ignored Items
		if len(ignoreItems) > 0 {
			localCfg, err := harvester.LoadLocalConfig(fs, repoPath)
			if err != nil {
				return fmt.Errorf("failed loading local config: %w", err)
			}
			for _, item := range ignoreItems {
				localCfg.AddIgnore(item.Type, item.Name, item.HarnessID)
			}
			if err := harvester.SaveLocalConfig(fs, repoPath, localCfg); err != nil {
				return fmt.Errorf("failed saving local ignore config: %w", err)
			}
		}

		// Handle Imported Items
		if len(importItems) > 0 {
			creator := harvester.NewCreator(harvester.CreatorOptions{
				Fs:       fs,
				HomeDir:  homeDir,
				RepoPath: repoPath,
				InitGit:  false,
			})
			if err := creator.HarvestCapabilities(importItems); err != nil {
				return fmt.Errorf("failed harvesting capabilities: %w", err)
			}
		}

		// Print Summary and PR instructions
		fmt.Println(tui.BadgeSuccess("SUCCESS"), lipgloss.NewStyle().Bold(true).Render("Discovered capabilities processed!"))
		fmt.Printf("  - Imported & Linked: %d\n", len(importItems))
		fmt.Printf("  - Ignored Locally:   %d (saved to .koharness.local.yaml)\n\n", len(ignoreItems))

		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("Next Steps:"))
		fmt.Printf("  1. cd %s\n", repoPath)
		fmt.Println("  2. git status")
		fmt.Println("  3. Commit new capabilities and submit a pull request")

		return nil
	},
}

func init() {
	discoverCmd.Flags().BoolVar(&discoverNonInteractive, "non-interactive", false, "skip TUI checklist and automatically import non-ignored capabilities")
	RootCmd.AddCommand(discoverCmd)
}
