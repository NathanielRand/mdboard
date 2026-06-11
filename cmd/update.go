package cmd

import (
	"fmt"

	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var (
	updateTitle string
	updateBody  string
	updateTask  int
	updateCol   string
)

var updateCmd = &cobra.Command{
	Use:   "update [card title]",
	Short: "Edit a card's title or body content",
	Args: func(cmd *cobra.Command, args []string) error {
		task, _ := cmd.Flags().GetInt("task")
		if task == 0 && len(args) < 1 {
			return fmt.Errorf("requires card title or --task/-t flag")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
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

		card, col, _, err := resolveCard(b, updateTask, updateCol, args)
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
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title for the card")
	updateCmd.Flags().StringVarP(&updateBody, "body", "b", "", "New body/content for the card")
	updateCmd.Flags().IntVarP(&updateTask, "task", "t", 0, "1-based card index within the column")
	updateCmd.Flags().StringVarP(&updateCol, "col", "c", "", "Column for index-based lookup (name, partial, or index)")
	rootCmd.AddCommand(updateCmd)
}
