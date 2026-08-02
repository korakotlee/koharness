package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/lint"
	"github.com/korakotlee/koharness/pkg/tui"
	"github.com/spf13/cobra"
)

var lintPathOverride string

// LintCmd represents the `koharness lint` subcommand.
var LintCmd = &cobra.Command{
	Use:   "lint [repo-path]",
	Short: "Validate repository assets (JSON/YAML syntax, script permissions, skill metadata)",
	Long: `Lint validates repository assets for quality assurance. It recursively checks JSON/YAML files in mcp/ and harnesses/ for syntax errors, verifies executable permissions (+x) on script files in skills/*/scripts/, and validates YAML frontmatter metadata in SKILL.md files. Exits with code 0 if all checks pass, or non-zero if issues are found.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := lintPathOverride
		if targetPath == "" && len(args) > 0 {
			targetPath = args[0]
		}

		out := cmd.OutOrStdout()
		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("Running koharness lint quality checks..."))

		opts := lint.LintOptions{
			RepoRoot: targetPath,
		}

		result, err := lint.Run(opts)
		if err != nil {
			return fmt.Errorf("lint check failed to execute: %w", err)
		}

		if result.HasErrors() {
			fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorRed).Render(fmt.Sprintf("\nFound %d lint issue(s):", len(result.Issues))))
			for _, issue := range result.Issues {
				catStyle := lipgloss.NewStyle().Foreground(tui.ColorCyan).Render(fmt.Sprintf("[%s]", issue.Category))
				pathStyle := lipgloss.NewStyle().Bold(true).Render(issue.Path)
				fmt.Fprintf(out, "  %s %s: %s\n", catStyle, pathStyle, issue.Message)
			}
			return fmt.Errorf("lint checks failed with %d issue(s)", len(result.Issues))
		}

		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorGreen).Render("\n✓ All lint checks passed successfully!"))
		return nil
	},
}

func init() {
	LintCmd.Flags().StringVarP(&lintPathOverride, "path", "d", "", "override repository path to lint")
	RootCmd.AddCommand(LintCmd)
}
