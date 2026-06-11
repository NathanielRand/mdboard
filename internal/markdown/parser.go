package markdown

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/NathanielRand/mdboard/internal/board"
)

var (
	reFrontmatterDelim = regexp.MustCompile(`^---\s*$`)
	reColumnHeading    = regexp.MustCompile(`^##\s+(.+)$`)
	reCardHeading      = regexp.MustCompile(`^###\s+(.+)$`)
	reMetaComment      = regexp.MustCompile(`<!--(.+)-->`)
	reEmoji            = regexp.MustCompile(`(?:[\x{2600}-\x{27BF}]|[\x{1F300}-\x{1FFFF}])\s*`)
)

// Parse reads a markdown board file into a Board struct
func Parse(path string) (*board.Board, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	b := &board.Board{}
	scanner := bufio.NewScanner(f)

	inFrontmatter := false
	frontmatterDone := false
	lineNum := 0
	var currentCol *board.Column
	var currentCard *board.Card

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Handle frontmatter
		if lineNum == 1 && reFrontmatterDelim.MatchString(line) {
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			if reFrontmatterDelim.MatchString(line) {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
			if strings.HasPrefix(line, "board:") {
				b.Title = strings.TrimSpace(strings.TrimPrefix(line, "board:"))
			}
			continue
		}
		_ = frontmatterDone

		// Column heading (## ...)
		if m := reColumnHeading.FindStringSubmatch(line); m != nil {
			rawName := m[1]
			emoji := extractEmoji(rawName)
			name := strings.TrimSpace(reEmoji.ReplaceAllString(rawName, ""))
			currentCol = &board.Column{Name: name, Emoji: emoji}
			b.Columns = append(b.Columns, currentCol)
			currentCard = nil
			continue
		}

		// Card heading (### ...)
		if m := reCardHeading.FindStringSubmatch(line); m != nil {
			if currentCol == nil {
				continue
			}
			now := time.Now()
			currentCard = &board.Card{
				Title:   strings.TrimSpace(m[1]),
				Status:  board.StatusKey(currentCol.Name),
				Created: &now,
			}
			currentCol.Cards = append(currentCol.Cards, currentCard)
			continue
		}

		// Metadata comment <!-- key: val | key: val -->
		if m := reMetaComment.FindStringSubmatch(line); m != nil && currentCard != nil {
			parseMeta(currentCard, m[1])
			continue
		}

		// Note lines (- ...) attached to current card
		if currentCard != nil && strings.HasPrefix(strings.TrimSpace(line), "- ") {
			currentCard.Notes = append(currentCard.Notes, strings.TrimSpace(line))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if b.Title == "" {
		b.Title = strings.TrimSuffix(path, ".md")
	}

	return b, nil
}

func extractEmoji(s string) string {
	m := reEmoji.FindString(s)
	return strings.TrimSpace(m)
}

func parseMeta(c *board.Card, raw string) {
	parts := strings.Split(raw, "|")
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		switch key {
		case "user":
			c.User = val
		case "status":
			c.Status = val
		case "claimed":
			if t, err := time.Parse("2006-01-02", val); err == nil {
				c.Claimed = &t
			}
		case "completed":
			if t, err := time.Parse("2006-01-02", val); err == nil {
				c.Completed = &t
			}
		case "created":
			if t, err := time.Parse("2006-01-02", val); err == nil {
				c.Created = &t
			}
		}
	}
}
