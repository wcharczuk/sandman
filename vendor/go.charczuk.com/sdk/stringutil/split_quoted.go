package stringutil

// SplitQuoted splits a corpus on a given string but treats quoted strings,
// specifically any text within `"` as whole values.
func SplitQuoted(text, sep string) (output []string) {
	if text == "" || sep == "" {
		return
	}

	// generally we read the text rune by rune
	// if we see a rune that is the start of the separator
	// we consume the separator until we either miss
	// or we reach the end of the separator
	//
	// if we miss, we move the separator contents to the normal
	// accumulation working word.
	//
	// if we run out of separator, we collect the previous
	// accumulation working word as an output.
	//
	// we aim generally to do a single pass through the text
	// and prioritize _not_ re-reading the same rune multiple times.

	// fsm states
	const (
		stateWord   = iota
		stateSep    = iota
		stateQuoted = iota
	)

	var state int
	var working, workingSep []rune
	var openingQuote rune
	sepRunes := []rune(sep)
	var sepIndex int
	for _, r := range text {
		switch state {
		case stateWord:
			{
				if r == sepRunes[0] {
					state = stateSep
					workingSep = append(workingSep, r)
					sepIndex = 1
					continue
				}

				working = append(working, r)

				if fieldsIsQuote(r) {
					openingQuote = r
					state = stateQuoted
					continue
				}

				continue
			}

		case stateSep:
			{
				if sepIndex == len(sepRunes) {
					workingSep = nil
					sepIndex = 0

					if len(working) > 0 {
						output = append(output, string(working))
					}

					working = []rune{r}

					if fieldsIsQuote(r) {
						openingQuote = r
						state = stateQuoted
						continue
					}

					state = stateWord
					continue
				}

				if r == sepRunes[sepIndex] {
					workingSep = append(workingSep, r)
					sepIndex++
					continue
				}

				// if we have a separator miss, add
				// whatever we've collected so far to the
				// working word
				working = append(working, workingSep...)
				workingSep = nil
				sepIndex = 0

				working = append(working, r)

				if fieldsIsQuote(r) {
					openingQuote = r
					state = stateQuoted
					continue
				}

				state = stateWord
				continue
			}

		case stateQuoted:
			{
				// if we hit a quote, and it matches the "opening" quote
				// switch back to normal word mode
				if fieldsIsQuote(r) && fieldsMatchesQuote(openingQuote, r) {
					state = stateWord
				}
				working = append(working, r)
				continue
			}

		}
	}

	if len(workingSep) > 0 {
		if string(workingSep) != sep {
			working = append(working, workingSep...)
		}
	}
	if len(working) > 0 {
		output = append(output, string(working))
	}
	return
}
