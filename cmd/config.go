package cmd

import (
	"fmt"

	"github.com/nrand/mdboard/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or set mdboard configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print current config",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Printf("Config file: %s\n\n", config.Path())
		fmt.Printf("github_user:      %s\n", cfg.GitHubUser)
		fmt.Printf("default_columns:\n")
		for _, c := range cfg.DefaultColumns {
			fmt.Printf("  - %s\n", c)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a config value (e.g. github_user nrand)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		key, val := args[0], args[1]
		switch key {
		case "github_user":
			cfg.GitHubUser = val
		default:
			return fmt.Errorf("unknown config key %q (available: github_user)", key)
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("✅ Set %s = %s\n", key, val)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
