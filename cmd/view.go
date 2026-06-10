package cmd

import (
	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/NathanielRand/mdboard/internal/tui"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Interactive TUI board viewer",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Resolve the path just like your other commands
		path, err := resolveBoardPath(cmd)
		if err != nil {
			return err
		}

		b, err := markdown.Parse(path)
		if err != nil {
			return err
		}

		// 2. Pass BOTH the board and the path to the TUI
		return tui.Run(b, path)
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
