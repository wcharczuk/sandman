package web

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"
)

// LocalizationEntry is a specific language translation for a string.
type LocalizationEntry struct {
	Tag     string
	Key     string
	Message any
}

// Localization holds the translation catalog and lets
// us create a printer to emit translated strings.
type Localization struct {
	entries []LocalizationEntry
	catalog *catalog.Builder
}

// AddEntries adds entries to the localization catalog.
func (l *Localization) AddEntries(entries ...LocalizationEntry) {
	l.entries = append(l.entries, entries...)
}

// Translations are language => english => translated
type Translations map[string]map[string]string

// AddYAML reads a given reader as yaml and adds translation entries.
func (l *Localization) AddYAML(yamlContents []byte) {
	var translations Translations
	if err := yaml.Unmarshal(yamlContents, &translations); err != nil {
		panic(fmt.Errorf("cannot decode localization translation yaml: %w", err))
	}
	for tag, translation := range translations {
		for from, to := range translation {
			l.entries = append(l.entries, LocalizationEntry{
				Tag:     tag,
				Key:     from,
				Message: to,
			})
		}
	}
}

// InitializeLocalization parses the localization entries
func (l *Localization) Initialize() error {
	l.catalog = catalog.NewBuilder()
	for _, e := range l.entries {
		tag, err := language.Parse(e.Tag)
		if err != nil {
			return err
		}
		switch msg := e.Message.(type) {
		case string:
			l.catalog.SetString(tag, e.Key, msg)
		case catalog.Message:
			l.catalog.Set(tag, e.Key, msg)
		case []catalog.Message:
			l.catalog.Set(tag, e.Key, msg...)
		default:
			return fmt.Errorf("invalid message type %T", e.Message)
		}
	}
	return nil
}

// Printer yields a printer based on the localization catalog
// for a given language tag.
//
// You _must_ call `.Initialize()` before using this method!!
func (l *Localization) Printer(preferred ...string) Printer {
	if l.catalog == nil {
		panic("localization; called `Printer` before `Initialize()`")
	}
	t, _ := language.MatchStrings(l.catalog.Matcher(), preferred...)
	p := message.NewPrinter(t, message.Catalog(l.catalog))
	return &messagePrinter{p}
}

type messagePrinter struct {
	p *message.Printer
}

func (mp *messagePrinter) Printf(format string, args ...any) string {
	return mp.p.Sprintf(format, args...)
}
func (mp *messagePrinter) Print(args ...any) string {
	return mp.p.Sprintf(fmt.Sprint(args...))
}

type Printer interface {
	Printf(string, ...any) string
	Print(...any) string
}
