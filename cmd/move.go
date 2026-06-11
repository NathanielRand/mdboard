package cmd

import (
	"fmt"

	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var (
	moveCol  string
	moveTask int
	moveFrom string
)

var moveCmd = &cobra.Command{
	Use:   "move [card title] [column]",
	Short: "Move a card to a different column",
	Args: func(cmd *cobra.Command, args []string) error {
		task, _ := cmd.Flags().GetInt("task")
		if task > 0 {
			from, _ := cmd.Flags().GetString("from")
			col, _ := cmd.Flags().GetString("col")
			if from == "" {
				return fmt.Errorf("--task/-t requires --from/-C to specify source column")
			}
			if col == "" && len(args) < 1 {
				return fmt.Errorf("requires target column as positional arg or --col/-c flag")
			}
			return nil
		}
		if len(args) < 1 {
			return fmt.Errorf("requires card title and column (or use --task/-t with --from/-C)")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolveBoardPath(cmd)
		if err != nil {
			return err
		}

		b, err := markdown.Parse(path)
		if err != nil {
			return err
		}

		var card *board.Card
		var fromCol *board.Column
		var fromIdx int
		var toColName string

		if moveTask > 0 {
			srcCol, err := board.FindColumn(b, moveFrom)
			if err != nil {
				return err
			}
			c, idx, err := board.FindCardByIndex(srcCol, moveTask)
			if err != nil {
				return err
			}
			card, fromCol, fromIdx = c, srcCol, idx
			if moveCol != "" {
				toColName = moveCol
			} else {
				toColName = joinArgs(args)
			}
		} else {
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
			card, fromCol, fromIdx, err = board.FindCard(b, cardTitle)
			if err != nil {
				return err
			}
			toColName = colName
		}

		toCol, err := board.FindColumn(b, toColName)
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
	moveCmd.Flags().IntVarP(&moveTask, "task", "t", 0, "1-based card index within the source column")
	moveCmd.Flags().StringVarP(&moveFrom, "from", "C", "", "Source column for index-based lookup (name, partial, or index)")
	rootCmd.AddCommand(moveCmd)
}
