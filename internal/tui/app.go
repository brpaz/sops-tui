// Package tui implements the sops-tui terminal application: a flat,
// k9s-style table of SOPS-managed files with vim-style navigation.
package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/brpaz/sops-tui/internal/secrets"
)

const paneHelpLine = "esc/q: back to list"

// screenMarginX/Y keep every page off the terminal's own border.
const (
	screenMarginX = 2
	screenMarginY = 1
)

// logoArt is a k9s-style ASCII wordmark shown in the header's top-right
// corner. logoWidth/logoHeight size the header column and box that hold
// it, since a TextView won't report the width of its own art.
const logoArt = `█████  ███  ████  █████
█     █   █ █   █ █
█████ █   █ ████  █████
    █ █   █ █         █
█████  ███  █     █████`

const (
	logoWidth  = 24
	logoHeight = 5

	// headerHeight fits logoHeight plus the header box's own top/bottom
	// border (2) and inside padding (2).
	headerHeight = logoHeight + 4
)

// Catppuccin Macchiato palette (https://catppuccin.com/palette), used
// throughout for the k9s-style chrome. accentTag/hotkeyKeyTag are the
// matching tview hex color tags (used in dynamic-color text) kept in sync
// with their tcell.Color counterparts by hand, since tview has no reverse
// lookup from Color to tag string.
const ctpRGB = tcell.ColorIsRGB | tcell.ColorValid

const (
	ctpMauve    = ctpRGB | 0xc6a0f6
	ctpBlue     = ctpRGB | 0x8aadf4
	ctpText     = ctpRGB | 0xcad3f5
	ctpBase     = ctpRGB | 0x24273a
	ctpSurface0 = ctpRGB | 0x363a4f

	accentColor   = ctpMauve
	accentTag     = "#c6a0f6"
	headerBgColor = ctpSurface0
	headerFgColor = ctpText
	selectedBg    = ctpBlue
	selectedFg    = ctpBase

	hotkeyKeyColor = "#eed49f"
)

const (
	pageList    = "list"
	pageHelp    = "help"
	pageConfirm = "confirm"
	pagePane    = "pane"
	pageError   = "error"

	footerHelp  = "help"
	footerInput = "input"
)

// inputMode identifies which single-line input, if any, is currently
// capturing keystrokes instead of the table.
type inputMode int

const (
	inputNone inputMode = iota
	inputFilter
	inputCommand
)

// viewMode selects which status subset of allEntries the table shows.
// Tab cycles through them, k9s-style, independently of filterQuery.
type viewMode int

const (
	viewEncrypted viewMode = iota
	viewPlaintext
	viewAll
)

// next returns the view that Tab switches to from v.
func (v viewMode) next() viewMode {
	return (v + 1) % 3
}

func (v viewMode) String() string {
	switch v {
	case viewPlaintext:
		return "plaintext"
	case viewAll:
		return "all"
	default:
		return "encrypted"
	}
}

// confirmPane is a pending encrypt/decrypt action awaiting a y/n
// confirmation, rendered as its own full-screen page describing the
// specific action before anything touches disk.
type confirmPane struct {
	kind string // "encrypt" or "decrypt"
	path string
}

