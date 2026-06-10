package markdown

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nrand/mdboard/internal/board"
)

// Write serializes a Board back to its markdown file
func Write(path string, b *board.Board) error {
	var sb strings.Builder

	// Frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("board: %s\n", b.Title))
	sb.WriteString("---\n\n")

	for _, col := range b.Columns {
		emoji := col.Emoji
		if emoji == "" {
			emoji = board.DefaultEmoji(col.Name)
		}
		sb.WriteString(fmt.Sprintf("## %s %s\n", emoji, col.Name))

		for _, card := range col.Cards {
			sb.WriteString(fmt.Sprintf("\n### %s\n", card.Title))
			sb.WriteString(buildMeta(card))
			for _, note := range card.Notes {
				sb.WriteString(note + "\n")
			}
		}

		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func buildMeta(c *board.Card) string {
	parts := []string{}

	if c.Status != "" {
		parts = append(parts, fmt.Sprintf("status: %s", c.Status))
	}
	if c.User != "" {
		parts = append(parts, fmt.Sprintf("user: %s", c.User))
	}
	if c.Claimed != nil {
		parts = append(parts, fmt.Sprintf("claimed: %s", c.Claimed.Format("2006-01-02")))
	}
	if c.Completed != nil {
		parts = append(parts, fmt.Sprintf("completed: %s", c.Completed.Format("2006-01-02")))
	}
	if c.Created != nil {
		parts = append(parts, fmt.Sprintf("created: %s", c.Created.Format("2006-01-02")))
	}

	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("<!-- %s -->\n", strings.Join(parts, " | "))
}

// FormatDate formats a time pointer safely
func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
