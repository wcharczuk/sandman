package commands

import "github.com/urfave/cli/v3"

func Root() *cli.Command {
	return &cli.Command{
		Name:  "sandctl",
		Usage: "Control sandman servers",
		Commands: []*cli.Command{
			Timers(),
		},
	}
}
