// Package tui implements the sops-tui BubbleTea application: a flat,
// k9s-style table of SOPS-managed files with vim-style navigation.
package tui

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"

	"github.com/brpaz/sops-tui/internal/secrets"
)

const (
	statusColumnWidth = 12
	pathColumnWidth   = 60
	helpLine          = "j/k: move  gg/G: top/bottom  r: refresh  q: quit"
	reservedRows      = 3 // help line + table header + margin
)

// Model is the root BubbleTea model for sops-tui.
type Model struct {
	root  string
	table table.Model
	// err is a fatal error that replaces the table view entirely, e.g.
	// the scan root could not be read.
	err error
}

// New builds the TUI model for the given scan root. It requires the sops
// binary to already be available on PATH; callers should check that via
// secrets.CheckAvailable before calling New.
func New(root string) (Model, error) {
	m := Model{
		root: root,
		table: table.New(
			table.WithColumns([]table.Column{
				{Title: "Status", Width: statusColumnWidth},
				{Title: "Path", Width: pathColumnWidth},
			}),
			table.WithFocused(true),
		),
	}

	if err := m.refresh(); err != nil {
		return Model{}, err
	}

	return m, nil
}

// refresh re-scans the root directory and repopulates the table.
func (m *Model) refresh() error {
	entries, err := secrets.List(m.root)
	if err != nil {
		return fmt.Errorf("scanning %s: %w", m.root, err)
	}

	rows := make([]table.Row, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, table.Row{e.Status.String(), e.Path})
	}
	m.table.SetRows(rows)

	return nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(max(msg.Height-reservedRows, 1))
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if err := m.refresh(); err != nil {
				m.err = err
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	if m.err != nil {
		v := tea.NewView(fmt.Sprintf("Error: %v\n\npress q to quit\n", m.err))
		v.AltScreen = true
		return v
	}

	v := tea.NewView(m.table.View() + "\n" + helpLine)
	v.AltScreen = true
	return v
}