// App is the root tview application for sops-tui.
type App struct {
	root string

	app         *tview.Application
	pages       *tview.Pages
	table       *tview.Table
	footerPages *tview.Pages
	input       *tview.InputField
	listFlex    *tview.Flex

	headerFlex *tview.Flex
	headerLeft *tview.TextView
	headerKeys *tview.TextView
	headerLogo *tview.TextView
	filterBar  *tview.TextView
	hotkeys    *tview.TextView

	helpView    *tview.TextView
	confirmView *tview.TextView
	paneView    *tview.TextView
	errorView   *tview.TextView

	// allEntries is the raw result of the last scan; the table shows it
	// filtered by view and filterQuery.
	allEntries []secrets.FileEntry
	// view selects which status subset of allEntries is shown; Tab
	// cycles it independently of filterQuery.
	view viewMode

	// filterQuery is the confirmed filter substring (case-insensitive,
	// matched against the path column), or "" for no filter.
	filterQuery string
	// filterPrevQuery snapshots filterQuery when entering filter input,
	// so esc can restore it.
	filterPrevQuery string
	// inputMode is which single-line input is active, if any.
	inputMode inputMode

	// showHelp, when true, replaces the list with the keybinding overlay.
	showHelp bool
	// pane, when non-nil, is a full-screen overlay (e.g. decrypted file
	// content) shown instead of the list.
	pane *pane
	// confirm, when non-nil, is a pending encrypt/decrypt action awaiting
	// a yes/no confirmation before it touches disk.
	confirm *confirmPane
	// fatalErr replaces the table view entirely, e.g. the scan root
	// could not be read.
	fatalErr error

	// suspend runs f with the terminal UI suspended, e.g. to run `sops
	// <file>` in the foreground. It wraps app.Suspend in production; tests
	// override it to call f directly, since Suspend is a no-op before the
	// application's screen has been initialized by Run.
	suspend func(f func())
	// edit performs the actual `sops <file>` edit for startEdit. It wraps
	// secrets.Edit in production; tests override it to simulate an edit's
	// outcome without shelling out to a real editor.
	edit func(path string) error
}

// New builds the TUI app for the given scan root. It requires the sops
// binary to already be available on PATH; callers should check that via
// secrets.CheckAvailable before calling New.
func New(root string) (*App, error) {
	// tview's default theme fills every primitive's background with a
	// dark slate color; resetting it to ColorDefault before building any
	// primitives lets them show the terminal's own background instead,
	// so only the chrome we explicitly color (filter bar, table border/
	// header, selection) stands out.
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorDefault

	a := &App{root: root, app: tview.NewApplication()}
	a.suspend = func(f func()) { a.app.Suspend(f) }
	a.edit = secrets.Edit

	a.buildHeader()
	a.buildTable()
	a.buildFooter()
	a.buildOverlays()

	a.listFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.headerFlex, headerHeight, 0, false).
		AddItem(a.filterBar, 1, 0, false).
		AddItem(tview.NewBox(), 1, 0, false). // breathing room before the table
		AddItem(a.table, 0, 1, true).
		AddItem(a.footerPages, 2, 0, false)

	a.pages = tview.NewPages().
		AddPage(pageList, a.listFlex, true, true).
		AddPage(pageHelp, a.helpView, true, false).
		AddPage(pageConfirm, a.confirmView, true, false).
		AddPage(pagePane, a.paneView, true, false).
		AddPage(pageError, a.errorView, true, false)

	// Every page sits inside a fixed screen margin so nothing touches the
	// terminal's own border.
	screenRoot := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), screenMarginY, 0, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(tview.NewBox(), screenMarginX, 0, false).
				AddItem(a.pages, 0, 1, true).
				AddItem(tview.NewBox(), screenMarginX, 0, false),
			0, 1, true,
		).
		AddItem(tview.NewBox(), screenMarginY, 0, false)

	a.app.SetRoot(screenRoot, true).SetFocus(a.table)
	a.app.SetInputCapture(a.handleKey)

	if err := a.refresh(); err != nil {
		return nil, err
	}
	a.render()

	return a, nil
}

func (a *App) buildHeader() {
	a.headerLeft = tview.NewTextView().SetDynamicColors(true)

	a.headerKeys = tview.NewTextView().SetDynamicColors(true)

	a.headerLogo = tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight)
	a.headerLogo.SetText(fmt.Sprintf("[%s]%s[-]", accentTag, logoArt))

	a.headerFlex = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.headerLeft, 0, 1, false).
		AddItem(a.headerKeys, 0, 1, false).
		AddItem(a.headerLogo, logoWidth, 0, false)
	a.headerFlex.SetBorder(true).SetBorderColor(accentColor)
	a.headerFlex.SetBorderPadding(1, 1, 2, 2)
}

