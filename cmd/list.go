package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nrand/mdboard/internal/markdown"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all .md board files in the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := os.ReadDir(".")
		if err != nil {
			return err
		}

		var boards []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				boards = append(boards, e.Name())
			}
		}

		if len(boards) == 0 {
			fmt.Println("No .md board files found in current directory.")
			fmt.Println("Run: mdboard new \"My Board\"")
			return nil
		}

		fmt.Printf("Found %d board(s):\n\n", len(boards))
		for _, name := range boards {
			b, err := markdown.Parse(name)
			if err != nil {
				fmt.Printf("  📄 %-30s (parse error)\n", name)
				continue
			}

			totalCards := 0
			for _, col := range b.Columns {
				totalCards += len(col.Cards)
			}

			fmt.Printf("  📋 %-30s %s  (%d cards)\n",
				name,
				b.Title,
				totalCards,
			)
		}
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print a quick text summary of the board",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolveBoardPath(cmd)
		if err != nil {
			return err
		}

		b, err := markdown.Parse(path)
		if err != nil {
			return err
		}

		abs, _ := filepath.Abs(path)
		fmt.Printf("\n📋 %s\n%s\n\n", b.Title, abs)

		for _, col := range b.Columns {
			emoji := col.Emoji
			if emoji == "" {
				emoji = "•"
			}
			fmt.Printf("%s %s (%d)\n", emoji, col.Name, len(col.Cards))
			if len(col.Cards) == 0 {
				fmt.Println("   (empty)")
			}
			for _, card := range col.Cards {
				claim := ""
				if card.User != "" {
					claim = fmt.Sprintf(" @%s", card.User)
				}
				fmt.Printf("   • %s%s\n", card.Title, claim)
			}
			fmt.Println()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)
}
