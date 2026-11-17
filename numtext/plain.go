package numtext

import (
	"unicode"
	"unicode/utf8"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
)

func readText(text string) (Value, Reason) {
	if !utf8.ValidString(text) {
		return Value{}, ReasonInvalidText
	}
	for _, r := range text {
		if unicode.IsControl(r) {
			return Value{}, ReasonControlCharacter
		}
	}
	if utf8.RuneCountInString(text) > maxTextLength {
		return Value{}, ReasonTooLong
	}
	return Value{Type: Text, Text: text}, ""
}

func readBoolean(text string) (Value, Reason) {
	switch text {
	case "true":
		return Value{Type: Boolean, Text: text, Boolean: true}, ""
	case "false":
		return Value{Type: Boolean, Text: text}, ""
	case "":
		return Value{}, ReasonEmpty
	}
	return Value{}, ReasonNotBoolean
}

func looksLikeDate(text string) bool {
	if len(text) != 10 || text[4] != '-' || text[7] != '-' {
		return false
	}
	return allDigits(text[0:4]) && allDigits(text[5:7]) && allDigits(text[8:10])
}

func readDate(text string) (Value, Reason) {
	if text == "" {
		return Value{}, ReasonEmpty
	}
	if !looksLikeDate(text) {
		return Value{}, ReasonNotDate
	}
	date, err := datefn.Parse(text)
	if err != nil {
		return Value{}, ReasonNoSuchDay
	}
	return Value{Type: Date, Text: text, Date: date}, ""
}
