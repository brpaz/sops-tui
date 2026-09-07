# Usage

## Views

`Tab` cycles through three views, each filtered by status:

| View | Shows |
|---|---|
| **encrypted** (default) | Files sops-tui detects as already SOPS-encrypted |
| **plaintext** | Files with no sops metadata yet — candidates to encrypt |
| **all** | Everything, regardless of status |

The table's border title reflects the active view and row count, e.g. `files(encrypted)[3]`.

## Viewing content

`v` or `Enter` opens a read-only, full-screen view of the selected file:

- On an **encrypted** row, it runs `sops -d` and shows the decrypted content. The file on disk is
  never modified.
- On a **plaintext** row, it just reads the file directly — sops itself would refuse to decrypt a
  file with no sops metadata, so there's no point shelling out for one.

Press `Esc` or `q` to return to the list.

## Encrypting and decrypting

`e` (on a plaintext row) and `d` (on an encrypted row) each raise a full-screen confirmation
before anything touches disk — a distinct screen per action, naming the file and describing
exactly what's about to happen:

```
Encrypt secret.yaml?
This writes ciphertext to disk using .sops.yaml creation rules.

[y] yes   [n/esc] cancel
```

Press `y` to proceed or `n`/`Esc` to cancel. On success the list refreshes automatically — no
need to press `r`. On failure (e.g. sops finds no matching `.sops.yaml` creation rule for
encrypt), an error pane names the file and shows sops's own reported reason, and the file is left
untouched.

Both keys are only shown in the footer and header hotkey lists when they apply to the currently
selected row — no `e` on an already-encrypted file, no `d` on a plaintext one.

## Editing

`E` suspends the TUI and runs `sops <file>` in the foreground: sops decrypts to a temp file,
launches your `$EDITOR` on it, and re-encrypts on save. The list refreshes once you exit the
editor. Saving without changing anything ("the file has not been modified, exiting.") is treated
as a normal, no-op exit — the list simply refreshes with no error shown.

## Filtering

`/` opens a live filter over the **path** column, narrowing the table as you type. `Enter`
confirms the filter (it survives a manual `r` refresh); `Esc` cancels and restores whatever
filter was active before you opened it.

## Command mode

`:` opens a command prompt that runs one of the row actions by name against the currently
selected row:

- `view` — same as `v`/`Enter`
- `encrypt` — same as `e`
- `decrypt` — same as `d`
- `edit` — same as `E`
- `refresh` — same as `r`

An unrecognized command name shows an error pane naming it, without side effects.

See [Keybindings](keybindings.md) for the complete key reference.
