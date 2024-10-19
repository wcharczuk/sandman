package cliutil

import (
	"io"
	"os"
	"strings"
)

// FileOrStdin reads a file or from standard input if the file
// path is the literal string "-".
func FileOrStdin(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}
