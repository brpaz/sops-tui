package tui

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/brpaz/sops-tui/internal/secrets"
)

// sopsExitFileUnchanged is the exit code sops itself uses for "the file
// has not been modified, exiting.": the user opened the editor and saved
// without changing anything. It's not a failure, so startEdit treats it
// like a clean exit rather than surfacing an error pane.
const sopsExitFileUnchanged = 200

func isSopsFileUnchanged(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == sopsExitFileUnchanged
}

// startEncrypt raises a confirmation for encrypting the selected plaintext
// row via .sops.yaml creation rules. A no-op on anything but a plaintext
// row; nothing touches disk until the confirmation is accepted.
func (a *App) startEncrypt() {
	if a.selectedStatus() != secrets.StatusPlaintext {
		return
	}
	if path := a.selectedPath(); path != "" {
		a.confirm = &confirmPane{kind: "encrypt", path: path}
	}
}

// startDecrypt raises a confirmation for decrypting the selected encrypted
// row in place. A no-op on anything but an encrypted row; nothing touches
// disk until the confirmation is accepted.
func (a *App) startDecrypt() {
	if a.selectedStatus() != secrets.StatusEncrypted {
		return
	}
	if path := a.selectedPath(); path != "" {
		a.confirm = &confirmPane{kind: "decrypt", path: path}
	}
}

// confirmText renders c's full-screen confirmation prompt, worded for its
// specific action so encrypt and decrypt read as distinct screens rather
// than a generic yes/no dialog.
func confirmText(c *confirmPane) string {
	switch c.kind {
	case "encrypt":
		return fmt.Sprintf(
			"Encrypt %s?\nThis writes ciphertext to disk using .sops.yaml creation rules.\n\n[y] yes   [n/esc] cancel",
			c.path,
		)
	default:
		return fmt.Sprintf(
			"Decrypt %s in place?\nThis writes plaintext to disk.\n\n[y] yes   [n/esc] cancel",
			c.path,
		)
	}
}

// confirmYes performs the pending encrypt/decrypt action and refreshes the
// list. It clears the pending confirmation either way. On failure an error
// pane names the file and shows the underlying reason, e.g. sops finding
// no matching creation rule.
func (a *App) confirmYes() {
	if a.confirm == nil {
		return
	}
	kind, path := a.confirm.kind, a.confirm.path
	a.confirm = nil

	var err error
	switch kind {
	case "encrypt":
		err = secrets.Encrypt(a.fullPath(path))
	case "decrypt":
		err = secrets.Decrypt(a.fullPath(path))
	}
	if err != nil {
		a.pane = errorPane(path, err)
		return
	}

	if err := a.refresh(); err != nil {
		a.fatalErr = err
	}
}

// executeCommand runs a ":" command by name against the currently
// selected row: view, encrypt, decrypt, edit, or refresh. An unrecognized
// command name shows an error pane naming it.
func (a *App) executeCommand(name string) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "view":
		if path := a.selectedPath(); path != "" {
			a.pane = a.viewPane(path)
		}
	case "encrypt":
		a.startEncrypt()
	case "decrypt":
		a.startDecrypt()
	case "edit":
		a.startEdit()
	case "refresh":
		if err := a.refresh(); err != nil {
			a.fatalErr = err
		}
	default:
		a.pane = &pane{path: ":" + name, content: fmt.Sprintf("Error: unknown command %q", name)}
	}
}

// startEdit suspends the TUI and runs `sops <file>` in the foreground on
// the selected row, refreshing the list once it exits. A no-op if no row
// is selected.
func (a *App) startEdit() {
	path := a.selectedPath()
	if path == "" {
		return
	}

	var editErr error
	a.suspend(func() {
		editErr = a.edit(a.fullPath(path))
	})

	if err := a.refresh(); err != nil {
		a.fatalErr = err
		return
	}
	if editErr != nil && !isSopsFileUnchanged(editErr) {
		a.pane = errorPane(path, fmt.Errorf("edit did not complete: %w", editErr))
	}
}
