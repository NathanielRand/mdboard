package cmd

import (
	"fmt"

	"github.com/nrand/mdboard/internal/board"
	"github.com/nrand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [card title]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a card from the board",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolveBoardPath(cmd)
		if err != nil {
			return err
		}

		b, err := markdown.Parse(path)
		if err != nil {
			return err
		}

		title := joinArgs(args)
		card, err := board.RemoveCard(b, title)
		if err != nil {
			return err
		}

		if err := markdown.Write(path, b); err != nil {
			return err
		}

		fmt.Printf("🗑️  Removed \"%s\"\n", card.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
