Status: ready-for-agent

# sops-tui

## Problem Statement

Working with SOPS-encrypted secret files today means dropping to the shell for every operation: `sops -d file.yaml` to peek at a value, `sops file.yaml` to edit, `sops -e -i file.yaml` to encrypt a new plaintext file — each a separate command, each requiring the operator to already know the exact file path and remember its current encryption state. There's no single view of "what secret files exist in this project, and which of them are currently protected." Users familiar with `k9s` for Kubernetes resources have no equivalent for SOPS-managed files: a live, navigable list they can act on directly.

## Solution

A terminal UI, modeled on `k9s`'s interaction style, that recursively scans a directory for YAML/JSON files, lists them in a single filterable table annotated with their SOPS encryption status, and lets the user view, edit, encrypt, or decrypt the selected file with a single keypress — delegating all actual cryptographic work to the user's existing `sops` binary and `.sops.yaml` configuration rather than reimplementing any of it.

## User Stories

1. As an engineer managing SOPS secrets, I want to launch a TUI in a project directory, so that I immediately see every YAML/JSON file relevant to secrets management without hunting through the tree manually.
2. As an engineer, I want the file list to show each file's encryption status (Encrypted / Plaintext), so that I can spot at a glance which files still need to be protected.
3. As an engineer, I want to navigate the file list with vim-style keys (`j`/`k`, `gg`/`G`), so that the tool feels familiar if I already use `k9s`, `lazygit`, or similar terminal tools.
4. As an engineer, I want to filter the file list with `/`, so that I can quickly narrow a large tree down to the file I'm looking for.
5. As an engineer, I want a `:` command mode, so that I have a k9s-consistent way to trigger actions beyond single-key bindings.
6. As an engineer, I want a `?` help overlay, so that I can discover available keybindings without leaving the tool or reading external docs.
7. As an engineer, I want to press a key to view a selected encrypted file's decrypted contents in a read-only pane, so that I can inspect a secret's value without ever writing plaintext to disk.
8. As an engineer, I want to press a key to edit a selected encrypted file, so that the TUI suspends, hands me `sops <file>` directly (my normal `$EDITOR` flow, re-encrypting on save), and resumes when I'm done.
9. As an engineer, I want to press a key to decrypt a selected file in place, so that I can produce a plaintext copy on disk when I explicitly need one (e.g. for a one-off local task).
10. As an engineer, I want a confirmation prompt before an in-place decrypt executes, so that I don't accidentally leave plaintext secrets on disk from a stray keypress.
11. As an engineer, I want to press a key to encrypt a selected plaintext file, so that I can protect a new secret file without leaving the TUI.
12. As an engineer, when a plaintext file doesn't match any `.sops.yaml` creation rule, I want a clear in-TUI error naming the file and pointing at `.sops.yaml`, so that I know exactly why encryption was refused and how to fix it.
13. As an engineer, I want the list to only ever reflect `.sops.yaml`-driven key selection (no ad-hoc key entry in the TUI), so that key management stays centralized in the config file everyone on the team already uses.
14. As an engineer, I want the tool to scan recursively from my current working directory by default, or from an explicit path I pass as an argument, so that I can point it at any project without extra configuration.
15. As an engineer, I want common non-secret directories (`.git`, `node_modules`, `vendor`) excluded from the scan automatically, so that the list isn't cluttered with irrelevant files.
16. As an engineer, I want the tool to work with zero config file of its own, so that I can run it in any project immediately without a setup step.
17. As an engineer, I want to trigger a manual refresh (`r`) of the file list, so that I can pick up changes made outside the TUI (e.g. `git pull`, another terminal).
18. As an engineer, I want the file list to auto-refresh immediately after any encrypt/decrypt/edit action I perform inside the TUI, so that the status column always reflects what I just did without a manual step.
19. As an engineer, I want the tool to only ever call out to the `sops` binary already on my `PATH` (never bundle or reimplement SOPS's crypto), so that behavior always matches whatever `sops` version and configuration I already trust in this project.
20. As a maintainer, I want the core file-scan + sops-action logic exposed as a single well-defined package, so that it can be tested end-to-end against a real `sops` binary independent of the terminal UI layer.

## Implementation Decisions

- **Language/stack**: Go, using BubbleTea for the TUI layer. Scaffolded from the `brpaz/copier-go` template (`project_type: cli`), module `github.com/brpaz/sops-tui`.
- **SOPS integration**: shell out to the `sops` binary via subprocess (`exec.Command`) for every cryptographic operation — encrypt, decrypt, view, edit. No embedding of `go.mozilla.org/sops/v3` or any other SOPS library. `sops` must be resolvable on `PATH`; absence is a startup-time error.
- **Core seam**: a single `secrets` package is the boundary between the TUI and the outside world, exposing operations equivalent to:
  - `List(root string) ([]FileEntry, error)` — recursive scan, filtered to `.yaml`/`.yml`/`.json`, excluding `.git`, `node_modules`, `vendor`; each `FileEntry` carries path and encryption status.
  - `IsEncrypted(path string) (bool, error)` — status detection by checking for SOPS metadata in the parsed file (a `sops` top-level key), not by filename convention.
  - `View(path string) (string, error)` — returns decrypted content via `sops -d`, does not touch disk.
  - `Edit(path string) error` — invokes `sops <path>` as a foreground subprocess with inherited stdio, for the TUI to call after suspending itself.
  - `Decrypt(path string) error` — `sops -d -i <path>`, decrypts in place.
  - `Encrypt(path string) error` — `sops -e -i <path>`, relies entirely on `.sops.yaml` creation rules; surfaces sops's own error (e.g. no matching creation rule) back to the caller unmodified, for the TUI to display naming the file.
  - The BubbleTea `Update` loop is a thin consumer of this package: it calls these functions in response to key events and updates its own displayed state from the results. It has no independent seam of its own.
- **Navigation model**: single flat, filterable table of all matched files (relative path shown), not a directory drill-down. Matches k9s's resource-table idiom.
- **Keybindings**: vim-style — `j`/`k` move, `gg`/`G` jump to top/bottom, `/` filter, `:` command mode, `?` help overlay — mirroring k9s directly.
- **Actions and their keys** (single-file only, no multi-select/bulk actions in this scope):
  - view (read-only decrypted pane): safe, non-destructive default action on the selected row.
  - edit: suspends the TUI process, execs `sops <file>` with inherited terminal, resumes and refreshes the list on return.
  - decrypt-in-place: requires an explicit confirmation prompt before executing.
  - encrypt: acts on a plaintext row; on failure (no `.sops.yaml` match), shows the file path and the underlying `sops` error in the TUI rather than a generic failure message.
- **Discovery/config**: zero-config. Scan root is cwd by default, overridable via a single optional CLI positional argument (`sops-tui [path]`). Ignore list (`.git`, `node_modules`, `vendor`) is hardcoded, not user-configurable in this scope.
- **File format scope**: YAML and JSON only. Dotenv and binary SOPS formats are out of scope (see Out of Scope).
- **Refresh**: manual only, bound to `r`, plus an automatic refresh triggered by the TUI itself immediately after any encrypt/decrypt/edit action completes. No filesystem watcher (no `fsnotify` or equivalent dependency).
- **Bulk actions**: none. Every action operates on exactly one selected file.

## Testing Decisions

- Tests must exercise observable behavior of the `secrets` package's public functions (`List`, `IsEncrypted`, `View`, `Decrypt`, `Encrypt`; `Edit` is harder to assert automatically since it hands off to an interactive subprocess — cover it with a thin integration test that stubs `$EDITOR` to a non-interactive script, e.g. one that appends a fixed line, and asserts the file round-trips through encrypt/decrypt correctly). Do not assert on the literal `exec.Command` arguments passed to `sops` — that tests wiring, not behavior.
- Tests run against the **real `sops` binary** with a **generated, throwaway `age` keypair** per the earlier decision to avoid mocking the exec boundary — this is the only way to catch real drift in `sops`'s CLI behavior. No `testcontainers` needed; `age`/`sops` are local binaries, not services.
- Test setup: generate an age keypair once (`age-keygen`), write a scratch `.sops.yaml` pointing at its public key into a temp directory per test, then exercise `secrets` package functions against fixture YAML/JSON files placed in that temp dir.
- Prior art: none yet in this repo (freshly scaffolded) — follow the `golang-testing` skill's table-driven conventions and the project's existing `.golangci.yml`/`lefthook.yml` setup for style enforcement. CI (`.github/workflows/ci.yml`, already scaffolded) needs `sops` and `age` installed as steps before the test job runs.
- The BubbleTea `Update` loop itself is not unit-tested for rendered output in this scope — its only logic is dispatching to the `secrets` package, which is what's actually tested.

## Out of Scope

- Dotenv (`.env`) and binary SOPS file formats.
- Multi-select / bulk encrypt or decrypt of several files at once.
- Ad-hoc key entry in the TUI (age recipient, PGP fingerprint, KMS ARN) when no `.sops.yaml` rule matches — user must fix `.sops.yaml` instead.
- A dedicated app config file (themes, custom ignore lists, custom scan roots beyond the single CLI argument).
- Filesystem-watcher-based auto-refresh.
- Embedding the SOPS Go library instead of shelling out to the binary.

## Further Notes

- This spec was produced via a `/grilling` session (see conversation history) that resolved every open design branch before scaffolding; no further product-decision interview should be needed before implementation starts.
- Project scaffolding is already done: `brpaz/copier-go` template applied with `project_type: cli`, producing the standard `cmd/`, `internal/app`, `internal/commands/{root,hello}` cobra-style skeleton, plus GoReleaser, GitHub Actions CI, devenv, and Taskfile. The `hello` sample command should be removed once the real `secrets`/TUI commands land.
- Repo issue tracker is local markdown (`.scratch/`, no GitHub remote pushed yet) — see `docs/agents/issue-tracker.md`. Domain docs are single-context with ADRs at `docs/decisions/` — see `docs/agents/domain.md`.
- `triage` skill is not installed for this repo, so no triage-label vocabulary exists beyond the single `ready-for-agent` status this spec is marked with.