// updateHeaderStats refreshes the header's left info panel: title,
// subtitle, scan root, and the overall (unfiltered) encrypted/total
// counts, k9s-style.
func (a *App) updateHeaderStats() {
	total := len(a.allEntries)
	encrypted := 0
	for _, e := range a.allEntries {
		if e.Status == secrets.StatusEncrypted {
			encrypted++
		}
	}
	a.headerLeft.SetText(fmt.Sprintf(
		"[%s::b]sops-tui[-:-:-]\nSops Browser\n\nRoot: %s\n%d encrypted, %d total",
		accentTag, a.root, encrypted, total,
	))
}

// selectedActionKey returns the hotkey and label for whichever of
// encrypt/decrypt applies to the currently selected row, or "" for both
// when neither does (e.g. nothing selected, or its status is unknown).
func (a *App) selectedActionKey() (hotkey, label string) {
	switch a.selectedStatus() {
	case secrets.StatusPlaintext:
		return "e", "Encrypt"
	case secrets.StatusEncrypted:
		return "d", "Decrypt"
	default:
		return "", ""
	}
}

// updateHeaderKeys refreshes the header's middle column with a short,
// k9s-style vertical list of the most relevant shortcuts for the current
// selection; the full list stays in the footer and the "?" help overlay.
func (a *App) updateHeaderKeys() {
	var lines []string
	if hotkey, label := a.selectedActionKey(); hotkey != "" {
		lines = append(lines, headerKeyLine(hotkey, label))
	}
	lines = append(lines,
		headerKeyLine("v/enter", "View"),
		headerKeyLine("E", "Edit"),
		headerKeyLine("tab", fmt.Sprintf("View: %s", a.view)),
		headerKeyLine("?", "Help"),
	)
	a.headerKeys.SetText(strings.Join(lines, "\n"))
}

func headerKeyLine(hotkey, label string) string {
	return fmt.Sprintf("[%s::b]<%s>[-:-:-] %s", hotkeyKeyColor, hotkey, label)
}

func (a *App) buildTable() {
	a.table = tview.NewTable().
		SetSelectable(true, false).
		SetFixed(1, 0)
	a.table.SetSelectedStyle(tcell.StyleDefault.Background(selectedBg).Foreground(selectedFg))
	a.table.SetBorder(true).SetBorderColor(accentColor).SetTitleAlign(tview.AlignLeft)
	a.table.SetBorderPadding(1, 1, 2, 2)
	a.setTableHeader()
}

func (a *App) setTableHeader() {
	// PATH gets expansion so its (and the header's) background fills the
	// full row width regardless of how short the visible paths are. There
	// is no Status column: each view mode (encrypted/plaintext/all)
	// already tells you the status of what you're looking at.
	a.table.SetCell(0, 0, headerCell("PATH", 1))
}

func headerCell(text string, expansion int) *tview.TableCell {
	return tview.NewTableCell(text).
		SetSelectable(false).
		SetTextColor(headerFgColor).
		SetBackgroundColor(headerBgColor).
		SetAttributes(tcell.AttrBold).
		SetExpansion(expansion)
}

// updateTableTitle refreshes the table's k9s-style border title with the
// current view mode and row count, e.g. " files(encrypted)[3] ".
func (a *App) updateTableTitle(n int) {
	a.table.SetTitle(fmt.Sprintf(" files(%s)[%d] ", a.view, n))
}

func (a *App) buildFooter() {
	a.hotkeys = tview.NewTextView().SetDynamicColors(true)
	a.updateHotkeys()

	a.input = tview.NewInputField()
	a.input.SetChangedFunc(func(text string) {
		if a.inputMode == inputFilter {
			a.filterQuery = text
			a.rebuildRows()
		}
	})

	a.footerPages = tview.NewPages().
		AddPage(footerHelp, a.hotkeys, true, true).
		AddPage(footerInput, a.input, true, false)
}

