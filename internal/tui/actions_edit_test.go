package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/sops-tui/internal/secrets"
)

// fakeSuspend replaces App.suspend in tests: it calls f synchronously,
// standing in for app.Suspend (which is a no-op before Run has
// initialized the screen).
func fakeSuspend(f func()) { f() }

func TestEditAction_RunsEditOnSelectedRow(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	var editedPath string
	a.suspend = fakeSuspend
	a.edit = func(p string) error {
		editedPath = p
		return nil
	}

	a.handleKey(key("E"))

	require.Equal(t, path, editedPath, "edit must run against the selected row's full path")
}

func TestEditAction_NoOpWithNoRowsSelected(t *testing.T) {
	root := t.TempDir()

	a, err := New(root)
	require.NoError(t, err)
	require.Empty(t, tableRows(a))

	called := false
	a.suspend = fakeSuspend
	a.edit = func(string) error {
		called = true
		return nil
	}

	a.handleKey(key("E"))

	require.False(t, called, "edit must not run when nothing is selected")
}

func TestEditAction_SuccessfulEditRefreshesList(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.suspend = fakeSuspend
	// Simulate the subprocess having encrypted the file while suspended.
	a.edit = func(p string) error {
		return os.WriteFile(p, []byte("password: hunter2\nsops:\n    version: 3\n"), 0o644)
	}

	a.handleKey(key("E"))

	require.Nil(t, a.pane)
	require.Equal(t, secrets.StatusEncrypted, entryStatus(t, a, "secret.yaml"), "list reflects disk state after resume")
}

func TestEditAction_UnchangedFileExitCodeIsNotAnError(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.suspend = fakeSuspend
	// sops itself exits 200 ("the file has not been modified, exiting.")
	// when the editor closed without changes; that's not a failure.
	a.edit = func(string) error {
		return exec.Command("sh", "-c", "exit 200").Run()
	}

	a.handleKey(key("E"))

	require.Nil(t, a.pane, "an unmodified-file exit must not open an error pane")
	require.Len(t, tableRows(a), 1, "list is still refreshed")
}

func TestEditAction_FailedEditShowsErrorPaneAndStillRefreshes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()

	a.suspend = fakeSuspend
	a.edit = func(string) error {
		return fmt.Errorf("exit status 1")
	}

	a.handleKey(key("E"))

	require.NotNil(t, a.pane)
	require.Contains(t, a.pane.content, "edit did not complete")
	require.Len(t, tableRows(a), 1, "list is still refreshed after a failed edit")
}
