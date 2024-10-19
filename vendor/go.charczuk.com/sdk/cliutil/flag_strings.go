package cliutil

import "flag"

var _ flag.Value = (*FlagStrings)(nil)

// FlagStrings is a type you can use to represent multiple flag
// values that are passed in separately as `--foo=bar --foo=moo`.
type FlagStrings struct {
	Usage  string
	Values []string
}

// String implements [flag.Value].
func (fs *FlagStrings) String() string {
	return fs.Usage
}

// Set implements [flag.Value].
func (fs *FlagStrings) Set(value string) error {
	fs.Values = append(fs.Values, value)
	return nil
}

// HasOrUnset is a helper to test if a value is presesent as a
// passed in flag value.
func (fs *FlagStrings) HasOrUnset(value string) bool {
	if len(fs.Values) == 0 {
		return true
	}
	for _, v := range fs.Values {
		if v == value {
			return true
		}
	}
	return false
}
