package cmd

import (
	"fmt"

	"github.com/kerbaras/stacked/pkg/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize stacked in the current repository",
	RunE:  runInit,
}

var hooksFlag bool

func runInit(cmd *cobra.Command, args []string) error {
	repo, err := openRepo()
	if err != nil {
		return err
	}

	if hooksFlag {
		if repo.HasHooks() {
			ui.Info("hooks already installed")
			return nil
		}
		if err := repo.InstallHooks(); err != nil {
			return fmt.Errorf("install hooks: %w", err)
		}
		ui.Success("git hooks installed")
	}

	ui.Success("stacked initialized")
	return nil
}

func init() {
	initCmd.Flags().BoolVar(&hooksFlag, "hooks", false, "install git hooks")
	rootCmd.AddCommand(initCmd)
}
