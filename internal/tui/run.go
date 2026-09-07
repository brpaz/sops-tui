package tui

import (
	"context"

	"github.com/brpaz/sops-tui/internal/secrets"
)

// Run checks that sops is available, then builds and runs the TUI
// against root until the user quits.
func Run(ctx context.Context, root string) error {
	if err := secrets.CheckAvailable(); err != nil {
		return err
	}

	a, err := New(root)
	if err != nil {
		return err
	}

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			a.app.Stop()
		case <-done:
		}
	}()

	return a.app.Run()
}
