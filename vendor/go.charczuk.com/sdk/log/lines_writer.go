package log

// LinesWriter is a writer that writes
// each `Write` calls data to a backing array.
//
// It is useful for tests and debugging logging calls.
type LinesWriter []string

func (lo *LinesWriter) Write(data []byte) (int, error) {
	*lo = append(*lo, string(data))
	return len(data), nil
}
