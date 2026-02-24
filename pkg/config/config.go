// Package config defines application configuration types and defaults.
package config

const DefaultBranchTemplate = "stack/%name/%02d-%slug"

// Config holds application-wide configuration.
type Config struct {
	BranchTemplate string `mapstructure:"branch_template"`
}
