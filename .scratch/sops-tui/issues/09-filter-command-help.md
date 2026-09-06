# 09 — Filter (`/`), command mode (`:`), help overlay (`?`)

**What to build:** k9s-consistency navigation polish layered onto the working list and actions: `/` opens a filter input that narrows the visible rows by path substring/fuzzy match; `:` opens a command input for triggering the same actions by name instead of a raw keypress; `?` opens a help overlay listing all available keybindings and what they do.

**Blocked by:** 04

**Status:** done

- [x] `/` opens a filter input; typing narrows the table to matching paths live (case-insensitive substring); clearing/cancelling (`esc`) restores the full list. A confirmed filter (`enter`) persists across manual/auto refreshes.
- [x] `:` opens a command input accepting `view`, `encrypt`, `decrypt`, `edit`, `refresh` (case-insensitive), applied to the currently selected row; an unrecognized name opens an error pane naming it
- [x] `?` opens an overlay listing every bound key and a one-line description, including everything added by tickets 05-08
- [x] All three overlays/inputs can be dismissed (`esc`, and `q` for the pane/help) without side effects, returning to the list at the same selection

Note: while a filter or command input is focused, list-level single-key bindings (including `q`) are correctly *not* intercepted — they're ordinary typed characters instead, confirmed via a pty smoke test of the real binary (typing "q" while filtering inserts it into the filter text rather than quitting).
