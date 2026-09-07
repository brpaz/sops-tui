package tui

import (
	"os"
	"path/filepath"
	"testing"

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

	a, err := New(root)
	require.NoError(t, err)

	a.handleKey(key("d"))

	require.Equal(t, &confirmPane{kind: "decrypt", path: "secret.yaml"}, a.confirm)

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after, "confirmation prompt alone must not touch disk")
}

func TestDecryptAction_ConfirmDecryptsAndRefreshes(t *testing.T) {
	root := t.TempDir()
	path := encryptedFixture(t, root)

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.handleKey(key("d"))
	a.handleKey(key("y"))

	require.Nil(t, a.confirm)
	require.Equal(t, secrets.StatusPlaintext, entryStatus(t, a, "secret.yaml"), "list refreshed without a manual r")

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(raw), "password: hunter2")
}

func TestDecryptAction_CancelLeavesFileUnchanged(t *testing.T) {
	root := t.TempDir()
	path := encryptedFixture(t, root)
	before, err := os.ReadFile(path)
	require.NoError(t, err)

	a, err := New(root)
	require.NoError(t, err)

	a.handleKey(key("d"))
	a.handleKey(key("n"))

	require.Nil(t, a.confirm)
	require.Equal(t, secrets.StatusEncrypted, entryStatus(t, a, "secret.yaml"))

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestDecryptAction_NoOpOnPlaintextRow(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "plain.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.handleKey(key("d"))

	require.Nil(t, a.confirm, "no confirmation should be raised for a plaintext row")
}
