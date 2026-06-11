package cmd

import (
	"fmt"
	"os"

	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/config"
	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var (
	claimUser string
	claimTask int
	claimCol  string
)

var claimCmd = &cobra.Command{
	Use:   "claim [card title]",
	Short: "Claim a card with your git username",
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

		user := claimUser
		if user == "" {
			cwd, _ := os.Getwd()
			if pc, err := config.LoadProjectAt(cwd); err == nil {
				user = pc.GitUser
			}
		}
		if user == "" {
			return fmt.Errorf("no git user set — use --user flag or run: mdboard config set git_user <username>")
		}

		card, _, _, err := resolveCard(b, claimTask, claimCol, args)
		if err != nil {
			return err
		}

		if card.User != "" && card.User != user {
			fmt.Printf("⚠️  Card \"%s\" is already claimed by @%s\n", card.Title, card.User)
			fmt.Print("Override? [y/N] ")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		board.ClaimCard(card, user)

		if err := markdown.Write(path, b); err != nil {
			return err
		}

		fmt.Printf("🙋 Claimed \"%s\" → @%s\n", card.Title, user)
		return nil
	},
}

var (
	unclaimTask int
	unclaimCol  string
)

var unclaimCmd = &cobra.Command{
	Use:   "unclaim [card title]",
	Short: "Remove your claim from a card",
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

		card, _, _, err := resolveCard(b, unclaimTask, unclaimCol, args)
		if err != nil {
			return err
		}

		if card.User == "" {
			fmt.Printf("ℹ️  Card \"%s\" has no claim\n", card.Title)
			return nil
		}

		prev := card.User
		board.UnclaimCard(card)

		if err := markdown.Write(path, b); err != nil {
			return err
		}

		fmt.Printf("🔓 Unclaimed \"%s\" (was @%s)\n", card.Title, prev)
		return nil
	},
}

func init() {
	claimCmd.Flags().StringVarP(&claimUser, "user", "u", "", "Git username to claim with (overrides config)")
	claimCmd.Flags().IntVarP(&claimTask, "task", "t", 0, "1-based card index within the column")
	claimCmd.Flags().StringVarP(&claimCol, "col", "c", "", "Column for index-based lookup (name, partial, or index)")

	unclaimCmd.Flags().IntVarP(&unclaimTask, "task", "t", 0, "1-based card index within the column")
	unclaimCmd.Flags().StringVarP(&unclaimCol, "col", "c", "", "Column for index-based lookup (name, partial, or index)")

	rootCmd.AddCommand(claimCmd)
	rootCmd.AddCommand(unclaimCmd)
}
