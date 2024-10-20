package main

import (
	"context"
	"os"
	"sandman/sandctl/commands"

	"go.charczuk.com/sdk/cliutil"
)

func main() {
	if err := commands.Root().Run(context.Background(), os.Args); err != nil {
		cliutil.Fatal(err)
	}
}
