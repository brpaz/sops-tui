# 05 — View action (read-only decrypted pane)

**What to build:** Selecting an encrypted file and triggering the view action opens a read-only pane showing its decrypted content, using the `secrets` package's `View` function. The file on disk is never modified. A keypress returns to the list.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Triggering view on an encrypted row shows its decrypted content in-pane
- [ ] The underlying file's contents on disk are byte-for-byte unchanged after viewing
- [ ] Returning from the pane goes back to the list at the same selection
- [ ] Triggering view on a file that fails to decrypt (e.g. no access to the key) shows the `sops` error in-pane instead of crashing the TUI
