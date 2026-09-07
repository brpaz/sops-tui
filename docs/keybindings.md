# Keybindings

## Navigation

| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` / `Home` | Jump to top |
| `G` / `End` | Jump to bottom |

## Actions

| Key | Action |
|---|---|
| `v` / `Enter` | View content, read-only (decrypts encrypted files; reads plaintext directly) |
| `e` | Encrypt the selected plaintext file (asks to confirm) — only shown when the selection is plaintext |
| `d` | Decrypt the selected file in place (asks to confirm) — only shown when the selection is encrypted |
| `E` | Edit (suspends to `sops <file>`) |
| `r` | Refresh (re-scan the directory) |

## Views and filtering

| Key | Action |
|---|---|
| `Tab` | Cycle view: encrypted → plaintext → all → encrypted |
| `/` | Filter by path (`Enter` confirms, `Esc` cancels) |
| `:` | Command mode — `view`, `encrypt`, `decrypt`, `edit`, `refresh` |

## Overlays and quitting

| Key | Action |
|---|---|
| `?` | Toggle the keybinding help overlay |
| `Esc` / `q` | Close the current overlay (help, content pane, error pane) and return to the list |
| `q` / `Ctrl+C` | Quit (from the list) |

The header's middle column always shows a short, selection-aware subset of these — whichever of
`e`/`d` currently applies, plus `v`/`Enter`, `E`, `Tab`, and `?` — mirroring what the footer shows
in full.
