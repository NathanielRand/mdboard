package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nrand/mdboard/internal/board"
	"github.com/nrand/mdboard/internal/config"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "Create a new board markdown file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		title := strings.Join(args, " ")
		slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
		filename := slug + ".md"

		if _, err := os.Stat(filename); err == nil {
			return fmt.Errorf("file %s already exists", filename)
		}

		b := &board.Board{Title: title}
		for _, colName := range cfg.DefaultColumns {
			b.Columns = append(b.Columns, &board.Column{
				Name:  colName,
				Emoji: board.DefaultEmoji(colName),
			})
		}

		// Write manually since we need clean output
		if err := writeBoardFile(filename, b); err != nil {
			return err
		}

		absPath, _ := filepath.Abs(filename)
		fmt.Printf("✅ Created board: %s\n   → %s\n", title, absPath)
		return nil
	},
}

func writeBoardFile(path string, b *board.Board) error {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("board: %s\n", b.Title))
	sb.WriteString("---\n\n")

	now := time.Now().Format("2006-01-02")

	for i, col := range b.Columns {
		emoji := col.Emoji
		if emoji == "" {
			emoji = board.DefaultEmoji(col.Name)
		}
		sb.WriteString(fmt.Sprintf("## %s %s\n", emoji, col.Name))
		if i == 0 {
			// Seed backlog with a placeholder card
			sb.WriteString(fmt.Sprintf("\n### Example card\n"))
			sb.WriteString(fmt.Sprintf("<!-- status: backlog | created: %s -->\n", now))
			sb.WriteString("- Add your notes here\n")
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func init() {
	rootCmd.AddCommand(newCmd)
}
