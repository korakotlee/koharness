package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/doctor"
	"github.com/korakotlee/koharness/pkg/tui"
	"github.com/spf13/cobra"
)

// DoctorCmd represents the `koharness doctor` subcommand.
var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Inspect developer environment health (symlinks, client harness configs)",
	Long: `Doctor checks developer environment health and diagnostic status. It verifies active symlinks and detects broken or dangling link targets, checks client AI harness configuration paths across Claude Code (~/.claude.json), Antigravity (~/.gemini), and Codex (~/.codex), and reports accessibility status.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("Running koharness doctor diagnostics...\n"))

		opts := doctor.DoctorOptions{}
		result, err := doctor.Run(opts)
		if err != nil {
			return fmt.Errorf("doctor diagnostic failed to execute: %w", err)
		}

		// 1. Client AI Harness Status
		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorCyan).Render("Client AI Harnesses:"))
		for _, status := range result.HarnessStatuses {
			var badge string
			if status.Installed && status.ConfigDirAccessible {
				badge = lipgloss.NewStyle().Foreground(tui.ColorGreen).Bold(true).Render("[ INSTALLED ]")
			} else if status.Installed {
				badge = lipgloss.NewStyle().Foreground(tui.ColorAmber).Bold(true).Render("[ DEGRADED  ]")
			} else {
				badge = lipgloss.NewStyle().Foreground(tui.ColorMuted).Render("[ NOT FOUND ]")
			}

			fmt.Fprintf(out, "  %-14s %s  %s\n", status.Name, badge, status.StatusMessage)
		}

		// 2. Symlink Health Status
		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorCyan).Render("\nActive Symlink Integrity:"))
		if len(result.SymlinkDiagnostics) == 0 {
			fmt.Fprintln(out, "  No active symlinks found in default harness directories.")
		} else {
			for _, diag := range result.SymlinkDiagnostics {
				if diag.IsBroken {
					badge := lipgloss.NewStyle().Foreground(tui.ColorRed).Bold(true).Render("[ BROKEN ]")
					fmt.Fprintf(out, "  %s %s -> %s (target missing)\n", badge, diag.LinkPath, diag.TargetPath)
				} else {
					badge := lipgloss.NewStyle().Foreground(tui.ColorGreen).Render("[   OK   ]")
					fmt.Fprintf(out, "  %s %s -> %s\n", badge, diag.LinkPath, diag.TargetPath)
				}
			}
		}

		if result.HasBrokenSymlinks() {
			fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorRed).Render(fmt.Sprintf("\n✘ Doctor found %d broken symlink(s) on your system.", result.BrokenSymlinkCount())))
			return fmt.Errorf("doctor detected %d broken symlink(s)", result.BrokenSymlinkCount())
		}

		fmt.Fprintln(out, lipgloss.NewStyle().Bold(true).Foreground(tui.ColorGreen).Render("\n✓ Developer environment health check complete!"))
		return nil
	},
}

func init() {
	RootCmd.AddCommand(DoctorCmd)
}
