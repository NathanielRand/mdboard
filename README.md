# mdboard

Markdown-based kanban boards for the terminal. Each `.md` file is a board. Cards live under `##` column headings as `###` card titles. Metadata is stored in HTML comments so files stay fully human-readable and git-diffable.

## Install

```bash
git clone https://github.com/NathanielRand/mdboard
cd mdboard
chmod +x install.sh
./install.sh
```

Requires Go 1.22+.

## Quick Start

```bash
# Set your GitHub username once
mdboard config set github_user <USERNAME>

# Create a new board
mdboard new "Project Roadmap"

# Add cards
mdboard add "Design the API"
mdboard add "Build the frontend" --col "In Progress"

# Claim a card
mdboard claim "Design the API"

# Move a card
mdboard move "Design the API" "In Progress"

# View as interactive TUI
mdboard view

# Print text summary
mdboard status

# List all boards in current directory
mdboard list
```

## Commands

| Command | Description |
|---|---|
| `mdboard new <title>` | Scaffold a new board `.md` file |
| `mdboard add <title>` | Add a card (defaults to first column) |
| `mdboard add <title> --col <column>` | Add a card to a specific column |
| `mdboard move <title> <column>` | Move a card to a different column |
| `mdboard shift <title> <up|down>` | Shift a card up or down within its column |
| `mdboard update <title>` | Edit a card's title or body (`--title`, `--body`) |
| `mdboard claim <title>` | Claim a card with your GitHub username |
| `mdboard claim <title> --user <name>` | Claim with a specific username |
| `mdboard unclaim <title>` | Remove claim from a card |
| `mdboard remove <title>` | Delete a card |
| `mdboard view` | Interactive TUI board viewer |
| `mdboard status` | Text summary of the board |
| `mdboard list` | List all `.md` boards in current directory |
| `mdboard config show` | Show current config |
| `mdboard config set <key> <value>` | Update a config value |

All commands support `--file <path>` to target a specific board when multiple `.md` files exist in the current directory. Defaults to `mdboard.md`.

## TUI Controls

| Key | Action |
|---|---|
| `←` / `a` | Previous column |
| `→` / `d` | Next column |
| `↑` / `w` | Previous card |
| `↓` / `s` | Next card |
| `A` / `D` | Move card left/right across columns |
| `W` / `S` | Shift card up/down within its column |
| `e` / `Enter`| Edit the selected card |
| `x` / `Del` | Delete the selected card |
| `q` / `Esc` | Quit |

## Board File Format

Boards are plain markdown files. You can edit them by hand at any time:

```markdown
---
board: Project Roadmap
---

## 📋 Backlog

### Design the API
<!-- status: backlog | created: 2026-05-28 -->
- Define endpoints
- Write OpenAPI spec

## 🔄 In Progress

### Build the frontend
<!-- status: in-progress | user: nrand | claimed: 2026-05-28 -->
- SvelteKit setup
- Connect to backend

## 🧪 Testing

## ✅ Done
```

## Config

Config is stored at `~/.mdboard/config.yaml`:

```yaml
github_user: <USERNAME>
default_columns:
  - Backlog
  - In Progress
  - Testing
  - Done
```

Customize `default_columns` to change what columns `mdboard new` scaffolds.

## Card Matching

Card titles are matched by **case-insensitive partial string**. So if your card is titled `"Fix authentication token expiry"`, you can refer to it as:

```bash
mdboard claim "auth token"
mdboard move "token expiry" Done
```

If the partial match is ambiguous, mdboard will list all matches and ask you to be more specific.
