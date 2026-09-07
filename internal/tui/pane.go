package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/brpaz/sops-tui/internal/secrets"
)

// pane holds a full-screen overlay shown on top of the list, such as a
// file's decrypted content. A nil pane means the list itself is shown.
type pane struct {
	path    string
	content string
}

// viewPane returns a pane showing relPath's content, or an error message
// in place of the content if that failed. The file on disk is never
// modified. Only encrypted rows go through `sops -d`; sops itself refuses
// to decrypt a file with no sops metadata, so plaintext rows are read
// directly instead.
func (a *App) viewPane(relPath string) *pane {
	full := a.fullPath(relPath)

	if a.selectedStatus() != secrets.StatusEncrypted {
		data, err := os.ReadFile(full)
		if err != nil {
			return errorPane(relPath, err)
		}
		return &pane{path: relPath, content: string(data)}
	}

	content, err := secrets.View(full)
	if err != nil {
		return errorPane(relPath, err)
	}
	return &pane{path: relPath, content: content}
}

// errorPane builds a pane naming relPath and showing err's message,
// unmodified, for actions that fail.
func errorPane(relPath string, err error) *pane {
	return &pane{path: relPath, content: fmt.Sprintf("Error: %v", err)}
}

// fullPath resolves a row's relative path against the scan root.
func (a *App) fullPath(relPath string) string {
	return filepath.Join(a.root, relPath)
}

// selectedPath returns the currently selected row's relative path, or ""
// if no row is selected.
func (a *App) selectedPath() string {
	row, _ := a.table.GetSelection()
	if row < 1 || row >= a.table.GetRowCount() {
		return ""
	}
	return a.table.GetCell(row, 0).Text
}

// selectedStatus returns the currently selected row's encryption status,
// or secrets.StatusUnknown if no row is selected. There is no visible
// status column (each view mode already implies it), so this reads back
// the reference rebuildRows attached to the row's cell.
func (a *App) selectedStatus() secrets.Status {
	row, _ := a.table.GetSelection()
	if row < 1 || row >= a.table.GetRowCount() {
		return secrets.StatusUnknown
	}
	status, _ := a.table.GetCell(row, 0).GetReference().(secrets.Status)
	return status
}

// moveSelection shifts the selected row by delta, clamped to the range of
// data rows (row 0 is the header and is never selected).
func (a *App) moveSelection(delta int) {
	n := a.table.GetRowCount() - 1
	if n <= 0 {
		return
	}
	row, col := a.table.GetSelection()
	row += delta
	if row < 1 {
		row = 1
	}
	if row > n {
		row = n
	}
	a.table.Select(row, col)
}

// selectTop selects the first data row, if any.
func (a *App) selectTop() {
	if a.table.GetRowCount() > 1 {
		a.table.Select(1, 0)
	}
}

// selectBottom selects the last data row, if any.
func (a *App) selectBottom() {
	if n := a.table.GetRowCount() - 1; n > 0 {
		a.table.Select(n, 0)
	}
}
