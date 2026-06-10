package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/NathanielRand/mdboard/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View or set mdboard configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print current project config",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		var cfg *config.Config
		var cfgDir string
		if projectDir, pc, _ := config.LoadProject(cwd); pc != nil {
			cfg = pc
			cfgDir = projectDir
		} else {
			cfg = &config.Config{}
			cfgDir = cwd
		}
		fmt.Printf("Config file: %s\n\n", config.ProjectConfigPath(cfgDir))
		fmt.Printf("board:            %s\n", cfg.Board)
		fmt.Printf("git_user:      %s\n", cfg.GitUser)
		fmt.Printf("default_columns:\n")
		cols := cfg.DefaultColumns
		if len(cols) == 0 {
			cols = config.Default().DefaultColumns
		}
		for _, c := range cols {
			fmt.Printf("  - %s\n", c)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a config value in the project config (e.g. git_user yourname)",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		key := args[0]
		val := strings.Join(args[1:], " ")

		updates := &config.Config{}
		switch key {
		case "git_user":
			updates.GitUser = val
		case "default_columns":
			return fmt.Errorf("use 'config set default_columns' is not supported via CLI — edit .mdboard/config.yaml directly")
		default:
			return fmt.Errorf("unknown config key %q (available: git_user)", key)
		}

		if err := config.SaveProject(cwd, updates); err != nil {
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
