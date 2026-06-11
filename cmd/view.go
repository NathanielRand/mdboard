package cmd

import (
	"os"

	"github.com/NathanielRand/mdboard/internal/config"
	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/NathanielRand/mdboard/internal/tui"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Interactive TUI board viewer",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolveBoardPath(cmd)
		if err != nil {
			return err
		}

		b, err := markdown.Parse(path)
		if err != nil {
			return err
		}

		primaryColor := ""
		gitUser := ""
		cwd, _ := os.Getwd()
		if pc, err := config.LoadProjectAt(cwd); err == nil {
			primaryColor = pc.PrimaryColor
			gitUser = pc.GitUser
		}

		return tui.Run(b, path, Version, primaryColor, gitUser)
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
