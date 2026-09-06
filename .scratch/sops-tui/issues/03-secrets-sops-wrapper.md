# 03 — secrets: sops binary wrapper + availability check

**What to build:** The `secrets` package's cryptographic primitives, each a thin wrapper shelling out to the `sops` binary: `View` (decrypt to string, no disk write), `Decrypt` (in-place), `Encrypt` (in-place, relies on `.sops.yaml` creation rules), and `Edit` (foreground subprocess with inherited stdio, for the caller to invoke after suspending itself). Also a `CheckAvailable` function that reports whether `sops` is resolvable on `PATH`.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] `View(path)` returns decrypted content as a string; the file on disk is unchanged
- [ ] `Decrypt(path)` decrypts the file in place; a subsequent status check reports it Plaintext
- [ ] `Encrypt(path)` encrypts a plaintext file in place per a test-local `.sops.yaml`; a subsequent status check reports it Encrypted
- [ ] `Encrypt(path)` on a file with no matching `.sops.yaml` creation rule returns an error naming the file, unmodified from what `sops` itself reports
- [ ] `Edit(path)` runs `sops <path>` as a foreground subprocess with inherited stdio; a non-interactive `$EDITOR` stub is used in tests to verify the file round-trips through the edit (still valid SOPS-encrypted content afterward)
- [ ] `CheckAvailable()` returns no error when `sops` is on `PATH`, and a clear error when it isn't
- [ ] All tests run against a real `sops` binary and a generated, throwaway `age` keypair — no mocking of `exec.Command`
