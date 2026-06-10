package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nrand/mdboard/internal/board"
	"github.com/nrand/mdboard/internal/markdown"
)

// ── Styles ──────────────────────────────────────────────────────────────────

var (
	// Modern Minimalist Theme
	colorAccent  = lipgloss.AdaptiveColor{Light: "#6D28D9", Dark: "#A78BFA"}
	colorText    = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#E5E7EB"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}
	colorDivider = lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"}
	colorBase    = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#0F172A"}

	// App Header & Footer
	styleBoardTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			Padding(0, 2, 1, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorDivider)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(1, 2, 0, 2).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(colorDivider)

	// Header Elements
	styleActiveTitle   = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	styleInactiveTitle = lipgloss.NewStyle().Foreground(colorMuted)

	styleActiveBadge = lipgloss.NewStyle().
				Background(colorAccent).
				Foreground(colorBase).
				Bold(true).
				Padding(0, 1)

	styleInactiveBadge = lipgloss.NewStyle().
				Background(colorDivider).
				Foreground(colorText).
				Padding(0, 1)
)

func getDelegate(isActiveColumn bool) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	d.SetSpacing(1)

	// LOOSENED TRUNCATION: Reduced left padding from 2 to 1 to give text more horizontal room
	if isActiveColumn {
		d.Styles.SelectedTitle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(colorAccent).
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 0, 0, 1)

		d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy().Foreground(colorMuted).Bold(false)
		d.Styles.NormalTitle = lipgloss.NewStyle().Foreground(colorText).Padding(0, 0, 0, 1)
		d.Styles.NormalDesc = d.Styles.NormalTitle.Copy().Foreground(colorMuted)
	} else {
		d.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 0, 0, 1)
		d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy()
		d.Styles.NormalTitle = lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 0, 0, 1)
		d.Styles.NormalDesc = d.Styles.NormalTitle.Copy()
	}
	return d
}

// ── List Item Wrapper ───────────────────────────────────────────────────────

type cardItem struct {
	*board.Card
}

func (c cardItem) Title() string { return c.Card.Title }
func (c cardItem) Description() string {
	var desc []string
	if c.Card.User != "" {
		desc = append(desc, "@"+c.Card.User)
	}
	if c.Card.Status != "" {
		desc = append(desc, strings.ToLower(string(c.Card.Status)))
	}
	return strings.Join(desc, " • ")
}
func (c cardItem) FilterValue() string { return c.Card.Title }

// ── Model ───────────────────────────────────────────────────────────────────

type Model struct {
	board     *board.Board
	boardPath string
	lists     []list.Model
	colIdx    int
	width     int
	height    int
}

func NewModel(b *board.Board, path string) Model {
	m := Model{
		board:     b,
		boardPath: path,
		lists:     make([]list.Model, len(b.Columns)),
	}

	for i, col := range b.Columns {
		items := make([]list.Item, len(col.Cards))
		for j, card := range col.Cards {
			items[j] = cardItem{card}
		}

		l := list.New(items, getDelegate(i == 0), 0, 0)
		l.SetShowTitle(false)
		l.SetShowStatusBar(false)
		l.SetShowPagination(false)
		l.SetShowHelp(false)
		l.SetShowFilter(false)

		m.lists[i] = l
	}

	return m
}

func (m Model) Init() tea.Cmd { return nil }

type editorFinishedMsg struct {
	err  error
	file string
	card *board.Card
}

func openEditorCmd(m Model) tea.Cmd {
	selected := m.lists[m.colIdx].SelectedItem()
	if selected == nil {
		return nil
	}
	card := selected.(cardItem).Card

	f, err := os.CreateTemp("", "mdboard-*.md")
	if err != nil {
		return nil
	}

	content := fmt.Sprintf("# %s\n\n%s", card.Title, card.Body)
	f.Write([]byte(content))
	f.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // Fallback
	}

	c := exec.Command(editor, f.Name())
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err, file: f.Name(), card: card}
	})
}

