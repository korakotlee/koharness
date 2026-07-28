// Package cmd implements CLI subcommands for the KoHarness application.
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/korakotlee/koharness/pkg/harness"
	"github.com/korakotlee/koharness/pkg/tui"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// RepoCmd represents the `koharness repo` subcommand to navigate or operate in the dotfiles repository.
var RepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Cd to or launch a subshell inside the local dotfiles repository directory",
	Long: `Navigate into or launch an interactive subshell inside your central dotfiles repository (~/.koharness/repo or configured KOHARNESS_REPO path).

Supports subshell spawning, raw path printing (--print-path / -p) for shell alias integration (e.g. cd "$(koharness repo -p)"), code editor launching (--editor / -e), and shell helper generators (--shell-init).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		shellInit, _ := cmd.Flags().GetBool("shell-init")
		if shellInit {
			PrintShellInit(cmd.OutOrStdout())
			return nil
		}

		customPath, _ := cmd.Flags().GetString("path")
		repoPath, err := harness.GetRepoPath(customPath)
		if err != nil {
			return fmt.Errorf("failed resolving repository path: %w", err)
		}

		// Verify target repository directory exists
		info, err := os.Stat(repoPath)
		if err != nil || !info.IsDir() {
			out := cmd.OutOrStderr()
			fmt.Fprintln(out, tui.BadgeError("REPOSITORY NOT FOUND"), lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Repository path %s does not exist.", repoPath)))
			fmt.Fprintln(out, tui.StyleMuted.Render("Run `koharness init` to bootstrap your repository or specify a valid path with `koharness repo -d [path]`."))
			return fmt.Errorf("repository directory not found: %s", repoPath)
		}

		printPath, _ := cmd.Flags().GetBool("print-path")
		isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
		if printPath || !isTerminal {
			fmt.Fprintln(cmd.OutOrStdout(), repoPath)
			return nil
		}

		// Launch in code editor if requested
		if cmd.Flags().Changed("editor") {
			editorBinary, _ := cmd.Flags().GetString("editor")
			if editorBinary == "" {
				editorBinary = os.Getenv("EDITOR")
			}
			if editorBinary == "" {
				editorBinary = "code"
			}

			fmt.Fprintln(cmd.OutOrStdout(), tui.BadgeInfo("LAUNCHING EDITOR"), fmt.Sprintf("Opening %s in %s...", repoPath, editorBinary))
			editorCmd := exec.Command(editorBinary, repoPath)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = cmd.OutOrStdout()
			editorCmd.Stderr = cmd.OutOrStderr()
			return editorCmd.Run()
		}

		// Default: Spawn interactive subshell inside repoPath
		shellPath := os.Getenv("SHELL")
		if shellPath == "" {
			if runtime.GOOS == "windows" {
				shellPath = "cmd.exe"
			} else {
				shellPath = "/bin/zsh"
			}
		}

		fmt.Fprintln(cmd.OutOrStdout(), tui.BadgeSuccess("ENTERED REPOSITORY"), lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Interactive subshell launched in %s", repoPath)))
		fmt.Fprintln(cmd.OutOrStdout(), tui.StyleMuted.Render("Type 'exit' to return to your previous directory.\n"))

		subshell := exec.Command(shellPath)
		subshell.Dir = repoPath
		subshell.Stdin = os.Stdin
		subshell.Stdout = cmd.OutOrStdout()
		subshell.Stderr = cmd.OutOrStderr()
		return subshell.Run()
	},
}

// PrintShellInit outputs shell functions for Zsh, Bash, and Fish shells.
func PrintShellInit(w io.Writer) {
	fmt.Fprintln(w, "# KoHarness Shell Integration")
	fmt.Fprintln(w, "# Add one of the following snippets to your shell configuration file:")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "# Zsh / Bash (~/.zshrc or ~/.bashrc):")
	fmt.Fprintln(w, `khcd() { cd "$(koharness repo -p "$@")"; }`)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "# Fish (~/.config/fish/config.fish):")
	fmt.Fprintln(w, `function khcd; cd (koharness repo -p $argv); end`)
}

func init() {
	RepoCmd.Flags().StringP("path", "d", "", "override repository directory path")
	RepoCmd.Flags().BoolP("print-path", "p", false, "output raw repository directory path to stdout")
	RepoCmd.Flags().StringP("editor", "e", "", "open repository directory in specified code editor (defaults to $EDITOR or 'code')")
	RepoCmd.Flags().Lookup("editor").NoOptDefVal = ""
	RepoCmd.Flags().Bool("shell-init", false, "output shell integration snippet for cd functions")

	RootCmd.AddCommand(RepoCmd)
}
