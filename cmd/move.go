package cmd

import (
	"fmt"

	"github.com/nrand/mdboard/internal/board"
	"github.com/nrand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var moveCol string

var moveCmd = &cobra.Command{
	Use:   "move [card title] [column]",
	Short: "Move a card to a different column",
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

		var colName, cardTitle string
		if moveCol != "" {
			colName = moveCol
			cardTitle = joinArgs(args)
		} else {
			if len(args) < 2 {
				return fmt.Errorf("requires card title and column (or use -c/--col flag)")
			}
			colName = args[len(args)-1]
			cardTitle = joinArgs(args[:len(args)-1])
		}

		card, fromCol, fromIdx, err := board.FindCard(b, cardTitle)
		if err != nil {
			return err
		}

		toCol, err := board.FindColumn(b, colName)
		if err != nil {
			return err
		}

		if fromCol.Name == toCol.Name {
			fmt.Printf("⚠️  Card \"%s\" is already in [%s]\n", card.Title, toCol.Name)
			return nil
		}

		board.MoveCard(b, card, fromCol, fromIdx, toCol)

		if err := markdown.Write(path, b); err != nil {
			return err
		}

		fmt.Printf("🚀 Moved \"%s\"\n   [%s] → [%s]\n", card.Title, fromCol.Name, toCol.Name)
		return nil
	},
}

func init() {
	moveCmd.Flags().StringVarP(&moveCol, "col", "c", "", "Target column (name, partial match, or 1-based index)")
	rootCmd.AddCommand(moveCmd)
}
