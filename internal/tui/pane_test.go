package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/sops-tui/internal/secrets"
)

// ageFixture generates a throwaway age keypair, points sops at its
// private key, and returns the public key for a test-local .sops.yaml.
func ageFixture(t *testing.T) (publicKey string) {
	t.Helper()

	out, err := exec.Command("age-keygen").Output()
	require.NoError(t, err, "age-keygen must be installed to run these tests")

	keyFile := filepath.Join(t.TempDir(), "keys.txt")
	require.NoError(t, os.WriteFile(keyFile, out, 0o600))
	t.Setenv("SOPS_AGE_KEY_FILE", keyFile)

	for _, line := range strings.Split(string(out), "\n") {
		if pk, ok := strings.CutPrefix(line, "# public key: "); ok {
			return pk
		}
	}
	t.Fatal("could not find public key in age-keygen output")
	return ""
}

func writeSopsConfig(t *testing.T, dir, publicKey string) {
	t.Helper()
	content := "creation_rules:\n  - path_regex: '.*\\.yaml$'\n    age: '" + publicKey + "'\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".sops.yaml"), []byte(content), 0o644))
}

func TestViewAction_ShowsDecryptedContentWithoutModifyingDisk(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))
	require.NoError(t, secrets.Encrypt(path))

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "v"})
	m = updated.(Model)

	require.NotNil(t, m.pane)
	require.Contains(t, m.pane.content, "password: hunter2")

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after, "viewing must not modify the file on disk")

	updated, _ = m.Update(tea.KeyPressMsg{Text: "esc"})
	m = updated.(Model)
	require.Nil(t, m.pane, "esc returns to the list")
}

func TestViewAction_QReturnsToPane(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))
	require.NoError(t, secrets.Encrypt(path))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "enter"})
	m = updated.(Model)
	require.NotNil(t, m.pane)

	updated, _ = m.Update(tea.KeyPressMsg{Text: "q"})
	m = updated.(Model)
	require.Nil(t, m.pane, "q closes the pane rather than quitting the app")
}

func TestViewAction_DecryptFailureShowsErrorInPane(t *testing.T) {
	root := t.TempDir()
	// A file that looks encrypted (has a top-level "sops" key so our
	// status detection reports it as Encrypted) but isn't a valid sops
	// document, so `sops -d` itself fails.
	path := filepath.Join(root, "broken.yaml")
	require.NoError(t, os.WriteFile(path, []byte("data: hello\nsops:\n    version: 3\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "v"})
	m = updated.(Model)

	require.NotNil(t, m.pane)
	require.Contains(t, m.pane.content, "Error:")
}
