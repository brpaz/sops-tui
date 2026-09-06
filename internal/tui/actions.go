package tui

import "github.com/brpaz/sops-tui/internal/secrets"

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
