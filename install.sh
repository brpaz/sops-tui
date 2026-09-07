#!/usr/bin/env sh
# Installs the latest sops-tui release for the current OS/arch.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/brpaz/sops-tui/main/install.sh | sh
#
# Env overrides:
#   VERSION       release tag to install, e.g. v0.1.0 (default: latest)
#   INSTALL_DIR   directory to install the binary into (default: /usr/local/bin,
#                 falling back to ~/.local/bin if that isn't writable)

set -eu

REPO="brpaz/sops-tui"
BIN_NAME="sops-tui"

os() {
  case "$(uname -s)" in
    Linux) echo "Linux" ;;
    Darwin) echo "Darwin" ;;
    *)
      echo "error: unsupported OS $(uname -s) — download a release manually from https://github.com/${REPO}/releases" >&2
      exit 1
      ;;
  esac
}

arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "x86_64" ;;
    aarch64 | arm64) echo "arm64" ;;
    i386 | i686) echo "i386" ;;
    *)
      echo "error: unsupported architecture $(uname -m) — download a release manually from https://github.com/${REPO}/releases" >&2
      exit 1
      ;;
  esac
}

latest_version() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name":' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

install_dir() {
  if [ -n "${INSTALL_DIR:-}" ]; then
    echo "$INSTALL_DIR"
  elif [ -w /usr/local/bin ]; then
    echo /usr/local/bin
  else
    echo "$HOME/.local/bin"
  fi
}

main() {
  os="$(os)"
  arch="$(arch)"
  version="${VERSION:-$(latest_version)}"
  if [ -z "$version" ]; then
    echo "error: could not resolve the latest release version" >&2
    exit 1
  fi
  version_number="${version#v}"

  dir="$(install_dir)"
  mkdir -p "$dir"

  asset="${BIN_NAME}_${os}_${arch}.tar.gz"
  checksums="${BIN_NAME}_${version_number}_checksums.txt"
  base_url="https://github.com/${REPO}/releases/download/${version}"

  workdir="$(mktemp -d)"
  trap 'rm -rf "$workdir"' EXIT

  echo "Downloading ${asset} (${version})..."
  curl -fsSL -o "${workdir}/${asset}" "${base_url}/${asset}"
  curl -fsSL -o "${workdir}/${checksums}" "${base_url}/${checksums}"

  echo "Verifying checksum..."
  (
    cd "$workdir"
    grep " ${asset}\$" "${checksums}" | sha256sum -c -
  )

  tar -xzf "${workdir}/${asset}" -C "$workdir"
  install -m 0755 "${workdir}/${BIN_NAME}" "${dir}/${BIN_NAME}"

  echo "Installed ${BIN_NAME} ${version} to ${dir}/${BIN_NAME}"
  case ":$PATH:" in
    *":${dir}:"*) ;;
    *) echo "note: ${dir} is not on your PATH — add it, e.g. export PATH=\"${dir}:\$PATH\"" ;;
  esac
}

main
