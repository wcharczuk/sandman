package cliutil

import (
	"fmt"
	"os"
)

// MaybeFatal exits the program with a code 1 printing an error
// to stderr if the passed err is not nil, otherwise it does nothing.
func MaybeFatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