// updateHotkeys rebuilds the single-row hotkey bar. Called from render on
// every keypress, since which of <e>/<d> applies depends on the currently
// selected row's status.
func (a *App) updateHotkeys() {
	pairs := [][2]string{
		{"j/k", "Up/Down"},
		{"gg/G", "Top/Bottom"},
		{"v/enter", "View"},
	}
	if hotkey, label := a.selectedActionKey(); hotkey != "" {
		pairs = append(pairs, [2]string{hotkey, label})
	}
	pairs = append(pairs,
		[2]string{"E", "Edit"},
		[2]string{"tab", fmt.Sprintf("View: %s", a.view)},
		[2]string{"/", "Filter"},
		[2]string{":", "Command"},
		[2]string{"r", "Refresh"},
		[2]string{"?", "Help"},
		[2]string{"q", "Quit"},
	)

	var b strings.Builder
	for i, kv := range pairs {
		if i > 0 {
			b.WriteString("  ")
		}
		fmt.Fprintf(&b, "[%s::b]<%s>[-:-:-] %s", hotkeyKeyColor, kv[0], kv[1])
	}
	a.hotkeys.SetText(b.String())
}

func (a *App) buildOverlays() {
	a.helpView = tview.NewTextView().SetText(helpOverlay)
	a.confirmView = tview.NewTextView()
	a.paneView = tview.NewTextView()
	a.errorView = tview.NewTextView()

	a.filterBar = tview.NewTextView().SetDynamicColors(true)
	a.updateFilterBar()
}

// updateFilterBar refreshes the persistent filter-status line shown above
// the table, independent of the crumbs bar's summarized mode string.
func (a *App) updateFilterBar() {
	query := "(none)"
	if a.filterQuery != "" {
		query = a.filterQuery
	}
	a.filterBar.SetText(fmt.Sprintf(" [%s::b]Filter:[-:-:-] %s", accentTag, query))
}

// refresh re-scans the root directory and repopulates the table, applying
// the current view mode and any active filter.
func (a *App) refresh() error {
	entries, err := secrets.List(a.root)
	if err != nil {
		return fmt.Errorf("scanning %s: %w", a.root, err)
	}

	a.allEntries = entries
	a.updateHeaderStats()
	a.rebuildRows()

	return nil
}

// visibleEntries returns allEntries filtered by the current view mode and
// filterQuery.
func (a *App) visibleEntries() []secrets.FileEntry {
	needle := strings.ToLower(a.filterQuery)

	var out []secrets.FileEntry
	for _, e := range a.allEntries {
		switch a.view {
		case viewEncrypted:
			if e.Status != secrets.StatusEncrypted {
				continue
			}
		case viewPlaintext:
			if e.Status != secrets.StatusPlaintext {
				continue
			}
		}
		if needle != "" && !strings.Contains(strings.ToLower(e.Path), needle) {
			continue
		}
		out = append(out, e)
	}
	return out
}

// rebuildRows recomputes the table's data rows from allEntries under the
// current view and filterQuery settings, preserving the selection when
// possible and recovering it to the top row when it isn't.
func (a *App) rebuildRows() {
	entries := a.visibleEntries()

	for a.table.GetRowCount() > 1 {
		a.table.RemoveRow(a.table.GetRowCount() - 1)
	}
	for i, e := range entries {
		row := i + 1
		a.table.SetCell(row, 0, tview.NewTableCell(e.Path).SetExpansion(1).SetReference(e.Status))
	}

	n := len(entries)
	row, col := a.table.GetSelection()
	switch {
	case n == 0:
		a.table.Select(0, 0)
	case row < 1 || row > n:
		a.table.Select(1, col)
	}

	a.updateTableTitle(n)
	a.updateFilterBar()
}

