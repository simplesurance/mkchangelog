package markdown

// TODO normalize text, or escape characters

import (
	"strings"
)

type Document struct {
	elements []Element
}

func (d *Document) AppendElement(e Element) {
	d.elements = append(d.elements, e)
}

func (d *Document) String() string {
	var result string

	for _, e := range d.elements {
		result += e.Markdown()
	}

	return strings.TrimPrefix(result, "\n")
}

type Element interface {
	Markdown() string
}

var textSpecialChars = []string{
	"\\",
	"`",
	"*",
	"{",
	"}",
	"[",
	"]",
	"(",
	")",
	"!",
	"|",
}

func EscapeText(in string) string {
	result := in

	// ineffecient but sufficient for now
	for _, char := range textSpecialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}

type HeadingLevel int

const (
	HeadingLvl1 HeadingLevel = iota + 1
	HeadingLvl2
	HeadingLvl3
	HeadingLvl4
	HeadingLvl5
	HeadingLvl6
)

type Heading struct {
	Level HeadingLevel
	Text  string
}

func (h *Heading) Markdown() string {
	var result string

	// TODO: Do not add a newline if it's the beginning of the document
	result += "\n"

	for i := 0; i < int(h.Level); i++ {
		result += "#"
	}

	result += " " + EscapeText(h.Text) + "\n"
	return result
}

type List struct {
	Entries []Element
}

func (l *List) Markdown() string {
	result := "\n"

	for _, entry := range l.Entries {
		result += entry.Markdown() + "\n"
	}

	return result
}

type ListItemLevel int

const (
	ListItemLevel1 ListItemLevel = iota + 1
	ListItemLevel2
	ListItemLevel3
	ListItemLevel4
)

type ListItem struct {
	Content Element
	Lvl     ListItemLevel
}

func (l *ListItem) Markdown() string {
	var result string

	for i := 0; i < (int(l.Lvl)-1)*2; i++ {
		result += " "
	}
	result += "* "

	result += l.Content.Markdown()

	return result
}

type Text struct {
	Text string
}

func (t *Text) Markdown() string {
	return EscapeText(t.Text)
}

type Link struct {
	Text string
	URL  string
}

func (l *Link) Markdown() string {
	if l.Text == "" {
		return "<" + l.URL + ">"
	}

	return "[" + EscapeText(l.Text) + "](" + l.URL + ")"
}
