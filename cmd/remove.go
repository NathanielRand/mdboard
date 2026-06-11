package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var (
	removeTask int
	removeCol  string
)

var removeCmd = &cobra.Command{
	Use:     "remove [card title]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a card from the board",
	Args: func(cmd *cobra.Command, args []string) error {
		task, _ := cmd.Flags().GetInt("task")
		if task == 0 && len(args) < 1 {
			return fmt.Errorf("requires card title or --task/-t flag")
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

		card, col, idx, err := resolveCard(b, removeTask, removeCol, args)
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
	removeCmd.Flags().IntVarP(&removeTask, "task", "t", 0, "1-based card index within the column")
	removeCmd.Flags().StringVarP(&removeCol, "col", "c", "", "Column for index-based lookup (name, partial, or index)")
	rootCmd.AddCommand(removeCmd)
}
