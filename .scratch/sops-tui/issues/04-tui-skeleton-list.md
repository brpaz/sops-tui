# 04 — TUI skeleton: flat list, vim nav, manual refresh

**What to build:** The first real, launchable TUI. Running `sops-tui [path]` starts a BubbleTea program that checks `sops` is available (erroring clearly at startup if not), scans the resolved path via the `secrets` package's discovery function, and renders a single flat table of matched files with a status column (Encrypted/Plaintext). Supports vim-style navigation (`j`/`k` move, `gg`/`G` jump top/bottom) and a manual refresh (`r`, re-scans and redraws).

**Blocked by:** 01, 02, 03

**Status:** ready-for-agent

- [ ] Launching in a directory with a mix of files shows a table with path and status columns, matching what the `secrets` discovery function returns
- [ ] `j`/`k` move the selection down/up one row; `gg`/`G` jump to first/last row
- [ ] `r` re-scans the root and redraws the table, reflecting any files changed on disk since launch
- [ ] If `sops` is not on `PATH`, the program exits (or shows a blocking error) before rendering the table, with a message naming the problem
- [ ] Manual/exploratory verification: run the built binary against a real directory tree and confirm the above interactively (BubbleTea `Update` transitions are not asserted via rendered-output snapshot tests in this ticket)
