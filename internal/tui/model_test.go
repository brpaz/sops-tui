package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
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

	m, err := New(root)
	require.NoError(t, err)

	rows := m.table.Rows()
	require.Len(t, rows, 2)

	byPath := map[string]string{}
	for _, r := range rows {
		byPath[r[1]] = r[0]
	}
	require.Equal(t, "Plaintext", byPath["plain.yaml"])
	require.Equal(t, "Encrypted", byPath["enc.yaml"])
}

func TestUpdate_RefreshKeyRescans(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "plain.yaml"), "data: hello\n")

	m, err := New(root)
	require.NoError(t, err)
	require.Len(t, m.table.Rows(), 1)

	writeFile(t, filepath.Join(root, "second.yaml"), "data: world\n")

	updated, _ := m.Update(tea.KeyPressMsg{Text: "r"})
	m = updated.(Model)
	require.Len(t, m.table.Rows(), 2)
}

func TestUpdate_QuitKeyReturnsQuitCmd(t *testing.T) {
	root := t.TempDir()
	m, err := New(root)
	require.NoError(t, err)

	_, cmd := m.Update(tea.KeyPressMsg{Text: "q"})
	require.NotNil(t, cmd)
	require.IsType(t, tea.QuitMsg{}, cmd())
}

func TestRun_ErrorsBeforeRenderingWhenSopsUnavailable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := Run(t.Context(), t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "sops binary not found")
}
