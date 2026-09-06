package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/brpaz/sops-tui/internal/secrets"
)

const (
	statusEncrypted = "Encrypted"
	statusPlaintext = "Plaintext"
)

// encryptSelected encrypts the selected plaintext row via .sops.yaml
// creation rules. On success the list is refreshed in place; on failure
// (e.g. no matching creation rule) an error pane names the file and shows
// sops's own reported reason. A no-op on anything but a plaintext row.
func (m *Model) encryptSelected() {
	if m.selectedStatus() != statusPlaintext {
		return
	}

	path := m.selectedPath()
	if path == "" {
		return
	}

	if err := secrets.Encrypt(m.fullPath(path)); err != nil {
		m.pane = errorPane(path, err)
		return
	}

	if err := m.refresh(); err != nil {
		m.err = err
	}
}

// startEdit suspends the TUI and runs `sops <file>` in the foreground on
// the selected row, resuming when the subprocess exits. A no-op if no row
// is selected.
func (m Model) startEdit() (Model, tea.Cmd) {
	path := m.selectedPath()
	if path == "" {
		return m, nil
	}

	m.editingPath = path
	cmd := secrets.EditCommand(m.fullPath(path))

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editDoneMsg{err: err}
	})
}

// confirmDecryptYes decrypts the file pending confirmation in place and
// refreshes the list. It clears the pending confirmation either way.
func (m *Model) confirmDecryptYes() {
	path := m.confirmDecrypt
	m.confirmDecrypt = ""

	if err := secrets.Decrypt(m.fullPath(path)); err != nil {
		m.pane = errorPane(path, err)
		return
	}

	if err := m.refresh(); err != nil {
		m.err = err
	}
}
