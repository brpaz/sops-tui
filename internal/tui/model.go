// Package tui implements the sops-tui BubbleTea application: a flat,
// k9s-style table of SOPS-managed files with vim-style navigation.
package tui

import (
	"fmt"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/brpaz/sops-tui/internal/secrets"
)

const (
	statusColumnWidth = 12
	pathColumnWidth   = 60
	helpLine          = "j/k: move  gg/G: top/bottom  v/enter: view  E: edit  e: encrypt  d: decrypt  /: filter  :: command  r: refresh  ?: help  q: quit"
	paneHelpLine      = "esc/q: back to list"
	reservedRows      = 3 // help line + table header + margin
)

// inputMode identifies which single-line input, if any, is currently
// capturing keystrokes instead of the table.
type inputMode int

const (
	inputNone inputMode = iota
	inputFilter
	inputCommand
)

// Model is the root BubbleTea model for sops-tui.
type Model struct {
	root  string
	table table.Model
	// err is a fatal error that replaces the table view entirely, e.g.
	// the scan root could not be read.
	err error
	// pane, when non-nil, is a full-screen overlay (e.g. decrypted file
	// content) shown instead of the list.
	pane *pane
	// confirmDecrypt, when non-empty, is the relative path awaiting a
	// yes/no confirmation before decrypting in place.
	confirmDecrypt string
	// editingPath is the relative path of the file being edited via a
	// suspended sops <file> subprocess, or "" when not editing.
	editingPath string
	// showHelp, when true, replaces the list with the keybinding overlay.
	showHelp bool

	// allRows is the unfiltered result of the last scan; the table shows
	// a filtered view of it whenever filterQuery is non-empty.
	allRows []table.Row
	// filterQuery is the confirmed filter substring (case-insensitive,
	// matched against the path column), or "" for no filter.
	filterQuery string
	// filterPrevQuery snapshots filterQuery when entering filter input,
	// so esc can restore it.
	filterPrevQuery string

	// inputMode is which single-line input is active, if any.
	inputMode inputMode
	input     textinput.Model
}

// editDoneMsg reports that the suspended `sops <file>` edit subprocess
// has exited.
type editDoneMsg struct{ err error }

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
		input: textinput.New(),
	}

	if err := m.refresh(); err != nil {
		return Model{}, err
	}

	return m, nil
}

// refresh re-scans the root directory and repopulates the table, applying
// any active filter.
func (m *Model) refresh() error {
	entries, err := secrets.List(m.root)
	if err != nil {
		return fmt.Errorf("scanning %s: %w", m.root, err)
	}

	rows := make([]table.Row, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, table.Row{e.Status.String(), e.Path})
	}
	m.allRows = rows
	m.applyFilter()

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

	case editDoneMsg:
		path := m.editingPath
		m.editingPath = ""
		if err := m.refresh(); err != nil {
			m.err = err
			return m, nil
		}
		if msg.err != nil {
			m.pane = errorPane(path, fmt.Errorf("edit did not complete: %w", msg.err))
		}
		return m, nil

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.showHelp {
			switch msg.String() {
			case "esc", "q", "?":
				m.showHelp = false
			}
			return m, nil
		}

		if m.inputMode != inputNone {
			switch msg.String() {
			case "esc":
				m.cancelInput()
				return m, nil
			case "enter":
				return m.confirmInput()
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			if m.inputMode == inputFilter {
				m.filterQuery = m.input.Value()
				m.applyFilter()
			}
			return m, cmd
		}

		if m.confirmDecrypt != "" {
			switch msg.String() {
			case "y":
				m.confirmDecryptYes()
			case "n", "esc":
				m.confirmDecrypt = ""
			}
			return m, nil
		}

		if m.pane != nil {
			switch msg.String() {
			case "esc", "q", "enter":
				m.pane = nil
			}
			return m, nil
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "?":
			m.showHelp = true
			return m, nil
		case "/":
			m.startInput(inputFilter, "/", m.filterQuery)
			return m, nil
		case ":":
			m.startInput(inputCommand, ":", "")
			return m, nil
		case "r":
			if err := m.refresh(); err != nil {
				m.err = err
			}
			return m, nil
		case "v", "enter":
			if path := m.selectedPath(); path != "" {
				m.pane = m.viewPane(path)
			}
			return m, nil
		case "e":
			m.encryptSelected()
			return m, nil
		case "E":
			return m.startEdit()
		case "d":
			if m.selectedStatus() == statusEncrypted {
				m.confirmDecrypt = m.selectedPath()
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

	if m.showHelp {
		v := tea.NewView(helpOverlay)
		v.AltScreen = true
		return v
	}

	if m.confirmDecrypt != "" {
		v := tea.NewView(fmt.Sprintf(
			"Decrypt %s in place?\nThis writes plaintext to disk.\n\n[y] yes   [n/esc] cancel",
			m.confirmDecrypt,
		))
		v.AltScreen = true
		return v
	}

	if m.pane != nil {
		v := tea.NewView(m.pane.path + "\n\n" + m.pane.content + "\n" + paneHelpLine)
		v.AltScreen = true
		return v
	}

	if m.inputMode != inputNone {
		v := tea.NewView(m.table.View() + "\n" + m.input.View())
		v.AltScreen = true
		return v
	}

	v := tea.NewView(m.table.View() + "\n" + helpLine)
	v.AltScreen = true
	return v
}
