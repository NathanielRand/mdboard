package tui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/markdown"
)

// ── Styles ──────────────────────────────────────────────────────────────────

var (
	// Modern Minimalist Theme
	colorAccent  = lipgloss.AdaptiveColor{Light: "#6D28D9", Dark: "#A78BFA"}
	colorText    = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#E5E7EB"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}
	colorDivider = lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"}
	colorBase    = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#0F172A"}

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
	index int // 1-based position within the column
}

func (c cardItem) Title() string { return c.Card.Title }
func (c cardItem) Description() string {
	parts := []string{fmt.Sprintf("%d", c.index)}
	if c.Card.User != "" {
		parts = append(parts, "@"+c.Card.User)
	}
	if c.Card.Status != "" {
		parts = append(parts, strings.ToLower(string(c.Card.Status)))
	}
	return strings.Join(parts, " · ")
}
func (c cardItem) FilterValue() string { return c.Card.Title }

// ── Messages ────────────────────────────────────────────────────────────────

type tickMsg time.Time
type versionCheckMsg struct{ latest string }
type editorFinishedMsg struct {
	err  error
	file string
	card *board.Card
}

// ── Model ───────────────────────────────────────────────────────────────────

type Model struct {
	board           *board.Board
	boardPath       string
	lists           []list.Model
	colIdx          int
	width           int
	height          int
	lastMod         time.Time
	creating        bool
	editing         bool
	editingCard     *board.Card
	input           textinput.Model
	frame           int
	version         string
	updateAvailable bool
	gitUser         string
}

