package tui

import (
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

// applyFilter sets the table's rows to the subset of allRows whose path
// column contains filterQuery, case-insensitively. An empty filterQuery
// shows every row.
func (m *Model) applyFilter() {
	if m.filterQuery == "" {
		m.table.SetRows(m.allRows)
		return
	}

	needle := strings.ToLower(m.filterQuery)
	rows := make([]table.Row, 0, len(m.allRows))
	for _, row := range m.allRows {
		if len(row) > 1 && strings.Contains(strings.ToLower(row[1]), needle) {
			rows = append(rows, row)
		}
	}
	m.table.SetRows(rows)
}

// startInput switches to the given input mode, seeding the text input
// with an optional prompt and initial value.
func (m *Model) startInput(mode inputMode, prompt, value string) {
	m.inputMode = mode
	m.input.Prompt = prompt
	m.input.SetValue(value)
	m.input.CursorEnd()
	m.input.Focus()
}

// cancelInput exits the active input without applying it. A cancelled
// filter reverts to whatever was active before the input was opened.
func (m *Model) cancelInput() {
	if m.inputMode == inputFilter {
		m.filterQuery = m.filterPrevQuery
		m.applyFilter()
	}
	m.inputMode = inputNone
	m.input.Blur()
}

// confirmInput applies the active input's value and closes it: a filter
// becomes the new confirmed filterQuery, a command is executed against
// the selected row.
func (m *Model) confirmInput() (Model, tea.Cmd) {
	mode := m.inputMode
	value := m.input.Value()
	m.inputMode = inputNone
	m.input.Blur()

	if mode == inputFilter {
		m.filterQuery = value
		m.applyFilter()
		return *m, nil
	}

	return m.executeCommand(value)
}
