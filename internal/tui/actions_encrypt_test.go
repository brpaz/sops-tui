package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/sops-tui/internal/secrets"
)

func TestEncryptAction_EncryptsAndRefreshesList(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()
	require.Equal(t, secrets.StatusPlaintext, entryStatus(t, a, "secret.yaml"))

	a.handleKey(key("e"))
	require.Equal(t, &confirmPane{kind: "encrypt", path: "secret.yaml"}, a.confirm, "encrypt asks for confirmation before touching disk")
	a.handleKey(key("y"))

	require.Nil(t, a.confirm)
	require.Nil(t, a.pane, "successful encrypt does not open an error pane")
	require.Equal(t, secrets.StatusEncrypted, entryStatus(t, a, "secret.yaml"), "list refreshed without a manual r")

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(raw), "sops")
}

func TestEncryptAction_NoMatchingRuleShowsErrorPane(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	// Config only matches "matched.yaml", not the file we'll try to encrypt.
	content := "creation_rules:\n  - path_regex: '^matched\\.yaml$'\n    age: '" + publicKey + "'\n"
	require.NoError(t, os.WriteFile(filepath.Join(root, ".sops.yaml"), []byte(content), 0o644))

	path := filepath.Join(root, "nomatch.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.handleKey(key("e"))
	a.handleKey(key("y"))

	require.Nil(t, a.confirm)
	require.NotNil(t, a.pane)
	require.Contains(t, a.pane.content, "nomatch.yaml")
	require.Equal(t, secrets.StatusPlaintext, entryStatus(t, a, "nomatch.yaml"), "file is left untouched")
}

func TestEncryptAction_NoOpOnAlreadyEncryptedRow(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))
	require.NoError(t, secrets.Encrypt(path))

	a, err := New(root)
	require.NoError(t, err)
	require.Equal(t, secrets.StatusEncrypted, entryStatus(t, a, "secret.yaml"))

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	a.handleKey(key("e"))

	require.Nil(t, a.confirm, "already-encrypted row raises no confirmation")
	require.Nil(t, a.pane)
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
}