func NewModel(b *board.Board, path, version, gitUser string) Model {
	m := Model{
		board:     b,
		boardPath: path,
		lists:     make([]list.Model, len(b.Columns)),
		version:   version,
		gitUser:   gitUser,
	}

	if info, err := os.Stat(path); err == nil {
		m.lastMod = info.ModTime()
	}

	inp := textinput.New()
	inp.Placeholder = "Card title..."
	inp.CharLimit = 0
	m.input = inp

	for i, col := range b.Columns {
		items := make([]list.Item, len(col.Cards))
		for j, card := range col.Cards {
			items[j] = cardItem{card, j + 1}
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

type animTickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func animTickCmd() tea.Cmd {
	return tea.Tick(350*time.Millisecond, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

func checkVersionCmd(current string) tea.Cmd {
	return func() tea.Msg {
		if current == "dev" {
			return versionCheckMsg{}
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("https://api.github.com/repos/NathanielRand/mdboard/releases/latest")
		if err != nil {
			return versionCheckMsg{}
		}
		defer resp.Body.Close()
		var payload struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return versionCheckMsg{}
		}
		return versionCheckMsg{latest: payload.TagName}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), animTickCmd(), checkVersionCmd(m.version))
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
		editor = "nano"
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
			items[j] = cardItem{card, j + 1}
		}
		m.lists[i].SetItems(items)
	}
	return m
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	listMsg := msg

	switch msg := msg.(type) {
	case animTickMsg:
		m.frame++
		return m, animTickCmd()

	case versionCheckMsg:
		if msg.latest != "" && msg.latest != m.version {
			m.updateAvailable = true
		}
		return m, nil

	case tickMsg:
		info, err := os.Stat(m.boardPath)
		if err == nil && info.ModTime().After(m.lastMod) {
			if b, err := markdown.Parse(m.boardPath); err == nil {
				m.board = b
				m.lastMod = info.ModTime()
				m = m.refreshLists()
			}
		}
		return m, tickCmd()

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

		// Reserved vertical space: header(2) + preview(8) + help(4) + footer(2) + margins(3) = 19
		listHeight := msg.Height - 19
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
				if info, err := os.Stat(m.boardPath); err == nil {
					m.lastMod = info.ModTime()
				}
				m = m.refreshLists()
			}
		}
		os.Remove(msg.file)
		return m, nil

	case tea.KeyMsg:
		if m.creating {
			switch msg.String() {
			case "enter":
				title := strings.TrimSpace(m.input.Value())
				if title != "" {
					col := m.board.Columns[m.colIdx]
					if _, _, err := board.AddCard(m.board, title, col.Name); err == nil {
						_ = markdown.Write(m.boardPath, m.board)
						if info, err := os.Stat(m.boardPath); err == nil {
							m.lastMod = info.ModTime()
						}
						m = m.refreshLists()
						m.lists[m.colIdx].Select(len(col.Cards) - 1)
					}
				}
				m.creating = false
				m.input.Blur()
				m.input.Reset()
				return m, nil
			case "esc":
				m.creating = false
				m.input.Blur()
				m.input.Reset()
				return m, nil
			default:
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
		if m.editing {
			switch msg.String() {
			case "enter":
				newTitle := strings.TrimSpace(m.input.Value())
				if newTitle != "" && m.editingCard != nil {
					m.editingCard.Title = newTitle
					_ = markdown.Write(m.boardPath, m.board)
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
					m = m.refreshLists()
				}
				m.editing = false
				m.editingCard = nil
				m.input.Blur()
				m.input.Reset()
				return m, nil
			case "esc":
				m.editing = false
				m.editingCard = nil
				m.input.Blur()
				m.input.Reset()
				return m, nil
			default:
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
		// Type-based pre-check catches ALL escape-sequence variants of shift+arrow keys.
		// Some terminals emit \x1b[1;4A (Alt bit set) instead of \x1b[1;2A for Shift+Up,
		// making msg.String() return "alt+shift+up" rather than "shift+up" — bypassing
		// the string switch. Checking msg.Type directly sidesteps that entirely.
		switch msg.Type {
		case tea.KeyShiftLeft:
			if m.colIdx > 0 {
				selected := m.lists[m.colIdx].SelectedItem()
				if selected != nil {
					card := selected.(cardItem).Card
					fromCol := m.board.Columns[m.colIdx]
					toCol := m.board.Columns[m.colIdx-1]
					idx := m.lists[m.colIdx].Index()
					board.MoveCard(m.board, card, fromCol, idx, toCol)
					_ = markdown.Write(m.boardPath, m.board)
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
					m = m.refreshLists()
					m.colIdx--
					m.lists[m.colIdx].Select(len(toCol.Cards) - 1)
				}
			}
			return m, nil
		case tea.KeyShiftRight:
			if m.colIdx < len(m.board.Columns)-1 {
				selected := m.lists[m.colIdx].SelectedItem()
				if selected != nil {
					card := selected.(cardItem).Card
					fromCol := m.board.Columns[m.colIdx]
					toCol := m.board.Columns[m.colIdx+1]
					idx := m.lists[m.colIdx].Index()
					board.MoveCard(m.board, card, fromCol, idx, toCol)
					_ = markdown.Write(m.boardPath, m.board)
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
					m = m.refreshLists()
					m.colIdx++
					m.lists[m.colIdx].Select(len(toCol.Cards) - 1)
				}
			}
			return m, nil
		case tea.KeyShiftUp:
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil {
				idx := m.lists[m.colIdx].Index()
				col := m.board.Columns[m.colIdx]
				if err := board.ShiftCard(col, idx, true); err == nil {
					_ = markdown.Write(m.boardPath, m.board)
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
					m = m.refreshLists()
					m.lists[m.colIdx].Select(idx - 1)
				}
			}
			return m, nil
		case tea.KeyShiftDown:
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil {
				idx := m.lists[m.colIdx].Index()
				col := m.board.Columns[m.colIdx]
				if err := board.ShiftCard(col, idx, false); err == nil {
					_ = markdown.Write(m.boardPath, m.board)
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
					m = m.refreshLists()
					m.lists[m.colIdx].Select(idx + 1)
				}
			}
			return m, nil
		}

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

		case "e":
			return m, openEditorCmd(m)

		case "enter":
			return m, openEditorCmd(m)

		case "n":
			m.creating = true
			return m, m.input.Focus()

		case "r":
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil {
				m.editingCard = selected.(cardItem).Card
				m.editing = true
				m.input.SetValue(m.editingCard.Title)
				return m, m.input.Focus()
			}
			return m, nil

		// Capital letters are shift+letter (reliable ASCII fallbacks for shift+arrow)
		case "A":
			if m.colIdx > 0 {
				selected := m.lists[m.colIdx].SelectedItem()
				if selected != nil {
					card := selected.(cardItem).Card
					fromCol := m.board.Columns[m.colIdx]
					toCol := m.board.Columns[m.colIdx-1]
					idx := m.lists[m.colIdx].Index()
					board.MoveCard(m.board, card, fromCol, idx, toCol)
					_ = markdown.Write(m.boardPath, m.board)
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
					m = m.refreshLists()
					m.colIdx--
					m.lists[m.colIdx].Select(len(toCol.Cards) - 1)
				}
			}
			return m, nil
		case "D":
			if m.colIdx < len(m.board.Columns)-1 {
				selected := m.lists[m.colIdx].SelectedItem()
				if selected != nil {
					card := selected.(cardItem).Card
					fromCol := m.board.Columns[m.colIdx]
					toCol := m.board.Columns[m.colIdx+1]
					idx := m.lists[m.colIdx].Index()
					board.MoveCard(m.board, card, fromCol, idx, toCol)
					_ = markdown.Write(m.boardPath, m.board)
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
					m = m.refreshLists()
					m.colIdx++
					m.lists[m.colIdx].Select(len(toCol.Cards) - 1)
				}
			}
			return m, nil
		case "W":
			selected := m.lists[m.colIdx].SelectedItem()
			if selected != nil {
				idx := m.lists[m.colIdx].Index()
				col := m.board.Columns[m.colIdx]
				if err := board.ShiftCard(col, idx, true); err == nil {
					_ = markdown.Write(m.boardPath, m.board)
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
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
					if info, err := os.Stat(m.boardPath); err == nil {
						m.lastMod = info.ModTime()
					}
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
				if info, err := os.Stat(m.boardPath); err == nil {
					m.lastMod = info.ModTime()
				}
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

		header := renderHeader(colData.Name, i+1, len(colData.Cards), m.frame, isActive, innerWidth)
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

	displayPath := filepath.Base(filepath.Dir(m.boardPath)) + "/" + filepath.Base(m.boardPath)

	user := m.gitUser
	if user == "" {
		user = "guest"
	}

	navLeft := lipgloss.NewStyle().Bold(true).Foreground(colorAccent).Render(m.board.Title)
	navCenter := lipgloss.NewStyle().Foreground(colorMuted).Render(user)
	navRight := lipgloss.NewStyle().Foreground(colorMuted).Render(displayPath)

	innerW := m.width - 8 // Width(m.width-4) minus Padding(0,2) on each side
	lw, cw, rw := lipgloss.Width(navLeft), lipgloss.Width(navCenter), lipgloss.Width(navRight)
	gap := innerW - lw - cw - rw
	if gap < 2 {
		gap = 2
	}
	lg, rg := gap/2, gap-gap/2

	boardTitle := lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(0, 2).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(colorDivider).
		Render(navLeft + strings.Repeat(" ", lg) + navCenter + strings.Repeat(" ", rg) + navRight)

	// Create the Preview Pane based on the currently selected item
	var previewPane string
	selected := m.lists[m.colIdx].SelectedItem()
	if selected != nil {
		card := selected.(cardItem).Card
		previewPane = renderPreviewPane(card, m.width)
	} else {
		previewPane = renderPreviewPane(nil, m.width)
	}

	var helpView string
	if m.creating {
		col := m.board.Columns[m.colIdx]
		prompt := lipgloss.NewStyle().Foreground(colorMuted).Render(fmt.Sprintf("New card in [%s]  ·  enter to create  ·  esc to cancel", col.Name))
		helpView = styleHelp.Copy().Width(m.width - 4).Render(prompt + "\n" + m.input.View())
	} else if m.editing {
		prompt := lipgloss.NewStyle().Foreground(colorMuted).Render("Rename card  ·  enter to save  ·  esc to cancel")
		helpView = styleHelp.Copy().Width(m.width - 4).Render(prompt + "\n" + m.input.View())
	} else {
		keys := "a/d or ←/→ cols  ·  w/s or ↑/↓ cards  ·  A/D move col  ·  W/S reorder  ·  n new  ·  r rename  ·  e edit  ·  x delete  ·  q quit"
		helpView = styleHelp.Copy().Width(m.width - 4).Render(keys)
	}

	// Footer: version left · license center · sponsor right
	verLabel := m.version
	if verLabel == "" || verLabel == "dev" {
		verLabel = "dev"
	}
	if m.updateAvailable {
		verLabel += "  ↑ update — run: mdb upgrade"
	}
	fLeft := lipgloss.NewStyle().Foreground(colorMuted).Render(verLabel)
	fCenter := lipgloss.NewStyle().Foreground(colorMuted).Render("MIT License")
	fRight := lipgloss.NewStyle().Foreground(colorAccent).Render("♥  sponsor mdboard")

	fInnerW := m.width - 8
	flw, fcw, frw := lipgloss.Width(fLeft), lipgloss.Width(fCenter), lipgloss.Width(fRight)
	fGap := fInnerW - flw - fcw - frw
	if fGap < 2 {
		fGap = 2
	}
	flg, frg := fGap/2, fGap-fGap/2

	footerView := lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(0, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(colorDivider).
		Render(fLeft + strings.Repeat(" ", flg) + fCenter + strings.Repeat(" ", frg) + fRight)

	boardView := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	return lipgloss.JoinVertical(lipgloss.Left, boardTitle, boardView, previewPane, helpView, footerView)
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

// columnIcon returns an animated pixel icon for a column when active, static when not.
func columnIcon(name string, isActive bool, frame int) string {
	lower := strings.ToLower(name)

	var activeFrames []string
	var staticChar string

	switch {
	case strings.Contains(lower, "progress") || strings.Contains(lower, "doing") || strings.Contains(lower, "active"):
		activeFrames = []string{"◐", "◓", "◑", "◒"}
		staticChar = "◎"
	case strings.Contains(lower, "test") || strings.Contains(lower, "review") || strings.Contains(lower, "qa"):
		activeFrames = []string{"⊙", "⊕", "⊙", "⊗"}
		staticChar = "⊙"
	case strings.Contains(lower, "done") || strings.Contains(lower, "complete") || strings.Contains(lower, "finish"):
		activeFrames = []string{"✦", "✧", "✦", "✧"}
		staticChar = "✦"
	default: // backlog, todo, icebox, etc.
		activeFrames = []string{"▤", "▥", "▤", "▥"}
		staticChar = "▤"
	}

	if isActive {
		icon := activeFrames[frame%len(activeFrames)]
		return lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(icon)
	}
	return lipgloss.NewStyle().Foreground(colorMuted).Render(staticChar)
}

func renderHeader(name string, colIdx, count, frame int, isActive bool, innerWidth int) string {
	icon := columnIcon(name, isActive, frame)
	idxName := fmt.Sprintf("%d  %s", colIdx, name)
	countText := fmt.Sprintf("%d", count)

	var title, badge string
	var borderStyle lipgloss.Border
	var borderColor lipgloss.AdaptiveColor

	if isActive {
		title = icon + " " + styleActiveTitle.Render(idxName)
		badge = styleActiveBadge.Render(countText)
		borderStyle = lipgloss.ThickBorder()
		borderColor = colorAccent
	} else {
		title = icon + " " + styleInactiveTitle.Render(idxName)
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

func Run(b *board.Board, path, version, primaryColor, gitUser string) error {
	if primaryColor != "" {
		colorAccent = lipgloss.AdaptiveColor{Light: primaryColor, Dark: primaryColor}
		styleActiveTitle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
		styleActiveBadge = lipgloss.NewStyle().
			Background(colorAccent).
			Foreground(colorBase).
			Bold(true).
			Padding(0, 1)
	}
	p := tea.NewProgram(NewModel(b, path, version, gitUser), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
