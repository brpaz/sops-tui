package app

import (
	"context"

	rootcmd "github.com/brpaz/sops-tui/internal/commands/root"
	hellocmd "github.com/brpaz/sops-tui/internal/commands/hello"
)

// App is the composition root for the sops-tui CLI.
type App struct {
	Info VersionInfo
}

// Option is a functional option for configuring an App.
type Option func(*App)

// WithVersionInfo sets the build-time version metadata.
func WithVersionInfo(info VersionInfo) Option {
	return func(a *App) { a.Info = info }
}

// New constructs an App with the provided options.
func New(opts ...Option) (*App, error) {
	appInstance := &App{
		Info: VersionInfo{
			Version:   "0.0.0-dev",
			Commit:    "n/a",
			BuildDate: "n/a",
		},
	}

	for _, opt := range opts {
		opt(appInstance)
	}

	return appInstance, nil
}

// Run builds the root command and executes it with the provided arguments.
func (app *App) Run(ctx context.Context, args []string) error {
	root := rootcmd.New(
		rootcmd.WithVersion(app.Info.String()),
		rootcmd.WithCommand(hellocmd.New()),
	)

	return root.Run(ctx, args)
}