func (m Model) refreshLists() Model {
	for i, col := range m.board.Columns {
		items := make([]list.Item, len(col.Cards))
		for j, card := range col.Cards {
			items[j] = cardItem{card}
		}
		m.lists[i].SetItems(items)
	}
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	listMsg := msg

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		dividers := len(m.lists) - 1
		colWidth := (msg.Width - dividers) / len(m.lists)

		// LOOSENED TRUNCATION: padding buffer reduced from 4 to 2
		innerWidth := colWidth - 2
		if innerWidth < 0 {
			innerWidth = 0
		}

		// Reserved vertical space increased from 10 to 18 to make room for the new Preview Pane
		listHeight := msg.Height - 18
		if listHeight < 0 {
			listHeight = 0
		}

		for i := range m.lists {
			m.lists[i].SetSize(innerWidth, listHeight)
		}
		return m, nil

	case editorFinishedMsg:
		if msg.err == nil {
			content, _ := os.ReadFile(msg.file)
			lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")

			var newTitle, newBody string
			if len(lines) > 0 {
				newTitle = strings.TrimSpace(strings.TrimPrefix(lines[0], "# "))
				if len(lines) > 1 {
					newBody = strings.TrimSpace(strings.Join(lines[1:], "\n"))
				}
			}

			if newTitle != "" {
				board.UpdateCard(msg.card, newTitle, newBody)
				_ = markdown.Write(m.boardPath, m.board)
				m = m.refreshLists()
			}
		}
		os.Remove(msg.file)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "a", "left":
			if m.colIdx > 0 {
				m.colIdx--
			}
			return m, nil
		case "d", "right":
			if m.colIdx < len(m.lists)-1 {
				m.colIdx++
			}
			return m, nil
		case "w":
			listMsg = tea.KeyMsg{Type: tea.KeyUp}
		case "s":
			listMsg = tea.KeyMsg{Type: tea.KeyDown}
		case "h", "j", "k", "l":
			return m, nil

		case "e", "enter":
			return m, openEditorCmd(m)

		case "A":
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil && m.colIdx > 0 {
				card := selected.(cardItem).Card
				fromCol := m.board.Columns[m.colIdx]
				toCol := m.board.Columns[m.colIdx-1]
				idx := m.lists[m.colIdx].Index()

				board.MoveCard(m.board, card, fromCol, idx, toCol)
				_ = markdown.Write(m.boardPath, m.board)
				m = m.refreshLists()

				m.colIdx--
				m.lists[m.colIdx].Select(len(toCol.Cards) - 1)
			}
			return m, nil
		case "D":
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil && m.colIdx < len(m.board.Columns)-1 {
				card := selected.(cardItem).Card
				fromCol := m.board.Columns[m.colIdx]
				toCol := m.board.Columns[m.colIdx+1]
				idx := m.lists[m.colIdx].Index()

				board.MoveCard(m.board, card, fromCol, idx, toCol)
				_ = markdown.Write(m.boardPath, m.board)
				m = m.refreshLists()

				m.colIdx++
				m.lists[m.colIdx].Select(len(toCol.Cards) - 1)
			}
			return m, nil

		case "W":
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil {
				idx := m.lists[m.colIdx].Index()
				col := m.board.Columns[m.colIdx]
				if err := board.ShiftCard(col, idx, true); err == nil {
					_ = markdown.Write(m.boardPath, m.board)
					m = m.refreshLists()
					m.lists[m.colIdx].Select(idx - 1)
				}
			}
			return m, nil
		case "S":
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil {
				idx := m.lists[m.colIdx].Index()
				col := m.board.Columns[m.colIdx]
				if err := board.ShiftCard(col, idx, false); err == nil {
					_ = markdown.Write(m.boardPath, m.board)
					m = m.refreshLists()
					m.lists[m.colIdx].Select(idx + 1)
				}
			}
			return m, nil

		case "x", "backspace", "delete":
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil {
				card := selected.(cardItem).Card
				_, _ = board.RemoveCard(m.board, card.Title)
				_ = markdown.Write(m.boardPath, m.board)
				m = m.refreshLists()
			}
			return m, nil
		}
	}

	if len(m.lists) > 0 {
		m.lists[m.colIdx], cmd = m.lists[m.colIdx].Update(listMsg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}

	var cols []string

	dividers := len(m.lists) - 1
	colWidth := (m.width - dividers) / len(m.lists)

	// Must match the math in WindowSizeMsg
	innerWidth := colWidth - 2

	for i, l := range m.lists {
		isActive := i == m.colIdx
		l.SetDelegate(getDelegate(isActive))

		colData := m.board.Columns[i]

		header := renderHeader("", colData.Name, len(colData.Cards), isActive, innerWidth)
		colContent := lipgloss.JoinVertical(lipgloss.Left, header, l.View())

		// LOOSENED TRUNCATION: Padding reduced from 2 to 1 on the columns themselves
		colStyle := lipgloss.NewStyle().Padding(0, 1).Width(colWidth)

		if i < len(m.lists)-1 {
			colStyle = colStyle.
				Border(lipgloss.NormalBorder(), false, true, false, false).
				BorderForeground(colorDivider)
		}

		cols = append(cols, colStyle.Render(colContent))
	}

	boardTitle := styleBoardTitle.Copy().Width(m.width - 4).Render("📋 " + m.board.Title)

	// Create the Preview Pane based on the currently selected item
	var previewPane string
	selected := m.lists[m.colIdx].SelectedItem()
	if selected != nil {
		card := selected.(cardItem).Card
		previewPane = renderPreviewPane(card, m.width)
	} else {
		previewPane = renderPreviewPane(nil, m.width)
	}

	helpView := styleHelp.Copy().Width(m.width - 4).Render("←/→/a/d: cols • ↑/↓/w/s: cards • A/D: move ↔ • W/S: shift ↕ • e: edit • x: del • q: quit")

	boardView := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	return lipgloss.JoinVertical(lipgloss.Left, boardTitle, boardView, previewPane, helpView)
}

