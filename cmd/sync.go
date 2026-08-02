package cmd

import (
	"context"
	"fmt"

	"github.com/korakotlee/koharness/pkg/sync"
	"github.com/spf13/cobra"
)

var (
	syncPathOverride   string
	syncNonInteractive bool
)

// SyncCmd represents the `koharness sync` subcommand.
var SyncCmd = &cobra.Command{
	Use:   "sync [repo-path]",
	Short: "Pull remote capability updates, check for dirty local state, and refresh client AI harness links",
	Long: `Sync upstream capability updates into ~/.koharness/repo (or specified repository path). Inspects Git status to prevent overwriting uncommitted local changes, executes git pull --rebase on clean repositories, and re-triggers symlink and MCP config merging across client AI harnesses.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := syncPathOverride
		if targetPath == "" && len(args) > 0 {
			targetPath = args[0]
		}

		out := cmd.OutOrStdout()
		errOut := cmd.OutOrStderr()

		opts := sync.SyncOptions{
			RepoPath:       targetPath,
			NonInteractive: syncNonInteractive,
			Out:            out,
			ErrOut:         errOut,
		}

		engine, err := sync.NewSyncEngine(opts)
		if err != nil {
			return fmt.Errorf("failed initializing sync engine: %w", err)
		}

		ctx := context.Background()
		result, err := engine.Run(ctx)
		if err != nil {
			if err == sync.ErrDirtyRepo {
				return err
			}
			return fmt.Errorf("sync failed: %w", err)
		}

		sync.PrintSyncSuccess(out, result)
		return nil
	},
}

func init() {
	SyncCmd.Flags().StringVarP(&syncPathOverride, "path", "d", "", "override repository path to synchronize")
	SyncCmd.Flags().BoolVar(&syncNonInteractive, "non-interactive", false, "skip interactive TUI warning prompt if uncommitted changes exist")

	RootCmd.AddCommand(SyncCmd)
}
