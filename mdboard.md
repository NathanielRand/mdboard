---
board: MDBoard
---

## 📋 Backlog

### Due dates on cards — set via mdb add / mdb update --due, visual indicator in TUI when overdue
<!-- status: backlog | created: 2026-06-10 -->

### Priority levels (p1/p2/p3) — sortable, shown as badge in card list
<!-- status: backlog | created: 2026-06-10 -->

### Tags/labels — freeform, filterable via mdb list --tag
<!-- status: backlog | created: 2026-06-10 -->

### Card history — track when a card moved columns, mdb log <card>
<!-- status: backlog | created: 2026-06-10 -->

### mdb archive — move all Done cards to a hidden archive section in the .md file
<!-- status: backlog | created: 2026-06-10 -->

### mdb stats — cards per column, avg time in each status, throughput over last 7/30 days
<!-- status: backlog | created: 2026-06-10 -->

### WIP limits — config max_cards per column, warn or block when exceeded
<!-- status: backlog | created: 2026-06-10 -->

### mdb search <query> — fuzzy search cards across all boards in a project
<!-- status: backlog | created: 2026-06-10 -->

### Undo last action (u key in TUI) — in-memory undo stack for move/delete/reorder
<!-- status: backlog | created: 2026-06-10 -->

### Multi-select in TUI — space to mark cards, then bulk move or delete
<!-- status: backlog | created: 2026-06-10 -->

### Column collapse in TUI — tab to toggle hide/show a column
<!-- status: backlog | created: 2026-06-10 -->

### mdb sync github — pull open issues from a repo into Backlog, push Done cards back as closed
<!-- status: backlog | created: 2026-06-10 -->

### Git hook mode — auto-commit board file on every card action with a formatted commit message
<!-- status: backlog | created: 2026-06-10 -->

### mdb export — render the board as styled HTML or markdown report
<!-- status: backlog | created: 2026-06-10 -->

### Board templates — mdb new --template scrum scaffolds sprint-specific columns
<!-- status: backlog | created: 2026-06-10 -->

### mdb upcoming — list all cards with due dates in the next N days across all boards
<!-- status: backlog | created: 2026-06-10 -->

### Shell completions — mdb completion bash/zsh/fish for card title and column tab-complete
<!-- status: backlog | created: 2026-06-10 -->

## 🔄 In Progress

## 🧪 Testing

### update tui uiux v3
<!-- status: testing | created: 2026-06-07 -->

### alt column flag as -c (on move command)
<!-- status: testing | created: 2026-06-10 -->

### column/board min digit number id (-c 1, -c 2)
<!-- status: testing | created: 2026-06-10 -->

### create card action/keybind from tui (n key)
<!-- status: testing | created: 2026-06-10 -->

### reactivness for the tui to update data live sync
<!-- status: testing | created: 2026-06-10 -->

### WASD/arrow keys control mirror; add keybind for all board actions in the TUI
<!-- status: testing | created: 2026-06-09 -->

## 📌 ✅ Done

### creating a new board asks for project name and then creates an .md file with that name. we need a way for us to auto set that project name in the config for that project's root so that when "mdb view or mdboard view" command runs it auto detects the mdboard in that project's root to use instead of searching for all .md files.
<!-- status: done | created: 2026-06-10 -->

### when pusing updates, can we add some type of updater command to fetch new versions
<!-- status: done | created: 2026-06-10 -->

### column/board flag by name should be NON-case sensitive + fuzzy match
<!-- status: done | created: 2026-06-09 -->

### deleting cards needs a confirmation
<!-- status: done | created: 2026-06-10 -->

### update readme with screenshot and AI code assist use-case
<!-- status: done | created: 2026-06-10 -->

### Inline card creation in TUI — type title directly without opening editor for simple cards
<!-- status: done | created: 2026-06-10 -->

### wasd for naivgation only
<!-- status: backlog | created: 2026-06-09 -->

### word wrap to show the full card text
<!-- status: backlog | created: 2026-06-09 -->

