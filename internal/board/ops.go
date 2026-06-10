package board

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sahilm/fuzzy"
)

// FindCard searches all columns for a card by title (case-insensitive partial match)
// Returns the card, its column, and its index within the column
func FindCard(b *Board, title string) (*Card, *Column, int, error) {
	lower := strings.ToLower(title)
	var matches []*struct {
		card  *Card
		col   *Column
		index int
	}

	for _, col := range b.Columns {
		for i, card := range col.Cards {
			if strings.Contains(strings.ToLower(card.Title), lower) {
				matches = append(matches, &struct {
					card  *Card
					col   *Column
					index int
				}{card, col, i})
			}
		}
	}

	if len(matches) == 0 {
		return nil, nil, -1, fmt.Errorf("no card matching %q found", title)
	}
	if len(matches) > 1 {
		titles := make([]string, len(matches))
		for i, m := range matches {
			titles[i] = fmt.Sprintf("  [%s] %s", m.col.Name, m.card.Title)
		}
		return nil, nil, -1, fmt.Errorf("ambiguous match — multiple cards found:\n%s", strings.Join(titles, "\n"))
	}

	m := matches[0]
	return m.card, m.col, m.index, nil
}

// FindColumn searches for a column by name (case-insensitive partial match) or by
// 1-based numeric index (e.g. "1" for the first column).
func FindColumn(b *Board, name string) (*Column, error) {
	if n, err := strconv.Atoi(name); err == nil {
		if n >= 1 && n <= len(b.Columns) {
			return b.Columns[n-1], nil
		}
		return nil, fmt.Errorf("column index %d out of range (1-%d)", n, len(b.Columns))
	}

	lower := strings.ToLower(name)
	var matches []*Column
	for _, col := range b.Columns {
		if strings.Contains(strings.ToLower(col.Name), lower) {
			matches = append(matches, col)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, c := range matches {
			names[i] = c.Name
		}
		return nil, fmt.Errorf("ambiguous column match: %s", strings.Join(names, ", "))
	}

	// Fuzzy fallback
	colNames := make([]string, len(b.Columns))
	for i, col := range b.Columns {
		colNames[i] = col.Name
	}
	results := fuzzy.Find(name, colNames)
	if len(results) == 0 {
		return nil, fmt.Errorf("no column matching %q", name)
	}
	return b.Columns[results[0].Index], nil
}

// MoveCard removes a card from its current column and appends it to the target
func MoveCard(b *Board, card *Card, fromCol *Column, fromIdx int, toCol *Column) {
	// Remove from source
	fromCol.Cards = append(fromCol.Cards[:fromIdx], fromCol.Cards[fromIdx+1:]...)

	// Update status
	card.Status = StatusKey(toCol.Name)

	// Set completed timestamp if moving to Done
	if strings.ToLower(toCol.Name) == "done" {
		now := time.Now()
		card.Completed = &now
	} else {
		card.Completed = nil
	}

	toCol.Cards = append(toCol.Cards, card)
}

// ShiftCard moves a card up or down by one position within its current column.
// up=true moves it closer to index 0 (top of the column).
func ShiftCard(col *Column, idx int, up bool) error {
	targetIdx := idx + 1
	if up {
		targetIdx = idx - 1
	}

	// Check bounds
	if targetIdx < 0 || targetIdx >= len(col.Cards) {
		return fmt.Errorf("cannot shift card further in that direction")
	}

	// Swap the cards
	col.Cards[idx], col.Cards[targetIdx] = col.Cards[targetIdx], col.Cards[idx]
	return nil
}

// UpdateCard updates the title and/or body of a card.
// Pass empty strings if you don't want to change a specific field.
func UpdateCard(card *Card, newTitle string, newBody string) {
	if newTitle != "" {
		card.Title = newTitle
	}
	if newBody != "" {
		// Assuming your Card struct has a Body/Content field
		// for the markdown bullets under the title
		card.Body = newBody
	}
}

// ClaimCard stamps a git username on a card
func ClaimCard(card *Card, user string) {
	now := time.Now()
	card.User = user
	card.Claimed = &now
}

// UnclaimCard removes the user/claimed stamp from a card
func UnclaimCard(card *Card) {
	card.User = ""
	card.Claimed = nil
}

// AddCard appends a new card to the named column
func AddCard(b *Board, title, colName string) (*Card, *Column, error) {
	col, err := FindColumn(b, colName)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	card := &Card{
		Title:   title,
		Status:  StatusKey(col.Name),
		Created: &now,
	}
	col.Cards = append(col.Cards, card)
	return card, col, nil
}

// RemoveCard deletes a card from the board entirely
func RemoveCard(b *Board, title string) (*Card, error) {
	card, col, idx, err := FindCard(b, title)
	if err != nil {
		return nil, err
	}
	col.Cards = append(col.Cards[:idx], col.Cards[idx+1:]...)
	return card, nil
}
