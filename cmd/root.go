package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/koharness/koharness/pkg/tui"
	"github.com/koharness/koharness/pkg/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	showVer bool
)

// RootCmd represents the base command when called without any subcommands.
var RootCmd = &cobra.Command{
	Use:   "koharness",
	Short: "KoHarness - Cross-Harness AI Capabilities Manager",
	Long:  `koharness is a cross-harness CLI manager designed to centralize and synchronize AI capabilities across local developer environments.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVer {
			PrintVersion()
			return nil
		}
		return cmd.Help()
	},
}

// PrintVersion outputs the banner and detailed version/build metadata.
func PrintVersion() {
	fmt.Println(tui.RenderBanner())
	vInfo := version.Get()
	fmt.Printf("Version:   %s\n", vInfo.Version)
	if vInfo.Commit != "" {
		fmt.Printf("Commit:    %s\n", vInfo.Commit)
	}
	if vInfo.Date != "" {
		fmt.Printf("BuildDate: %s\n", vInfo.Date)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose debug logging")
	RootCmd.Flags().BoolVarP(&showVer, "version", "V", false, "display version and build metadata")

	RootCmd.SetHelpTemplate(getCustomHelpTemplate())
}

func getCustomHelpTemplate() string {
	header := tui.RenderBanner()
	return header + "\n" +
		lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("USAGE:") + "\n" +
		"  {{.UseLine}}\n\n" +
		lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("DESCRIPTION:") + "\n" +
		"  {{.Long}}\n" +
		"{{if .HasAvailableSubCommands}}\n" +
		lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("AVAILABLE COMMANDS:") + "\n" +
		"{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name \"help\"))}}  " +
		lipgloss.NewStyle().Foreground(tui.ColorCyan).Render("{{rpad .Name .NamePadding}}") +
		" {{.Short}}\n" +
		"{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}\n" +
		lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("FLAGS:") + "\n" +
		"{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}\n" +
		"{{end}}{{if .HasAvailableInheritedFlags}}\n" +
		lipgloss.NewStyle().Bold(true).Foreground(tui.ColorViolet).Render("GLOBAL FLAGS:") + "\n" +
		"{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}\n" +
		"{{end}}\n"
}
