package tui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/brpaz/sops-tui/internal/secrets"
)

// Run checks that sops is available, then builds and runs the TUI
// against root until the user quits.
func Run(ctx context.Context, root string) error {
	if err := secrets.CheckAvailable(); err != nil {
		return err
	}

	m, err := New(root)
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(m, tea.WithContext(ctx)).Run()
	return err
}
