package cmd

import (
	"fmt"
	"strings"

	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var (
	shiftTask int
	shiftCol  string
)

var shiftCmd = &cobra.Command{
	Use:   "shift [card title] [up|down]",
	Short: "Shift a card up or down within its column",
	Args: func(cmd *cobra.Command, args []string) error {
		task, _ := cmd.Flags().GetInt("task")
		if task > 0 {
			if len(args) < 1 {
				return fmt.Errorf("requires direction: up or down")
			}
			return nil
		}
		if len(args) < 2 {
			return fmt.Errorf("requires card title and direction (up|down), or use --task/-t with direction")
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

		var direction string
		var card *board.Card
		var col *board.Column
		var idx int

		if shiftTask > 0 {
			direction = strings.ToLower(args[0])
			c, cl, i, err := resolveCard(b, shiftTask, shiftCol, nil)
			if err != nil {
				return err
			}
			card, col, idx = c, cl, i
		} else {
			direction = strings.ToLower(args[len(args)-1])
			cardTitle := joinArgs(args[:len(args)-1])
			c, cl, i, err := board.FindCard(b, cardTitle)
			if err != nil {
				return err
			}
			card, col, idx = c, cl, i
		}

		if direction != "up" && direction != "down" {
			return fmt.Errorf("direction must be 'up' or 'down', got %q", direction)
		}
		isUp := direction == "up"

		if err := board.ShiftCard(col, idx, isUp); err != nil {
			// This catches if they try to move it out of bounds
			return fmt.Errorf("⚠️  Could not shift \"%s\": %v", card.Title, err)
		}

		if err := markdown.Write(path, b); err != nil {
			return err
		}

		fmt.Printf("↕️  Shifted \"%s\" %s in [%s]\n", card.Title, direction, col.Name)
		return nil
	},
}

func init() {
	shiftCmd.Flags().IntVarP(&shiftTask, "task", "t", 0, "1-based card index within the column")
	shiftCmd.Flags().StringVarP(&shiftCol, "col", "c", "", "Column for index-based lookup (name, partial, or index)")
	rootCmd.AddCommand(shiftCmd)
}
