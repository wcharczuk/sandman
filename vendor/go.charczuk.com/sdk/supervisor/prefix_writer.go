package supervisor

import (
	"io"
	"strings"
)

type PrefixWriter struct {
	Prefix string
	Writer io.Writer
}

func (pw PrefixWriter) Write(data []byte) (int, error) {
	written := len(data)
	if _, err := pw.Writer.Write([]byte(applyPrefix(pw.Prefix, string(data)))); err != nil {
		return 0, err
	}
	return written, nil
}

func applyPrefix(prefix, data string) string {
	newlines := strings.Count(data, "\n")
	data = strings.Replace(data, "\n", "\n"+prefix, newlines-1)
	return prefix + data
}
