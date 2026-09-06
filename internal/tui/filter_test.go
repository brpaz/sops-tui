package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func typeText(t *testing.T, m Model, s string) Model {
	t.Helper()
	for _, r := range s {
		updated, _ := m.Update(tea.KeyPressMsg{Text: string(r)})
		m = updated.(Model)
	}
	return m
}

func TestFilter_LiveNarrowsRows(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "aaa.yaml"), []byte("x: 1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "bbb.yaml"), []byte("x: 1\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)
	require.Len(t, m.table.Rows(), 2)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "/"})
	m = updated.(Model)
	require.Equal(t, inputFilter, m.inputMode)

	m = typeText(t, m, "aaa")
	require.Len(t, m.table.Rows(), 1)
	require.Equal(t, "aaa.yaml", m.table.Rows()[0][1])
}

func TestFilter_EscCancelRestoresFullList(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "aaa.yaml"), []byte("x: 1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "bbb.yaml"), []byte("x: 1\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "/"})
	m = updated.(Model)
	m = typeText(t, m, "aaa")
	require.Len(t, m.table.Rows(), 1)

	updated, _ = m.Update(tea.KeyPressMsg{Text: "esc"})
	m = updated.(Model)

	require.Equal(t, inputNone, m.inputMode)
	require.Empty(t, m.filterQuery)
	require.Len(t, m.table.Rows(), 2, "cancelling restores the unfiltered list")
}

func TestFilter_EnterConfirmsAndSurvivesRefresh(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "aaa.yaml"), []byte("x: 1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "bbb.yaml"), []byte("x: 1\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "/"})
	m = updated.(Model)
	m = typeText(t, m, "bbb")
	updated, _ = m.Update(tea.KeyPressMsg{Text: "enter"})
	m = updated.(Model)

	require.Equal(t, inputNone, m.inputMode)
	require.Equal(t, "bbb", m.filterQuery)
	require.Len(t, m.table.Rows(), 1)

	// A new file appears on disk; a manual refresh must keep the
	// confirmed filter applied.
	require.NoError(t, os.WriteFile(filepath.Join(root, "ccc.yaml"), []byte("x: 1\n"), 0o644))
	updated, _ = m.Update(tea.KeyPressMsg{Text: "r"})
	m = updated.(Model)

	require.Len(t, m.table.Rows(), 1)
	require.Equal(t, "bbb.yaml", m.table.Rows()[0][1])
}

func TestCommandMode_ExecutesNamedActionOnSelectedRow(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)
	require.NoError(t, os.WriteFile(filepath.Join(root, "secret.yaml"), []byte("password: hunter2\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: ":"})
	m = updated.(Model)
	require.Equal(t, inputCommand, m.inputMode)

	m = typeText(t, m, "encrypt")
	updated, _ = m.Update(tea.KeyPressMsg{Text: "enter"})
	m = updated.(Model)

	require.Equal(t, inputNone, m.inputMode)
	require.Equal(t, "Encrypted", m.table.Rows()[0][0])
}

func TestCommandMode_UnknownCommandShowsErrorPane(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "secret.yaml"), []byte("password: hunter2\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: ":"})
	m = updated.(Model)
	m = typeText(t, m, "bogus")
	updated, _ = m.Update(tea.KeyPressMsg{Text: "enter"})
	m = updated.(Model)

	require.NotNil(t, m.pane)
	require.Contains(t, m.pane.content, "unknown command")
}

func TestCommandMode_EscCancelsWithoutSideEffects(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: ":"})
	m = updated.(Model)
	m = typeText(t, m, "encrypt")
	updated, _ = m.Update(tea.KeyPressMsg{Text: "esc"})
	m = updated.(Model)

	require.Equal(t, inputNone, m.inputMode)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "password: hunter2\n", string(raw), "cancelling a command must not execute it")
}

func TestHelpOverlay_TogglesAndDismisses(t *testing.T) {
	root := t.TempDir()
	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "?"})
	m = updated.(Model)
	require.True(t, m.showHelp)
	require.Contains(t, m.View().Content, "keybindings")

	updated, cmd := m.Update(tea.KeyPressMsg{Text: "q"})
	m = updated.(Model)
	require.False(t, m.showHelp)
	require.Nil(t, cmd, "q inside the help overlay closes it, it must not quit the app")
}
