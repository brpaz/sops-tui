package tui

// startInput switches to the given input mode, seeding the input field
// with a prompt and initial value, and moves focus to it.
func (a *App) startInput(mode inputMode, prompt, value string) {
	a.inputMode = mode
	a.input.SetLabel(prompt + " ")
	a.input.SetText(value)
	a.footerPages.SwitchToPage(footerInput)
	a.app.SetFocus(a.input)
}

// cancelInput exits the active input without applying it. A cancelled
// filter reverts to whatever was active before the input was opened.
func (a *App) cancelInput() {
	if a.inputMode == inputFilter {
		a.filterQuery = a.filterPrevQuery
		a.rebuildRows()
	}
	a.inputMode = inputNone
	a.footerPages.SwitchToPage(footerHelp)
	a.app.SetFocus(a.table)
}

// confirmInput applies the active input's value and closes it: a filter
// becomes the new confirmed filterQuery, a command is executed against
// the selected row.
func (a *App) confirmInput() {
	mode := a.inputMode
	value := a.input.GetText()
	a.inputMode = inputNone
	a.footerPages.SwitchToPage(footerHelp)
	a.app.SetFocus(a.table)

	if mode == inputFilter {
		a.filterQuery = value
		a.rebuildRows()
		return
	}

	a.executeCommand(value)
}
