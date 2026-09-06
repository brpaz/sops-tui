# 03 — secrets: sops binary wrapper + availability check

**What to build:** The `secrets` package's cryptographic primitives, each a thin wrapper shelling out to the `sops` binary: `View` (decrypt to string, no disk write), `Decrypt` (in-place), `Encrypt` (in-place, relies on `.sops.yaml` creation rules), and `Edit` (foreground subprocess with inherited stdio, for the caller to invoke after suspending itself). Also a `CheckAvailable` function that reports whether `sops` is resolvable on `PATH`.

**Blocked by:** None — can start immediately.

**Status:** done

- [x] `View(path)` returns decrypted content as a string; the file on disk is unchanged
- [x] `Decrypt(path)` decrypts the file in place; a subsequent status check reports it Plaintext
- [x] `Encrypt(path)` encrypts a plaintext file in place per a test-local `.sops.yaml`; a subsequent status check reports it Encrypted
- [x] `Encrypt(path)` on a file with no matching `.sops.yaml` creation rule returns an error naming the file, unmodified from what `sops` itself reports
- [x] `Edit(path)` runs `sops <path>` as a foreground subprocess with inherited stdio; a non-interactive `$EDITOR` stub is used in tests to verify the file round-trips through the edit (still valid SOPS-encrypted content afterward)
- [x] `CheckAvailable()` returns no error when `sops` is on `PATH`, and a clear error when it isn't
- [x] All tests run against a real `sops` binary and a generated, throwaway `age` keypair — no mocking of `exec.Command`

Note: `sops`'s own `.sops.yaml` discovery is relative to the process's working directory, not just the target file's absolute path — every wrapper function here runs `sops` with `cmd.Dir` set to the target file's directory and passes only its basename, so discovery works correctly regardless of where `sops-tui` itself is launched from. The CI test job (`.github/workflows/ci.yml`) now installs `sops` v3.13.3 and `age` v1.3.2 before running tests; two pre-existing scaffold YAML bugs in that file (a missing newline that merged two job definitions, and mis-indented steps in `build-binary`) were also fixed since they left the file un-parseable.
