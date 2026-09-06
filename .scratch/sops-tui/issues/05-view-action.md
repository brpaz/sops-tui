# 05 — View action (read-only decrypted pane)

**What to build:** Selecting an encrypted file and triggering the view action opens a read-only pane showing its decrypted content, using the `secrets` package's `View` function. The file on disk is never modified. A keypress returns to the list.

**Blocked by:** 04

**Status:** done

- [x] Triggering view on an encrypted row shows its decrypted content in-pane (bound to `v` and `enter`)
- [x] The underlying file's contents on disk are byte-for-byte unchanged after viewing
- [x] Returning from the pane goes back to the list at the same selection (bound to `esc` and `q`; `q` closes the pane rather than quitting the whole app while a pane is open, `ctrl+c` always quits)
- [x] Triggering view on a file that fails to decrypt (e.g. no access to the key) shows the `sops` error in-pane instead of crashing the TUI
