package stringutil

import "unicode"

// Fields splits a corpus on space but treats quoted strings,
// specifically any text within `"` as individual fields.
func Fields(text string) (output []string) {
	if text == "" {
		return
	}

	// fsm states
	const (
		stateLeadingSpace    = iota
		stateWord            = iota
		stateIntraSpace      = iota
		stateLeadingQuoted   = iota
		stateIntraWordQuoted = iota
	)

	var state int
	var word []rune
	var opened rune
	for _, r := range text {
		switch state {
		case stateLeadingSpace: //leading whitespace until quote or alpha
			if !unicode.IsSpace(r) {
				if fieldsIsQuote(r) { // start a quoted section
					opened = r
					state = stateLeadingQuoted
				} else {
					state = stateWord
					word = append(word, r)
				}
			}
		case stateWord: // within a word
			if fieldsIsQuote(r) {
				opened = r
				word = append(word, r)
				state = stateIntraWordQuoted
			} else if unicode.IsSpace(r) {
				if len(word) > 0 {
					output = append(output, string(word))
					word = nil
				}
				state = stateIntraSpace
			} else {
				word = append(word, r)
			}
		case stateIntraSpace: // we've seen a space after we've seen at least one word
			// consume spaces until a non-space character
			if !unicode.IsSpace(r) {
				if fieldsIsQuote(r) { // start a quoted section
					opened = r
					state = stateLeadingQuoted
				} else {
					state = stateWord
					word = append(word, r)
				}
			}
		case stateLeadingQuoted: // leading quoted section
			// if we close a quoted section, switch
			// back to normal word mode
			if fieldsMatchesQuote(opened, r) {
				state = stateWord
			} else {
				word = append(word, r)
			}
		case stateIntraWordQuoted: // quoted section within a word
			// if we close a quoted section, switch
			// back to normal word mode
			if fieldsMatchesQuote(opened, r) {
				state = stateWord
			}
			word = append(word, r)
		}
	}

	if len(word) > 0 {
		output = append(output, string(word))
	}
	return
}

func fieldsIsQuote(r rune) bool {
	return r == '"' ||
		r == '\'' ||
		r == '“' ||
		r == '”' ||
		r == '`' ||
		r == '‘' ||
		r == '’'
}

func fieldsMatchesQuote(a, b rune) bool {
	if a == '“' && b == '”' {
		return true
	}
	if a == '”' && b == '“' {
		return true
	}
	if a == '‘' && b == '’' {
		return true
	}
	if a == '’' && b == '‘' {
		return true
	}
	return a == b
}
