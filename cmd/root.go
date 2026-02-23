package cmd

import (
	"os"
	"path/filepath"

	"github.com/kerbaras/stacked/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	cfg     config.Config
)

var rootCmd = &cobra.Command{
	Use:          "stacked",
	Short:        "Manage stacked pull requests",
	Long:         "A CLI tool for managing stacked pull requests — create, navigate, rebase, push, and sync chains of dependent branches.",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return loadConfig()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadConfig() error {
	viper.SetDefault("branch_template", config.DefaultBranchTemplate)

	// 1. Global config: $HOME/.config/stacked.yaml
	if home, err := os.UserHomeDir(); err == nil {
		viper.SetConfigFile(filepath.Join(home, ".config", "stacked.yaml"))
		_ = viper.MergeInConfig()
	}

	// 2. Explicit config file from --config flag
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		_ = viper.MergeInConfig()
	}

	// 3. Environment variables with STACKED_ prefix
	viper.SetEnvPrefix("STACKED")
	viper.AutomaticEnv()
	_ = viper.BindEnv("branch_template")
	_ = viper.BindEnv("github_token")

	return viper.Unmarshal(&cfg)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", ".stacked.yaml", "config file path")
}
