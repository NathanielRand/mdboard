package board

import "time"

// Card represents a single task/item on the board
type Card struct {
	Title     string
	Notes     []string // bullet lines under the card
	Status    string
	User      string
	Body      string
	Claimed   *time.Time
	Completed *time.Time
	Created   *time.Time
}

// Column represents a vertical lane on the board
type Column struct {
	Name  string
	Emoji string
	Cards []*Card
}

// Board represents a full .md board file
type Board struct {
	Title   string
	Columns []*Column
}

// StatusKey returns the canonical slug for a column name
func StatusKey(name string) string {
	switch name {
	case "In Progress", "in-progress", "in_progress":
		return "in-progress"
	case "Done", "done":
		return "done"
	case "Testing", "testing":
		return "testing"
	default:
		return "backlog"
	}
}

// DefaultEmoji returns a sensible emoji for a column name
func DefaultEmoji(name string) string {
	switch name {
	case "Backlog":
		return "📋"
	case "In Progress":
		return "🔄"
	case "Testing":
		return "🧪"
	case "Done":
		return "✅"
	default:
		return "📌"
	}
}
