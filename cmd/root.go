package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NathanielRand/mdboard/internal/board"
	"github.com/NathanielRand/mdboard/internal/config"
	"github.com/spf13/cobra"
)

var boardFile string

var rootCmd = &cobra.Command{
	Use:   "mdboard | mdb",
	Short: "Markdown-based kanban boards for the terminal",
	Long: `mdboard turns markdown files into interactive kanban boards.

Each .md file is a board. Cards live under ## column headings as ### card titles.
Metadata is stored in HTML comments so the file stays human-readable.

✨ Aliases: You can use the 'mdb' command interchangeably with 'mdboard'.

Examples:
  mdb new "Project Roadmap"
  mdb add "Fix login bug" --col "In Progress"
  mdb claim "Fix login bug"
  mdb move "Fix login bug" Done
  mdb view
  mdb status`,
}

// Execute is the entry point for the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&boardFile, "file", "f", "", "Board file to operate on (default: mdboard.md)")
	rootCmd.Version = Version
}

// resolveBoardPath finds the target board file for a command.
// Precedence: --file flag → project config (walk-up) → mdboard.md fallback.
func resolveBoardPath(cmd *cobra.Command) (string, error) {
	// 1. Explicit --file flag always wins
	if boardFile != "" {
		return boardFile, nil
	}

	// 2. Walk up from CWD looking for a project-local .mdboard/config.yaml
	cwd, err := os.Getwd()
	if err == nil {
		if projectDir, pc, _ := config.LoadProject(cwd); pc != nil {
			candidate := filepath.Join(projectDir, pc.Board)
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}

	// 3. Fall back to mdboard.md in CWD
	defaultBoard := "mdboard.md"
	_, err = os.Stat(defaultBoard)
	if err == nil {
		return defaultBoard, nil
	}
	if os.IsNotExist(err) {
		return "", fmt.Errorf("board file '%s' not found — run: mdboard new \"<title>\"", defaultBoard)
	}
	return "", err
}

// joinArgs joins CLI args with spaces (handles multi-word card titles)
func joinArgs(args []string) string {
	return strings.Join(args, " ")
}

// resolveCard finds a card by index (-t task, -c col) or by text from args.
// taskIdx is 1-based; 0 means not set. colSpec is the column name/index for scoping.
func resolveCard(b *board.Board, taskIdx int, colSpec string, args []string) (*board.Card, *board.Column, int, error) {
	if taskIdx > 0 {
		if colSpec == "" {
			return nil, nil, -1, fmt.Errorf("--task/-t requires --col/-c to specify the column")
		}
		col, err := board.FindColumn(b, colSpec)
		if err != nil {
			return nil, nil, -1, err
		}
		card, idx, err := board.FindCardByIndex(col, taskIdx)
		return card, col, idx, err
	}
	return board.FindCard(b, joinArgs(args))
}
