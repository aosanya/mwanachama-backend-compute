package problem

import "strings"

const (
	maxShownTextCharacters = 200
	maxShownListItems      = 20
)

type Value struct {
	text   string
	quoted bool
}

func PlainText(received string) Value {
	return Value{text: cutText(received)}
}

func QuotedText(received string) Value {
	return Value{text: cutText(received), quoted: true}
}

func List(items []string) Value {
	return Value{text: showList(items)}
}

func (v Value) Field() string {
	return v.text
}

func (v Value) Sentence() string {
	if v.quoted {
		return `"` + v.text + `"`
	}
	return v.text
}

func cutText(received string) string {
	characters := []rune(received)
	if len(characters) <= maxShownTextCharacters {
		return received
	}
	return string(characters[:maxShownTextCharacters]) + "..."
}

func showList(items []string) string {
	shown := items
	more := false
	if len(items) > maxShownListItems {
		shown = items[:maxShownListItems]
		more = true
	}
	text := strings.Join(shown, ", ")
	if more {
		text += ", ..."
	}
	return "[" + text + "]"
}
