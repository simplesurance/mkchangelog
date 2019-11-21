package changelog

import (
	"strings"
)

type DescriptionCreator interface {
	Parse(string) []string
}

// Changelog is a collection of entries that describe changes.
type Changelog struct {
	Entries             []*Entry
	ReleaseName         string
	DescriptionCreators []DescriptionCreator
}

func (c *Changelog) Parse(commitMsg string) {
	spl := strings.Split(commitMsg, "\n")
	entry := Entry{
		Headline: spl[0],
	}

	for _, descCreator := range c.DescriptionCreators {
		descriptions := descCreator.Parse(commitMsg)
		entry.Description = append(entry.Description, descriptions...)
	}

	c.Entries = append(c.Entries, &entry)
}

// Entry describes a particular change.
type Entry struct {
	Headline    string
	Description []string
}
