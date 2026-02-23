package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stacked",
	Short: "Manage stacked pull requests",
	Long:  "A CLI tool for managing stacked pull requests — create, navigate, rebase, push, and sync chains of dependent branches.",
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
