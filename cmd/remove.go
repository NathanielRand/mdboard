package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/markdown"
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

		// Find the card first so we can show the exact title in the prompt
		card, col, idx, err := board.FindCard(b, title)
		if err != nil {
			return err
		}

		fmt.Printf("Remove \"%s\" from [%s]? (y/N) ", card.Title, col.Name)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}

		col.Cards = append(col.Cards[:idx], col.Cards[idx+1:]...)

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
