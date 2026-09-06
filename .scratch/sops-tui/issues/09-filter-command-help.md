# 09 — Filter (`/`), command mode (`:`), help overlay (`?`)

**What to build:** k9s-consistency navigation polish layered onto the working list and actions: `/` opens a filter input that narrows the visible rows by path substring/fuzzy match; `:` opens a command input for triggering the same actions by name instead of a raw keypress; `?` opens a help overlay listing all available keybindings and what they do.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] `/` opens a filter input; typing narrows the table to matching paths live; clearing/cancelling restores the full list
- [ ] `:` opens a command input accepting at least the names of the actions already implemented at the time this ticket lands (view/encrypt/decrypt/edit/refresh), applied to the currently selected row
- [ ] `?` opens an overlay listing every bound key and a one-line description; any key added by tickets 05-08 appears here
- [ ] All three overlays/inputs can be dismissed (e.g. `Esc`) without side effects, returning to the list at the same selection
