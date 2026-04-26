package dbutil

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// ErrDatabaseNameReserved is a validation failure.
	ErrDatabaseNameReserved = errors.New("dbutil; database name is reserved")
	// ErrDatabaseNameEmpty is a validation failure.
	ErrDatabaseNameEmpty = errors.New("dbutil; database name is empty")
	// ErrDatabaseNameInvalidFirstRune is a validation failure.
	ErrDatabaseNameInvalidFirstRune = errors.New("dbutil; database name must start with a letter or underscore")
	// ErrDatabaseNameInvalid is a validation failure.
	ErrDatabaseNameInvalid = errors.New("dbutil; database name must be composed of (in regex form) [a-zA-Z0-9_]")
	// ErrDatabaseNameTooLong is a validation failure.
	ErrDatabaseNameTooLong = errors.New("dbutil; database name must be 63 characters or fewer")
)

var (
	// ReservedDatabaseNames are names you cannot use to create a database with.
	ReservedDatabaseNames = []string{
		"postgres",
		"defaultdb",
		"template0",
		"template1",
	}
)

const (
	// DatabaseNameMaxLength is the maximum length of a database name.
	DatabaseNameMaxLength = 63
)

// ValidateDatabaseName validates a database name.
func ValidateDatabaseName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrDatabaseNameEmpty
	}
	if len(name) > DatabaseNameMaxLength {
		return ErrDatabaseNameTooLong
	}

	firstRune, _ := utf8.DecodeRuneInString(name)
	if !isValidDatabaseNameFirstRune(firstRune) {
		return ErrDatabaseNameInvalidFirstRune
	}

	for _, r := range name {
		if !isValidDatabaseNameRune(r) {
			return ErrDatabaseNameInvalid
		}
	}

	for _, reserved := range ReservedDatabaseNames {
		if strings.EqualFold(reserved, name) {
			return ErrDatabaseNameReserved
		}
	}
	return nil
}

// isValidDatabaseNameFirstRune returns if the rune is valid as a first rune of a database name.
func isValidDatabaseNameFirstRune(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// isValidDatabaseNameRune is a rune predicate that indicites a rune is a valid database name component.
func isValidDatabaseNameRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
