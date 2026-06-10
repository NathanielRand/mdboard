package cmd

import (
	"fmt"
	"os"

	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/config"
	"github.com/NathanielRand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var claimUser string

var claimCmd = &cobra.Command{
	Use:   "claim [card title]",
	Short: "Claim a card with your git username",
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

		card, _, _, err := board.FindCard(b, joinArgs(args))
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

var unclaimCmd = &cobra.Command{
	Use:   "unclaim [card title]",
	Short: "Remove your claim from a card",
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

		card, _, _, err := board.FindCard(b, joinArgs(args))
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
	rootCmd.AddCommand(claimCmd)
	rootCmd.AddCommand(unclaimCmd)
}
