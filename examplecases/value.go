package examplecases

import (
	"strings"
	"unicode/utf8"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
)

const (
	maxIntegerDigits = 18
	maxDecimalDigits = 30
	maxTextLength    = 200
)

type Value struct {
	Type  string
	Text  string
	Items []string
}

const (
	reasonBad     = "bad"
	reasonTooLong = "too long"
)

var scalarTypes = map[string]bool{
	"int": true, "dec": true, "float": true, "text": true, "bool": true, "date": true,
}

var listTypes = map[string]string{
	"list<int>": "int", "list<dec>": "dec", "list<text>": "text",
}

func (v Value) String() string {
	if element, isList := listTypes[v.Type]; isList {
		items := make([]string, len(v.Items))
		for i, item := range v.Items {
			items[i] = writeScalar(element, item)
		}
		return v.Type + ":[" + strings.Join(items, ", ") + "]"
	}
	return v.Type + ":" + writeScalar(v.Type, v.Text)
}

func writeScalar(kind string, text string) string {
	if kind != "text" {
		return text
	}
	escaped := strings.ReplaceAll(text, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

func parseTyped(text string) (Value, string) {
	kind, body, found := strings.Cut(text, ":")
	if !found {
		return Value{}, reasonBad
	}
	if scalarTypes[kind] {
		decoded, reason := parseScalar(kind, body)
		return Value{Type: kind, Text: decoded}, reason
	}
	element, isList := listTypes[kind]
	if !isList {
		return Value{}, reasonBad
	}
	items, reason := parseList(element, body)
	return Value{Type: kind, Items: items}, reason
}

func parseList(element string, body string) ([]string, string) {
	if len(body) < 2 || body[0] != '[' || body[len(body)-1] != ']' {
		return nil, reasonBad
	}
	inner := body[1 : len(body)-1]
	if strings.TrimSpace(inner) == "" {
		return []string{}, ""
	}
	parts, ok := splitOutside(inner, ", ")
	if !ok {
		return nil, reasonBad
	}
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		decoded, reason := parseScalar(element, part)
		if reason != "" {
			return nil, reason
		}
		items = append(items, decoded)
	}
	return items, ""
}

func parseScalar(kind string, body string) (string, string) {
	switch kind {
	case "int":
		return digitsValue(body, false, maxIntegerDigits)
	case "dec", "float":
		return digitsValue(body, true, maxDecimalDigits)
	case "bool":
		if body == "true" || body == "false" {
			return body, ""
		}
		return "", reasonBad
	case "date":
		if _, err := datefn.Parse(body); err != nil || len(body) != 10 {
			return "", reasonBad
		}
		return body, ""
	case "text":
		return textValue(body)
	}
	return "", reasonBad
}

func digitsValue(body string, allowDot bool, limit int) (string, string) {
	rest := strings.TrimPrefix(body, "-")
	whole, fraction, hasDot := strings.Cut(rest, ".")
	if whole == "" || (hasDot && (!allowDot || fraction == "")) {
		return "", reasonBad
	}
	if !allDigits(whole) || !allDigits(fraction) {
		return "", reasonBad
	}
	if len(whole)+len(fraction) > limit {
		return "", reasonTooLong
	}
	return body, ""
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func textValue(body string) (string, string) {
	decoded, rest, ok := readQuoted(body)
	if !ok || rest != "" {
		return "", reasonBad
	}
	if utf8.RuneCountInString(decoded) > maxTextLength {
		return "", reasonTooLong
	}
	return decoded, ""
}

func readQuoted(s string) (string, string, bool) {
	if s == "" || s[0] != '"' {
		return "", "", false
	}
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i+1 >= len(s) || (s[i+1] != '"' && s[i+1] != '\\') {
				return "", "", false
			}
			b.WriteByte(s[i+1])
			i++
		case '"':
			return b.String(), s[i+1:], true
		default:
			b.WriteByte(s[i])
		}
	}
	return "", "", false
}

func splitOutside(s string, separator string) ([]string, bool) {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '"':
			_, rest, ok := readQuoted(s[i:])
			if !ok {
				return nil, false
			}
			i = len(s) - len(rest) - 1
		case s[i] == '[':
			depth++
		case s[i] == ']':
			depth--
		case depth == 0 && strings.HasPrefix(s[i:], separator):
			parts = append(parts, s[start:i])
			i += len(separator) - 1
			start = i + 1
		}
	}
	if depth != 0 {
		return nil, false
	}
	return append(parts, s[start:]), true
}
