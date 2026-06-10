package cmd

import (
	"fmt"

	"github.com/nrand/mdboard/internal/board"
	"github.com/nrand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var (
	updateTitle string
	updateBody  string
)

var updateCmd = &cobra.Command{
	Use:   "update [card title]",
	Short: "Edit a card's title or body content",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If neither flag is set, there's nothing to do
		if updateTitle == "" && updateBody == "" {
			return fmt.Errorf("must specify --title or --body to update")
		}

		path, err := resolveBoardPath(cmd)
		if err != nil {
			return err
		}

		b, err := markdown.Parse(path)
		if err != nil {
			return err
		}

		cardTitle := joinArgs(args)
		card, col, _, err := board.FindCard(b, cardTitle)
		if err != nil {
			return err
		}

		// Store old title for the success message before updating
		oldTitle := card.Title

		// Apply the updates using the function we defined earlier
		board.UpdateCard(card, updateTitle, updateBody)

		if err := markdown.Write(path, b); err != nil {
			return err
		}

		if updateTitle != "" {
			fmt.Printf("✏️  Updated card title: \"%s\" → \"%s\"\n", oldTitle, card.Title)
		} else {
			fmt.Printf("✏️  Updated body for card \"%s\" in [%s]\n", card.Title, col.Name)
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title for the card")
	updateCmd.Flags().StringVarP(&updateBody, "body", "b", "", "New body/content for the card")
	rootCmd.AddCommand(updateCmd)
}
