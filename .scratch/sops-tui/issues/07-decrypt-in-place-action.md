# 07 — Decrypt-in-place action

**What to build:** Selecting an encrypted file and triggering the decrypt-in-place action shows a confirmation prompt first; on confirming, it calls the `secrets` package's `Decrypt` function, then the list auto-refreshes and the row's status flips to Plaintext. Cancelling the prompt leaves the file untouched.

**Blocked by:** 04

**Status:** done

- [x] Triggering decrypt-in-place on an encrypted row shows a confirmation prompt before anything happens on disk (bound to `d`, confirm `y`, cancel `n`/`esc`)
- [x] Confirming decrypts the file in place; the list refreshes and shows it as Plaintext without a manual `r`
- [x] Cancelling the prompt leaves the file's contents and status unchanged
- [x] Triggering decrypt-in-place on an already-plaintext row is a no-op
