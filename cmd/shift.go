package cmd

import (
	"fmt"
	"strings"

	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var shiftCmd = &cobra.Command{
	Use:   "shift [card title] [up|down]",
	Short: "Shift a card up or down within its column",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolveBoardPath(cmd)
		if err != nil {
			return err
		}

		b, err := markdown.Parse(path)
		if err != nil {
			return err
		}

		// Last arg is the direction, rest is the card title
		direction := strings.ToLower(args[len(args)-1])
		if direction != "up" && direction != "down" {
			return fmt.Errorf("direction must be 'up' or 'down', got %q", direction)
		}

		cardTitle := joinArgs(args[:len(args)-1])
		isUp := direction == "up"

		card, col, idx, err := board.FindCard(b, cardTitle)
		if err != nil {
			return err
		}

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
	rootCmd.AddCommand(shiftCmd)
}
