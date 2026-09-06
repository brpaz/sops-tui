package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"
)

func TestEditAction_StartsExecAndTracksPath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, cmd := m.Update(tea.KeyPressMsg{Text: "E"})
	m = updated.(Model)

	require.NotNil(t, cmd, "starting edit must return a tea.Cmd (tea.ExecProcess)")
	require.Equal(t, "secret.yaml", m.editingPath)
}

func TestEditAction_NoOpWithNoRowsSelected(t *testing.T) {
	root := t.TempDir()

	m, err := New(root)
	require.NoError(t, err)
	require.Empty(t, m.table.Rows())

	updated, cmd := m.Update(tea.KeyPressMsg{Text: "E"})
	m = updated.(Model)

	require.Nil(t, cmd)
	require.Empty(t, m.editingPath)
}

func TestEditAction_SuccessfulEditRefreshesList(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "E"})
	m = updated.(Model)
	require.Equal(t, "secret.yaml", m.editingPath)

	// Simulate the subprocess having encrypted the file while suspended,
	// then reporting a clean exit.
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\nsops:\n    version: 3\n"), 0o644))

	updated, _ = m.Update(editDoneMsg{err: nil})
	m = updated.(Model)

	require.Empty(t, m.editingPath)
	require.Nil(t, m.pane)
	require.Equal(t, "Encrypted", m.table.Rows()[0][0], "list reflects disk state after resume")
}

func TestEditAction_FailedEditShowsErrorPaneAndStillRefreshes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	m, err := New(root)
	require.NoError(t, err)

	updated, _ := m.Update(tea.KeyPressMsg{Text: "E"})
	m = updated.(Model)

	updated, _ = m.Update(editDoneMsg{err: fmt.Errorf("exit status 1")})
	m = updated.(Model)

	require.Empty(t, m.editingPath)
	require.NotNil(t, m.pane)
	require.Contains(t, m.pane.content, "edit did not complete")
	require.Len(t, m.table.Rows(), 1, "list is still refreshed after a failed edit")
}
