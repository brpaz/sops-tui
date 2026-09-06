# 04 — TUI skeleton: flat list, vim nav, manual refresh

**What to build:** The first real, launchable TUI. Running `sops-tui [path]` starts a BubbleTea program that checks `sops` is available (erroring clearly at startup if not), scans the resolved path via the `secrets` package's discovery function, and renders a single flat table of matched files with a status column (Encrypted/Plaintext). Supports vim-style navigation (`j`/`k` move, `gg`/`G` jump top/bottom) and a manual refresh (`r`, re-scans and redraws).

**Blocked by:** 01, 02, 03

**Status:** done

- [x] Launching in a directory with a mix of files shows a table with path and status columns, matching what the `secrets` discovery function returns
- [x] `j`/`k` move the selection down/up one row; `gg`/`G` jump to first/last row (via bubbles/v2 table's default keymap: single `g`/`home` and `G`/`end` for top/bottom, `j`/`k`/arrows for line up/down)
- [x] `r` re-scans the root and redraws the table, reflecting any files changed on disk since launch
- [x] If `sops` is not on `PATH`, the program exits (or shows a blocking error) before rendering the table, with a message naming the problem
- [x] Manual/exploratory verification — with caveats: this sandbox has no `/dev/tty`, so the real interactive BubbleTea program couldn't be launched directly. Verified instead with a Python-pty-driven smoke test of the built binary: it enters the alt screen, accepts keypresses, and exits cleanly (code 0) on `q`; separately confirmed the built binary (not just the Go test) exits code 1 with the PATH error, unrendered, when `sops` is stripped from `PATH`. Functional behavior (table contents from a real scan, refresh re-scanning, quit command) is covered by unit tests against `Model.Update`/`View` instead of a live terminal.

Uses `charm.land/bubbletea/v2` (v2.0.9), `charm.land/bubbles/v2` (v2.2.1) — Charm's v2 modules moved off the `github.com/charmbracelet/...` import path to `charm.land/...`.
