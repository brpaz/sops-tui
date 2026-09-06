package tui

import (
	"fmt"
	"path/filepath"

	"github.com/brpaz/sops-tui/internal/secrets"
)

// pane holds a full-screen overlay shown on top of the list, such as a
// file's decrypted content. A nil pane means the list itself is shown.
type pane struct {
	path    string
	content string
}

// viewPane decrypts relPath and returns a pane showing its content, or an
// error message in place of the content if decryption failed. The file on
// disk is never modified.
func (m Model) viewPane(relPath string) *pane {
	content, err := secrets.View(m.fullPath(relPath))
	if err != nil {
		content = fmt.Sprintf("Error: %v", err)
	}
	return &pane{path: relPath, content: content}
}

// fullPath resolves a row's relative path against the scan root.
func (m Model) fullPath(relPath string) string {
	return filepath.Join(m.root, relPath)
}

// selectedPath returns the currently selected row's relative path, or ""
// if no row is selected.
func (m Model) selectedPath() string {
	row := m.table.SelectedRow()
	if len(row) < 2 {
		return ""
	}
	return row[1]
}
