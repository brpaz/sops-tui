package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestNew_PopulatesTableFromScan(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plain.yaml"), "data: hello\n")
	writeFile(t, filepath.Join(root, "enc.yaml"), "data: ENC[...]\nsops:\n    version: 3\n")

	a, err := New(root)
	require.NoError(t, err)

	rows := tableRows(a)
	require.Len(t, rows, 1, "default view shows only encrypted files")
	require.Equal(t, "enc.yaml", rows[0])
}

func TestUpdate_TabCyclesViewModes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plain.yaml"), "data: hello\n")
	writeFile(t, filepath.Join(root, "enc.yaml"), "data: ENC[...]\nsops:\n    version: 3\n")

	a, err := New(root)
	require.NoError(t, err)
	require.Len(t, tableRows(a), 1, "default view shows only encrypted files")
	require.Equal(t, "enc.yaml", tableRows(a)[0])

	a.handleKey(key("tab"))
	rows := tableRows(a)
	require.Len(t, rows, 1, "plaintext view shows only plaintext files")
	require.Equal(t, "plain.yaml", rows[0])

	a.handleKey(key("tab"))
	require.Len(t, tableRows(a), 2, "all view shows every file")

	a.handleKey(key("tab"))
	require.Len(t, tableRows(a), 1, "cycling wraps back to the encrypted view")
	require.Equal(t, "enc.yaml", tableRows(a)[0])
}

func TestUpdate_RefreshKeyRescans(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plain.yaml"), "data: hello\n")

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewAll
	a.rebuildRows()
	require.Len(t, tableRows(a), 1)

	writeFile(t, filepath.Join(root, "second.yaml"), "data: world\n")

	a.handleKey(key("r"))
	require.Len(t, tableRows(a), 2)
}

func TestUpdate_QuitKeyStopsApp(t *testing.T) {
	root := t.TempDir()
	a, err := New(root)
	require.NoError(t, err)

	// app.Stop() is a no-op before the screen is initialized; this just
	// confirms the key is wired to it without a real terminal.
	require.NotPanics(t, func() { a.handleKey(key("q")) })
}

func TestRun_ErrorsBeforeRenderingWhenSopsUnavailable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := Run(t.Context(), t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "sops binary not found")
}
