package main

import (
	"context"
	"os"
	"sandman/sandctl/commands"

	"sandman/pkg/cliutil"
)

func main() {
	if err := commands.Root().Run(context.Background(), os.Args); err != nil {
		cliutil.MaybeFatal(err)
	}
}
