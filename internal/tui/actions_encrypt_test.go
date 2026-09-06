package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/sops-tui/internal/secrets"
)

func TestEncryptAction_EncryptsAndRefreshesList(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)
	require.Equal(t, "Plaintext", m.table.Rows()[0][0])

	updated, _ := m.Update(tea.KeyPressMsg{Text: "e"})
	m = updated.(Model)

	require.Nil(t, m.pane, "successful encrypt does not open an error pane")
	require.Equal(t, "Encrypted", m.table.Rows()[0][0], "list refreshed without a manual r")

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

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "e"})
	m = updated.(Model)

	require.NotNil(t, m.pane)
	require.Contains(t, m.pane.content, "nomatch.yaml")
	require.Equal(t, "Plaintext", m.table.Rows()[0][0], "file is left untouched")
}

func TestEncryptAction_NoOpOnAlreadyEncryptedRow(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))
	require.NoError(t, secrets.Encrypt(path))

	m, err := New(root)
	require.NoError(t, err)
	require.Equal(t, "Encrypted", m.table.Rows()[0][0])

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "e"})
	m = updated.(Model)

	require.Nil(t, m.pane)
	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after)
}
