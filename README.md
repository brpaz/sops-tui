# sops-tui

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/brpaz/sops-tui?style=for-the-badge)
![Go Report Card](https://goreportcard.com/badge/github.com/brpaz/sops-tui?style=for-the-badge)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/brpaz/sops-tui/ci.yml?branch=main&style=for-the-badge)](https://github.com/brpaz/sops-tui/actions)
[![License](https://img.shields.io/github/license/brpaz/sops-tui?style=for-the-badge)](./LICENSE)

> A k9s-style terminal UI for browsing, encrypting, and decrypting SOPS-protected secret files

## ✨ Features

- Flat, k9s-style table of every `.yaml`/`.yml`/`.json` file under a directory, with vim-style navigation
- Tab-cycled views: **encrypted**, **plaintext**, and **all**, so encrypt/decrypt candidates are never mixed up
- View decrypted content read-only, without ever writing plaintext to disk
- Encrypt/decrypt in place, each behind its own confirmation screen
- Edit via `sops <file>` (suspends the TUI, runs your `$EDITOR` through sops, re-encrypts on save)
- Live filtering by path and a `:` command mode
- Catppuccin Macchiato theme

## 🚀 Getting Started

### Prerequisites

- [sops](https://github.com/getsops/sops) installed and available on `PATH`
- A `.sops.yaml` with your creation rules and keys (age, PGP, or a cloud KMS) for the files you want to encrypt

### Installation

```bash
go install github.com/brpaz/sops-tui/cmd/sops-tui@latest
```

Or build from source:

```bash
git clone https://github.com/brpaz/sops-tui.git
cd sops-tui
go install ./cmd/sops-tui
```

## Usage

```bash
sops-tui [path]
```

`path` is the directory to scan for SOPS-managed files; it defaults to the current directory.

### Keybindings

| Key | Action |
|---|---|
| `j`/`k`, `↓`/`↑` | Move down / up |
| `g`/`G`, `Home`/`End` | Jump to top / bottom |
| `v` / `Enter` | View content, read-only (decrypts encrypted files; reads plaintext directly) |
| `e` | Encrypt the selected plaintext file (asks to confirm) |
| `d` | Decrypt the selected file in place (asks to confirm) |
| `E` | Edit (suspends to `sops <file>`) |
| `Tab` | Cycle view: encrypted / plaintext / all |
| `/` | Filter by path |
| `:` | Command mode (`view`, `encrypt`, `decrypt`, `edit`, `refresh`) |
| `r` | Refresh (re-scan) |
| `?` | Help |
| `q` / `Ctrl+C` | Quit |

## 🧱 Tech Stack

- [Go](https://go.dev/) with [urfave/cli](https://github.com/urfave/cli) for the CLI entrypoint
- [rivo/tview](https://github.com/rivo/tview) (backed by [gdamore/tcell](https://github.com/gdamore/tcell)) for the terminal UI
- [sops](https://github.com/getsops/sops) itself, shelled out to for every encrypt/decrypt/view/edit — this project never reimplements sops's own logic
- [Zensical](https://zensical.org/) for the [documentation site](https://brpaz.github.io/sops-tui/), deployed to GitHub Pages
- [draftsman](https://github.com/brpaz/draftsman) for changelog generation from Conventional Commits
- [Renovate](https://docs.renovatebot.com/) for dependency updates
- [devenv](https://devenv.sh/)/Nix for the reproducible dev shell, [GoReleaser](https://goreleaser.com/) for release builds, Docker and GitHub Actions for CI/CD

## 🤝 Contributing

All contributions are welcome. Please check [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 🫶 Support

If you find this project helpful and would like to support its development, there are a few ways you can contribute:

[![Sponsor me on GitHub](https://img.shields.io/badge/Sponsor-%E2%9D%A4-%23db61a2.svg?&logo=github&logoColor=red&&style=for-the-badge&labelColor=white)](https://github.com/sponsors/brpaz)

<a href="https://www.buymeacoffee.com/Z1Bu6asGV" target="_blank"><img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: auto !important;width: auto !important;" ></a>

## 👱 Contributors

- [Bruno Paz](https://brunopaz.dev) - Creator and maintainer

## 🤖 AI Usage

This project was built in close collaboration with [Claude Code](https://claude.com/claude-code):
feature implementation, UI iteration, CI/tooling fixes, and documentation were all done through an
AI pair-programming workflow, directed and reviewed by the maintainer at each step. Every change
was validated against the test suite, linters, and CI before being merged — the same bar as any
other contribution to this repo.

## ❤️ Acknowledgements

## 📃 License

Distributed under the MIT License. See [LICENSE](LICENSE) file for details.

## 📩 Contact

- 📧 **Email**: [bruno@brunopaz.dev](mailto:bruno@brunopaz.dev)
- 🐞 **Issues**: [GitHub Issues](https://github.com/brpaz/sops-tui/issues)
- 🖇️ **Source**: [https://github.com/brpaz/sops-tui](https://github.com/brpaz/sops-tui)
