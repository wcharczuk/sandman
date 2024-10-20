package commands

import "github.com/urfave/cli/v3"

func DefaultFlags(moreFlags ...cli.Flag) []cli.Flag {
	return append([]cli.Flag{
		&cli.BoolFlag{
			Name: "debug",
		},
	}, moreFlags...)
}

func DefaultClientFlags(moreFlags ...cli.Flag) []cli.Flag {
	return DefaultFlags(append([]cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Value: "localhost:8888",
		},
		&cli.StringFlag{
			Name:  "authority",
			Value: "sandman-srv.local",
		},
	}, moreFlags...)...)
}
