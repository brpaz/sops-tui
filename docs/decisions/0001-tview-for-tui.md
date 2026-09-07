# Use tview instead of bubbletea for the TUI

## Context and Problem Statement

`internal/tui` was originally built on `charm.land/bubbletea/v2` (Elm-architecture Update/View message loop) plus `bubbles/v2` (table, textinput) and `lipgloss/v2`. The user asked to replace this stack with `rivo/tview` (widget-based, backed by `gdamore/tcell/v2`).

## Considered Options

* Keep bubbletea/bubbles/lipgloss
* Switch to tview/tcell

## Decision Outcome

Chosen option: "Switch to tview/tcell" — this was a direct request, not an evaluation between frameworks on technical merit.

The whole of `internal/tui` was rewritten:

* `tview.Application.SetInputCapture` became the single central key dispatcher (`App.handleKey`), taking the place of bubbletea's `Update(msg)`. It keeps the same priority order the old code had (fatal error > help overlay > confirm-decrypt > content pane > filter/command input > normal list navigation), so almost every existing test translated near line-for-line — swap `tea.KeyPressMsg{Text: "x"}` for a constructed `*tcell.EventKey` and call `handleKey` directly instead of `Update`.
* `tview.Pages` (top-level, plus a nested one for the footer) replaced bubbletea's single `View()` string-building for switching between the list, help, confirm, pane, and error screens.
* The suspend-for-edit action (`E`) got simpler: `tview.Application.Suspend` runs its callback synchronously, so `sops <file>` now runs inline in `startEdit` with no async completion message (`editDoneMsg` is gone, and so is the `editingPath` field that tracked an in-flight edit).
* `internal/secrets` (SOPS shelling, file scanning) needed no changes — it was already framework-agnostic.

### Consequences

* Good, because the centralized-dispatcher shape carried over almost unchanged, so the framework swap didn't also become a logic rewrite.
* Good, because the edit action is simpler and easier to reason about now that it's synchronous.
* Bad, because `Application.Suspend` and `Application.Stop` are no-ops before `Run` has initialized a real screen, so two collaborators had to become test-only injection seams on `App` (`suspend func(f func())`, `edit func(path string) error`) purely to keep the edit action unit-testable without a live terminal.
