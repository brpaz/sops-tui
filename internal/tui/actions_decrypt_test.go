package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/sops-tui/internal/secrets"
)

func encryptedFixture(t *testing.T, root string) (path string) {
	t.Helper()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path = filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))
	require.NoError(t, secrets.Encrypt(path))
	return path
}

func TestDecryptAction_ShowsConfirmationBeforeTouchingDisk(t *testing.T) {
	root := t.TempDir()
	path := encryptedFixture(t, root)
	before, err := os.ReadFile(path)
	require.NoError(t, err)

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "d"})
	m = updated.(Model)

	require.Equal(t, "secret.yaml", m.confirmDecrypt)

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after, "confirmation prompt alone must not touch disk")
}

func TestDecryptAction_ConfirmDecryptsAndRefreshes(t *testing.T) {
	root := t.TempDir()
	path := encryptedFixture(t, root)

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "d"})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg{Text: "y"})
	m = updated.(Model)

	require.Empty(t, m.confirmDecrypt)
	require.Equal(t, "Plaintext", m.table.Rows()[0][0], "list refreshed without a manual r")

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(raw), "password: hunter2")
}

func TestDecryptAction_CancelLeavesFileUnchanged(t *testing.T) {
	root := t.TempDir()
	path := encryptedFixture(t, root)
	before, err := os.ReadFile(path)
	require.NoError(t, err)

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "d"})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyPressMsg{Text: "n"})
	m = updated.(Model)

	require.Empty(t, m.confirmDecrypt)
	require.Equal(t, "Encrypted", m.table.Rows()[0][0])

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestDecryptAction_NoOpOnPlaintextRow(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "plain.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "d"})
	m = updated.(Model)

	require.Empty(t, m.confirmDecrypt, "no confirmation should be raised for a plaintext row")
}
