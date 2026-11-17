package laws

import (
	"math/rand"
	"strings"
	"unicode/utf8"

	"github.com/aosanya/mwanachama-backend-compute/limit"
)

var spaceVariants = []string{" ", "  ", "\t", "\n", " \n  ", "\t \t", "\r\n"}

var twoCharacterOperators = []string{"<=", ">=", "==", "!=", "&&", "||"}

type token struct {
	text string
	gap  string
}

func isWordCharacter(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func tokenize(text string) []token {
	var tokens []token
	i := 0
	for i < len(text) {
		gapStart := i
		for i < len(text) && isSpace(text[i]) {
			i++
		}
		gap := text[gapStart:i]
		if i >= len(text) {
			tokens = append(tokens, token{text: "", gap: gap})
			break
		}
		start := i
		switch {
		case text[i] == '"':
			i++
			for i < len(text) && text[i] != '"' {
				if text[i] == '\\' {
					i++
				}
				i++
			}
			i++
		case isWordCharacter(text[i]):
			for i < len(text) && isWordCharacter(text[i]) {
				i++
				if i+1 < len(text) && (text[i] == '.' || text[i] == '@') && isWordCharacter(text[i+1]) {
					i++
				}
			}
		default:
			i++
			for _, operator := range twoCharacterOperators {
				if strings.HasPrefix(text[start:], operator) {
					i = start + 2
				}
			}
		}
		tokens = append(tokens, token{text: text[start:i], gap: gap})
	}
	return tokens
}

func SpellDifferently(standard string, seed int64) string {
	rng := rand.New(rand.NewSource(seed))
	room := limit.Proposed().FormulaLength - utf8.RuneCountInString(standard)
	var out strings.Builder
	space := func(replaced string) string {
		variant := spaceVariants[rng.Intn(len(spaceVariants))]
		growth := len(variant) - len(replaced)
		if growth > room {
			return replaced
		}
		room -= growth
		return variant
	}
	for index, t := range tokenize(standard) {
		switch {
		case t.gap != "":
			out.WriteString(space(t.gap))
		case index > 0 && rng.Intn(10) < 3:
			out.WriteString(space(""))
		case index == 0 && rng.Intn(2) == 0:
			out.WriteString(space(""))
		}
		out.WriteString(t.text)
	}
	return out.String()
}
