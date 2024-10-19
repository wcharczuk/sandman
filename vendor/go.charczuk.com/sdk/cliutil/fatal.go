package cliutil

import (
	"fmt"
	"os"
)

// Fatal exits the program with a code 1 printing an error
// to stderr if it's set.
//
// If the error is nil it does nothing.
func Fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
