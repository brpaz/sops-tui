package hello

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

const (
	name  = "hello"
	usage = "Say hello"
)

// New returns the hello subcommand.
func New() *cli.Command {
	return &cli.Command{
		Name:    name,
		Aliases: []string{"h"},
		Usage:   usage,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "name to greet",
				Value:   "World",
			},
		},
		Action: run,
	}
}

func run(_ context.Context, cmd *cli.Command) error {
	fmt.Fprintf(cmd.Writer, "Hello, %s!\n", cmd.String("name"))
	return nil
}
