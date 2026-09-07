package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/sops-tui/internal/secrets"
)

func TestFilter_LiveNarrowsRows(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "aaa.yaml"), []byte("x: 1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "bbb.yaml"), []byte("x: 1\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()
	require.Len(t, tableRows(a), 2)

	a.handleKey(key("/"))
	require.Equal(t, inputFilter, a.inputMode)

	typeText(t, a, "aaa")
	rows := tableRows(a)
	require.Len(t, rows, 1)
	require.Equal(t, "aaa.yaml", rows[0])
}

func TestFilter_EscCancelRestoresFullList(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "aaa.yaml"), []byte("x: 1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "bbb.yaml"), []byte("x: 1\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.handleKey(key("/"))
	typeText(t, a, "aaa")
	require.Len(t, tableRows(a), 1)

	a.handleKey(key("esc"))

	require.Equal(t, inputNone, a.inputMode)
	require.Empty(t, a.filterQuery)
	require.Len(t, tableRows(a), 2, "cancelling restores the unfiltered list")
}

func TestFilter_EnterConfirmsAndSurvivesRefresh(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "aaa.yaml"), []byte("x: 1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "bbb.yaml"), []byte("x: 1\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.handleKey(key("/"))
	typeText(t, a, "bbb")
	a.handleKey(key("enter"))

	require.Equal(t, inputNone, a.inputMode)
	require.Equal(t, "bbb", a.filterQuery)
	require.Len(t, tableRows(a), 1)

	// A new file appears on disk; a manual refresh must keep the
	// confirmed filter applied.
	require.NoError(t, os.WriteFile(filepath.Join(root, "ccc.yaml"), []byte("x: 1\n"), 0o644))
	a.handleKey(key("r"))

	rows := tableRows(a)
	require.Len(t, rows, 1)
	require.Equal(t, "bbb.yaml", rows[0])
}

func TestCommandMode_ExecutesNamedActionOnSelectedRow(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)
	require.NoError(t, os.WriteFile(filepath.Join(root, "secret.yaml"), []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.handleKey(key(":"))
	require.Equal(t, inputCommand, a.inputMode)

	typeText(t, a, "encrypt")
	a.handleKey(key("enter"))
	require.NotNil(t, a.confirm, "the encrypt command raises the same confirmation as the e key")
	a.handleKey(key("y"))

	require.Equal(t, inputNone, a.inputMode)
	require.Equal(t, secrets.StatusEncrypted, entryStatus(t, a, "secret.yaml"))
}

func TestCommandMode_UnknownCommandShowsErrorPane(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "secret.yaml"), []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)

	a.handleKey(key(":"))
	typeText(t, a, "bogus")
	a.handleKey(key("enter"))

	require.NotNil(t, a.pane)
	require.Contains(t, a.pane.content, "unknown command")
}

func TestCommandMode_EscCancelsWithoutSideEffects(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)

	a.handleKey(key(":"))
	typeText(t, a, "encrypt")
	a.handleKey(key("esc"))

	require.Equal(t, inputNone, a.inputMode)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "password: hunter2\n", string(raw), "cancelling a command must not execute it")
}

func TestHelpOverlay_TogglesAndDismisses(t *testing.T) {
	root := t.TempDir()
	a, err := New(root)
	require.NoError(t, err)

	a.handleKey(key("?"))
	require.True(t, a.showHelp)
	require.Contains(t, a.helpView.GetText(true), "keybindings")

	a.handleKey(key("q"))
	require.False(t, a.showHelp, "q inside the help overlay closes it, it must not quit the app")
}
