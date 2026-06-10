package cmd

import (
	"fmt"

	"github.com/nrand/mdboard/internal/board"
	"github.com/nrand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var addCol string

var addCmd = &cobra.Command{
	Use:   "add [title]",
	Short: "Add a new card to the board",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolveBoardPath(cmd)
		if err != nil {
			return err
		}

		b, err := markdown.Parse(path)
		if err != nil {
			return err
		}

		colName := addCol
		if colName == "" {
			// Default to first column
			if len(b.Columns) == 0 {
				return fmt.Errorf("board has no columns")
			}
			colName = b.Columns[0].Name
		}

		title := joinArgs(args)
		card, col, err := board.AddCard(b, title, colName)
		if err != nil {
			return err
		}

		if err := markdown.Write(path, b); err != nil {
			return err
		}

		fmt.Printf("➕ Added card \"%s\" → [%s]\n", card.Title, col.Name)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addCol, "col", "c", "", "Column to add the card to (default: first column)")
	rootCmd.AddCommand(addCmd)
}
