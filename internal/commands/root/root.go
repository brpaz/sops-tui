package root

import (
	"context"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

const (
	Name      = "sops-tui"
	usage     = "A k9s-style terminal UI for browsing, encrypting, and decrypting SOPS-protected secret files"
	pathArg   = "path"
	pathUsage = "Directory to scan for SOPS-managed files (default: current directory)"
)

// RunFunc launches the application against the resolved scan root.
type RunFunc func(ctx context.Context, rootPath string) error

// options holds the configuration for the root command.
type options struct {
	version  string
	commands []*cli.Command
	run      RunFunc
}

// Option is a functional option for configuring the root command.
type Option func(*options)

// WithVersion sets the version string shown by --version.
func WithVersion(v string) Option {
	return func(o *options) { o.version = v }
}

// WithCommand appends a single sub-command to the root command.
// Call it multiple times to register multiple sub-commands.
func WithCommand(cmd *cli.Command) Option {
	return func(o *options) { o.commands = append(o.commands, cmd) }
}

// WithRunFunc sets the function invoked with the resolved scan root when
// the root command runs.
func WithRunFunc(fn RunFunc) Option {
	return func(o *options) { o.run = fn }
}

// New returns the root *cli.Command with the supplied options applied.
func New(opts ...Option) *cli.Command {
	o := &options{
		version: "0.0.0-dev",
	}
	for _, opt := range opts {
		opt(o)
	}

	return &cli.Command{
		Name:                  Name,
		Version:               o.version,
		Usage:                 usage,
		EnableShellCompletion: true,
		Commands:              o.commands,
		Arguments: []cli.Argument{
			&cli.StringArg{Name: pathArg, UsageText: pathUsage},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			path, err := ResolvePath(cmd.StringArg(pathArg))
			if err != nil {
				return err
			}
			if o.run == nil {
				return nil
			}
			return o.run(ctx, path)
		},
	}
}

// ResolvePath resolves the scan root passed as the command's positional
// argument to an absolute path, defaulting to the current working directory
// when arg is empty.
func ResolvePath(arg string) (string, error) {
	if arg == "" {
		return os.Getwd()
	}
	return filepath.Abs(arg)
}
