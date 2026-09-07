# Getting Started

## Prerequisites

- [sops](https://github.com/getsops/sops) installed and available on `PATH`
- A `.sops.yaml` with your creation rules and keys (age, PGP, or a cloud KMS) for the files you
  want to encrypt — sops-tui doesn't manage keys itself, it just drives the `sops` binary you
  already have configured

## Install

```bash
go install github.com/brpaz/sops-tui/cmd/sops-tui@latest
```

Or build from source:

```bash
git clone https://github.com/brpaz/sops-tui.git
cd sops-tui
go install ./cmd/sops-tui
```

A Docker image is also published — see the [source repo](https://github.com/brpaz/sops-tui) for
the current image name and tags.

## Run it

```bash
sops-tui [path]
```

`path` is the directory to scan for SOPS-managed files; it defaults to the current directory.
sops-tui walks it recursively (respecting `.gitignore` when the directory is a git work tree),
picks up every `.yaml`/`.yml`/`.json` file, and classifies each one as encrypted, plaintext, or
unknown by checking for a top-level `sops` key.

By default you'll land on the **encrypted** view. Press `Tab` to cycle to **plaintext** if you're
looking for files to encrypt, or **all** to see everything at once. From there, `j`/`k` to move,
`v`/`Enter` to view a file's content, `e`/`d` to encrypt/decrypt, `E` to edit. See
[Usage](usage.md) for the full walkthrough and [Keybindings](keybindings.md) for the complete key
reference.