// render picks which page is visible, in priority order: a fatal error
// overrides everything, then the help overlay, then a pending encrypt/
// decrypt confirmation, then a content pane, and finally the list itself
// (whose footer independently shows the help line or an active
// filter/command input).
func (a *App) render() {
	a.updateHotkeys()
	a.updateHeaderKeys()

	switch {
	case a.fatalErr != nil:
		a.errorView.SetText(fmt.Sprintf("Error: %v\n\npress q to quit", a.fatalErr))
		a.pages.SwitchToPage(pageError)
	case a.showHelp:
		a.pages.SwitchToPage(pageHelp)
	case a.confirm != nil:
		a.confirmView.SetText(confirmText(a.confirm))
		a.pages.SwitchToPage(pageConfirm)
	case a.pane != nil:
		a.paneView.SetText(a.pane.path + "\n\n" + a.pane.content + "\n" + paneHelpLine)
		a.pages.SwitchToPage(pagePane)
	default:
		a.pages.SwitchToPage(pageList)
	}
}

// handleKey is the application's single input entry point, installed via
// Application.SetInputCapture. It mirrors the priority order of render:
// a full-screen overlay (help, confirm, pane) swallows every key except
// the ones that dismiss it; otherwise keys drive list navigation and
// actions. Returning nil consumes the event; returning it unchanged lets
// it fall through to the focused primitive (used for text input).
func (a *App) handleKey(event *tcell.EventKey) *tcell.EventKey {
	key := keyString(event)

	if key == "ctrl+c" {
		a.app.Stop()
		return nil
	}

	if a.showHelp {
		switch key {
		case "esc", "q", "?":
			a.showHelp = false
		}
		a.render()
		return nil
	}

	if a.inputMode != inputNone {
		switch key {
		case "esc":
			a.cancelInput()
			a.render()
			return nil
		case "enter":
			a.confirmInput()
			a.render()
			return nil
		}
		return event
	}

	if a.confirm != nil {
		switch key {
		case "y":
			a.confirmYes()
		case "n", "esc":
			a.confirm = nil
		}
		a.render()
		return nil
	}

	if a.pane != nil {
		switch key {
		case "esc", "q", "enter":
			a.pane = nil
		}
		a.render()
		return nil
	}

	switch key {
	case "q":
		a.app.Stop()
		return nil
	case "?":
		a.showHelp = true
	case "/":
		a.filterPrevQuery = a.filterQuery
		a.startInput(inputFilter, "/", a.filterQuery)
		return nil
	case ":":
		a.startInput(inputCommand, ":", "")
		return nil
	case "r":
		if err := a.refresh(); err != nil {
			a.fatalErr = err
		}
	case "tab":
		a.view = a.view.next()
		a.rebuildRows()
	case "v", "enter":
		if path := a.selectedPath(); path != "" {
			a.pane = a.viewPane(path)
		}
	case "e":
		a.startEncrypt()
	case "E":
		a.startEdit()
	case "d":
		a.startDecrypt()
	case "j", "down":
		a.moveSelection(1)
	case "k", "up":
		a.moveSelection(-1)
	case "g", "home":
		a.selectTop()
	case "G", "end":
		a.selectBottom()
	default:
		return event
	}

	a.render()
	return nil
}

// keyString normalizes a key event to a short identifier: a literal rune
// for character keys, or a named string for special keys.
func keyString(event *tcell.EventKey) string {
	switch event.Key() {
	case tcell.KeyEnter:
		return "enter"
	case tcell.KeyEscape:
		return "esc"
	case tcell.KeyCtrlC:
		return "ctrl+c"
	case tcell.KeyUp:
		return "up"
	case tcell.KeyDown:
		return "down"
	case tcell.KeyHome:
		return "home"
	case tcell.KeyEnd:
		return "end"
	case tcell.KeyTab:
		return "tab"
	case tcell.KeyRune:
		return string(event.Rune())
	default:
		return ""
	}
}