func renderPreviewPane(card *board.Card, width int) string {
	boxStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(colorDivider).
		Width(width - 4).
		Height(6) // Fixed 6-line height to prevent UI bouncing

	if card == nil {
		return boxStyle.Render("")
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render(card.Title)

	body := card.Body
	if body == "" {
		body = lipgloss.NewStyle().Italic(true).Foreground(colorMuted).Render("No additional details.")
	} else {
		body = strings.TrimSpace(body)
	}

	// Leverage Lipgloss to automatically wrap text to screen width
	wrappedBody := lipgloss.NewStyle().
		Width(width - 8).
		Foreground(colorText).
		Render(body)

	content := title + "\n\n" + wrappedBody

	// Failsafe to prevent vertical overflow from destroying the TUI layout
	lines := strings.Split(content, "\n")
	if len(lines) > 6 {
		lines = lines[:6]
		lines[5] = lipgloss.NewStyle().Foreground(colorMuted).Render("  ... press 'e' to read more")
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderHeader(emoji, name string, count int, isActive bool, innerWidth int) string {
	titleText := name
	if emoji != "" {
		titleText = fmt.Sprintf("%s %s", emoji, name)
	}
	countText := fmt.Sprintf("%d", count)

	var title, badge string
	var borderStyle lipgloss.Border
	var borderColor lipgloss.AdaptiveColor

	if isActive {
		title = styleActiveTitle.Render(titleText)
		badge = styleActiveBadge.Render(countText)
		borderStyle = lipgloss.ThickBorder()
		borderColor = colorAccent
	} else {
		title = styleInactiveTitle.Render(titleText)
		badge = styleInactiveBadge.Render(countText)
		borderStyle = lipgloss.NormalBorder()
		borderColor = colorDivider
	}

	titleW := lipgloss.Width(title)
	badgeW := lipgloss.Width(badge)
	gap := innerWidth - titleW - badgeW

	if gap < 0 {
		gap = 0
	}
	spacer := strings.Repeat(" ", gap)

	content := title + spacer + badge

	return lipgloss.NewStyle().
		Border(borderStyle, false, false, true, false).
		BorderForeground(borderColor).
		PaddingBottom(1).
		Render(content)
}

func Run(b *board.Board, path string) error {
	p := tea.NewProgram(NewModel(b, path), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
