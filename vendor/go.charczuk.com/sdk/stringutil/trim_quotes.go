package stringutil

import "strings"

// TrimQuotes trims the leading and trailing characters
// that match the `fieldIsQuote` function.
func TrimQuotes(v string) string {
	return strings.TrimFunc(v, fieldsIsQuote)
}
