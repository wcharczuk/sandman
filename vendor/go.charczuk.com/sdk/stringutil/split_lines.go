package stringutil

// SplitLines splits the contents by the ascii control character `\n` with
// the default options applied.
//
// To set options yourself, use `SplitLinesOptions{}.SplitLines(contents)`.
func SplitLines(contents string) []string {
	return DefaultSplitLinesOptions.SplitLines(contents)
}

// DefaultSplitLinesOptions are the default split lines options.
var DefaultSplitLinesOptions SplitLinesOptions

// SplitLinesOptions are options for the SplitLines function.
type SplitLinesOptions struct {
	SkipTrailingNewline bool
	SkipEmptyLines      bool
}

// SplitLines splits a corpus into individual lines by the ascii control character `\n`.
// You can control some behaviors of the splitting process with variadic options.
func (opts SplitLinesOptions) SplitLines(contents string) []string {
	contentRunes := []rune(contents)

	var output []string
	const newline = '\n'

	var line []rune
	var c rune
	for index := 0; index < len(contentRunes); index++ {
		c = contentRunes[index]

		// if we hit a newline
		if c == newline {

			// if we should omit newlines
			if !opts.SkipTrailingNewline {
				line = append(line, c)
			}

			// if we should omit empty lines
			if !opts.SkipTrailingNewline {
				if len(line) == 1 && !opts.SkipEmptyLines {
					line = nil
					continue
				}
			} else {
				if len(line) == 0 && !opts.SkipEmptyLines {
					line = nil
					continue
				}
			}

			// add to the output
			output = append(output, string(line))
			line = nil
			continue
		}

		// add non-newline characters to the line
		line = append(line, c)
		continue
	}

	// add anything left
	if len(line) > 0 {
		output = append(output, string(line))
	}
	return output
}
