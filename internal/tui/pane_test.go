package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/sops-tui/internal/secrets"
)

// ageFixture generates a throwaway age keypair, points sops at its
// private key, and returns the public key for a test-local .sops.yaml.
func ageFixture(t *testing.T) (publicKey string) {
	t.Helper()

	out, err := exec.Command("age-keygen").Output()
	require.NoError(t, err, "age-keygen must be installed to run these tests")

	keyFile := filepath.Join(t.TempDir(), "keys.txt")
	require.NoError(t, os.WriteFile(keyFile, out, 0o600))
	t.Setenv("SOPS_AGE_KEY_FILE", keyFile)

	for line := range strings.SplitSeq(string(out), "\n") {
		if pk, ok := strings.CutPrefix(line, "# public key: "); ok {
			return pk
		}
	}
	t.Fatal("could not find public key in age-keygen output")
	return ""
}

func writeSopsConfig(t *testing.T, dir, publicKey string) {
	t.Helper()
	content := "creation_rules:\n  - path_regex: '.*\\.yaml$'\n    age: '" + publicKey + "'\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".sops.yaml"), []byte(content), 0o644))
}

// key builds the tcell key event handleKey expects for a single logical
// keypress, mirroring the strings used throughout these tests: a named
// special key ("enter", "esc", "ctrl+c", "up", "down", "home", "end") or
// a single character rune.
func key(s string) *tcell.EventKey {
	switch s {
	case "enter":
		return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	case "esc":
		return tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	case "ctrl+c":
		return tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)
	case "up":
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case "down":
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case "home":
		return tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone)
	case "end":
		return tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)
	case "tab":
		return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	default:
		r := []rune(s)
		return tcell.NewEventKey(tcell.KeyRune, r[0], tcell.ModNone)
	}
}

// tableRows returns the table's data row paths (excluding the header), in
// display order.
func tableRows(a *App) []string {
	var rows []string
	for i := 1; i < a.table.GetRowCount(); i++ {
		rows = append(rows, a.table.GetCell(i, 0).Text)
	}
	return rows
}

// entryStatus looks up path's status in a.allEntries, the raw scan result
// underlying the table. Status is no longer a visible column (each view
// mode already implies it), so tests assert on the scan result directly.
func entryStatus(t *testing.T, a *App, path string) secrets.Status {
	t.Helper()
	for _, e := range a.allEntries {
		if e.Path == path {
			return e.Status
		}
	}
	t.Fatalf("no scanned entry for path %q", path)
	return secrets.StatusUnknown
}

// typeText feeds s through handleKey one rune at a time, forwarding
// whatever handleKey doesn't consume to the input field's own handler,
// exactly as Application.Run's event loop would.
func typeText(t *testing.T, a *App, s string) {
	t.Helper()
	for _, r := range s {
		if event := a.handleKey(key(string(r))); event != nil {
			a.input.InputHandler()(event, func(tview.Primitive) {})
		}
	}
}

func TestViewAction_ShowsDecryptedContentWithoutModifyingDisk(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))
	require.NoError(t, secrets.Encrypt(path))

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	a, err := New(root)
	require.NoError(t, err)

	a.handleKey(key("v"))

	require.NotNil(t, a.pane)
	require.Contains(t, a.pane.content, "password: hunter2")

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, before, after, "viewing must not modify the file on disk")

	a.handleKey(key("esc"))
	require.Nil(t, a.pane, "esc returns to the list")
}

func TestViewAction_QReturnsToPane(t *testing.T) {
	root := t.TempDir()
	publicKey := ageFixture(t)
	writeSopsConfig(t, root, publicKey)

	path := filepath.Join(root, "secret.yaml")
	require.NoError(t, os.WriteFile(path, []byte("password: hunter2\n"), 0o644))
	require.NoError(t, secrets.Encrypt(path))

	a, err := New(root)
	require.NoError(t, err)

	a.handleKey(key("enter"))
	require.NotNil(t, a.pane)

	a.handleKey(key("q"))
	require.Nil(t, a.pane, "q closes the pane rather than quitting the app")
}

func TestViewAction_PlaintextRowShowsRawContentWithoutSops(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "plain.yaml"), []byte("data: hello\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)
	a.view = viewPlaintext
	a.rebuildRows()

	a.handleKey(key("v"))

	require.NotNil(t, a.pane)
	require.Equal(t, "data: hello\n", a.pane.content, "plaintext rows are read directly, not through sops -d")
}

func TestViewAction_DecryptFailureShowsErrorInPane(t *testing.T) {
	root := t.TempDir()
	// A file that looks encrypted (has a top-level "sops" key so our
	// status detection reports it as Encrypted) but isn't a valid sops
	// document, so `sops -d` itself fails.
	path := filepath.Join(root, "broken.yaml")
	require.NoError(t, os.WriteFile(path, []byte("data: hello\nsops:\n    version: 3\n"), 0o644))

	a, err := New(root)
	require.NoError(t, err)

	a.handleKey(key("v"))

	require.NotNil(t, a.pane)
	require.Contains(t, a.pane.content, "Error:")
}
