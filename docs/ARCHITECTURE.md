# Architecture

sops-tui is a single-binary CLI: a thin `urfave/cli` command that resolves a scan root and hands
it to a `rivo/tview` terminal application. There's no server, no persistence beyond the files on
disk, and no dependency on sops as a library — every encrypt/decrypt/edit/view operation shells
out to the real `sops` binary, so behavior always matches what running `sops` by hand would do.

## Package layout

```
cmd/sops-tui/          entrypoint: wires build-time version info, calls internal/app
internal/app/          composition root: builds the root command, injects tui.Run
internal/commands/root/ the urfave/cli root command (positional `path` arg, --version)
internal/secrets/      sops shelling + directory scanning; framework-agnostic
internal/tui/          the tview application
```

Dependencies point inward: `cmd` depends on `internal/app`, which depends on
`internal/commands/root` and `internal/tui`, which both depend on `internal/secrets`.
`internal/secrets` depends on nothing else in this repo.

### `internal/secrets`

- `list.go` — `List(root)` walks `root` (respecting `.gitignore` via `git ls-files` when
  available, falling back to a plain walk otherwise), filters to `.yaml`/`.yml`/`.json`, and
  classifies each file's `Status` (`Encrypted`/`Plaintext`/`Unknown`) by checking for a top-level
  `sops` key — no shelling out needed just to determine status.
- `sops.go` — `View`, `Encrypt`, `Decrypt`, and `Edit`/`EditCommand` each shell out to the `sops`
  binary. `runSopsOnFile` always runs from the target file's own directory, since sops discovers
  `.sops.yaml` by walking up from the process's working directory.

### `internal/tui`

Built on `tview.Application`, with `App.handleKey` (in `app.go`) as the single input entry point
installed via `SetInputCapture`. It mirrors the priority order the UI renders in: a fatal error
overrides everything, then the help overlay, then a pending encrypt/decrypt confirmation, then a
content pane, and finally normal list navigation.

| File | Responsibility |
|---|---|
| `app.go` | Application setup, layout (header/filter bar/table/footer), rendering, key dispatch |
| `actions.go` | Encrypt/decrypt (each behind its own confirmation), edit, `:` commands |
| `pane.go` | Full-screen content panes (view/error) and table selection helpers |
| `filter.go` | `/` filter and `:` command input mode |
| `help.go` | The `?` keybinding overlay |
| `run.go` | `Run(ctx, root)`: checks sops is on `PATH`, builds the `App`, runs it |

The table itself has no visible status column — `Tab` cycles three views (encrypted / plaintext /
all) that filter by `secrets.Status` directly, so which bucket a row belongs to is implied by
which view you're looking at. A row's status still travels with it via `tview.TableCell`'s
`SetReference`, since actions like encrypt/decrypt need to know it without a visible column to
read from.

See [decisions/](decisions/) for the reasoning behind picking tview over other TUI frameworks.

## Testing

Every `tui` action is tested by constructing a `*tcell.EventKey` and calling `App.handleKey`
directly — no real terminal or `Application.Run` needed. Two collaborators exist purely as
test seams: `App.suspend` (wraps `Application.Suspend`, which is a no-op before `Run` has
initialized a screen) and `App.edit` (wraps `secrets.Edit`, so tests can simulate an editor's
outcome without shelling out to a real one). `internal/secrets`' own tests shell out to the real
`sops` and `age-keygen` binaries against a throwaway keypair, since its whole job is getting that
interaction right.
