package stringutil

import (
	"math/rand"
	"time"
)

var (
	// LowerLetters is a runset of lowercase letters.
	LowerLetters Runeset = []rune(`abcdefghijklmnopqrstuvwxyz`)

	// UpperLetters is a runset of uppercase letters.
	UpperLetters Runeset = []rune(`ABCDEFGHIJKLMNOPQRSTUVWXYZ`)

	// Letters is a runset of both lower and uppercase letters.
	Letters = append(LowerLetters, UpperLetters...)

	// Numbers is a runset of numeric characters.
	Numbers Runeset = []rune(`0123456789`)

	// LettersAndNumbers is a runset of letters and numeric characters.
	LettersAndNumbers = append(Letters, Numbers...)

	// Symbols is a runset of symbol characters.
	Symbols Runeset = []rune(`!@#$%^&*()_+-=[]{}\|:;`)

	// LettersNumbersAndSymbols is a runset of letters, numbers and symbols.
	LettersNumbersAndSymbols = append(LettersAndNumbers, Symbols...)
)

// Runeset is a set of runes
type Runeset []rune

// Len implements part of sorter.
func (rs Runeset) Len() int {
	return len(rs)
}

// Swap implements part of sorter.
func (rs Runeset) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

// Less implements part of sorter.
func (rs Runeset) Less(i, j int) bool {
	return uint16(rs[i]) < uint16(rs[j])
}

var (
	_provider = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Random returns a random selection of runes from the set.
func (rs Runeset) Random(length int) string {
	runes := make([]rune, length)
	for index := range runes {
		runes[index] = rs[_provider.Intn(len(rs)-1)]
	}
	return string(runes)
}

// RandomProvider returns a random selection of runes from the set.
func (rs Runeset) RandomProvider(provider *rand.Rand, length int) string {
	runes := make([]rune, length)
	for index := range runes {
		runes[index] = rs[provider.Intn(len(rs)-1)]
	}
	return string(runes)
}
