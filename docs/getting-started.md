# Getting Started

## Prerequisites

- [sops](https://github.com/getsops/sops) installed and available on `PATH`
- A `.sops.yaml` with your creation rules and keys (age, PGP, or a cloud KMS) for the files you
  want to encrypt — sops-tui doesn't manage keys itself, it just drives the `sops` binary you
  already have configured

## Install

Quick install (Linux/macOS):

```bash
curl -fsSL https://raw.githubusercontent.com/brpaz/sops-tui/main/install.sh | sh
```

Downloads the latest release for your OS/arch, verifies its checksum, and installs it to
`/usr/local/bin` (or `~/.local/bin` if that's not writable). See
[`install.sh`](https://github.com/brpaz/sops-tui/blob/main/install.sh) for the `VERSION` and
`INSTALL_DIR` overrides.

Or download a release manually from [GitHub Releases](https://github.com/brpaz/sops-tui/releases)
— pick the archive for your OS/arch, extract it, and put `sops-tui` on your `PATH`.

Or via `go install`:

```bash
go install github.com/brpaz/sops-tui/cmd/sops-tui@latest
```

If you use [devenv](https://devenv.sh/), run this **outside** the devenv shell, or override
`GOBIN` (e.g. `GOBIN=$HOME/go/bin go install ...`) — devenv points `GOPATH` at its own
project-local directory, so a plain `go install` from inside the shell won't land on your normal
`PATH`.

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
