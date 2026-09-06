# 06 — Encrypt action

**What to build:** Selecting a plaintext file and triggering the encrypt action calls the `secrets` package's `Encrypt` function. On success, the list auto-refreshes and the row's status flips to Encrypted. On failure (no matching `.sops.yaml` creation rule), an in-TUI error names the file and shows the underlying `sops` error rather than a generic failure message.

**Blocked by:** 04

**Status:** done

- [x] Triggering encrypt on a plaintext row that matches a `.sops.yaml` rule encrypts it in place; the list refreshes and shows it as Encrypted without a manual `r` (bound to `e`)
- [x] Triggering encrypt on a plaintext row with no matching `.sops.yaml` rule shows an error naming the file and the `sops`-reported reason; the file is left untouched
- [x] Triggering encrypt on an already-encrypted row is a no-op
