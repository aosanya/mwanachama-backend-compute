package fit

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

var twoCharSymbols = []string{"<=", ">=", "==", "!=", "&&", "||"}

func isWordByte(b byte) bool {
	return b == '_' || b == '@' || b == '.' || (b >= '0' && b <= '9') || (b|0x20 >= 'a' && b|0x20 <= 'z')
}

func tokens(source string) []string {
	var found []string
	for i := 0; i < len(source); {
		end := i + 1
		switch ch := source[i]; {
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			i++
			continue
		case ch == '"':
			end = endOfText(source, i)
		case isWordByte(ch):
			for end < len(source) && isWordByte(source[end]) {
				end++
			}
		case i+1 < len(source) && contains(twoCharSymbols, source[i:i+2]):
			end = i + 2
		}
		found = append(found, source[i:end])
		i = end
	}
	return found
}

func endOfText(source string, start int) int {
	for i := start + 1; i < len(source); i++ {
		switch source[i] {
		case '\\':
			i++
		case '"':
			return i + 1
		}
	}
	return len(source)
}

func contains(list []string, item string) bool {
	for _, entry := range list {
		if entry == item {
			return true
		}
	}
	return false
}

func endsWord(token string) bool   { return isWordByte(token[len(token)-1]) }
func startsWord(token string) bool { return isWordByte(token[0]) }

var gaps = []string{"", " ", "  ", "\t", "\n", " \t ", "\r\n", "   "}

func respace(source string, style int) string {
	found := tokens(source)
	random := rand.New(rand.NewSource(int64(style)))
	var out strings.Builder
	out.WriteString(pick(style, random, 0))
	for i, token := range found {
		if i > 0 {
			gap := pick(style, random, i)
			if gap == "" && endsWord(found[i-1]) && startsWord(token) {
				gap = " "
			}
			out.WriteString(gap)
		}
		out.WriteString(token)
	}
	out.WriteString(pick(style, random, len(found)))
	return out.String()
}

func pick(style int, random *rand.Rand, at int) string {
	switch style {
	case 0:
		return ""
	case 1:
		return "  "
	case 2:
		return []string{"\t", "\n"}[at%2]
	}
	return gaps[random.Intn(len(gaps))]
}

const variantStyles = 6

func variants(source string) []string {
	made := make([]string, variantStyles)
	for style := range made {
		made[style] = respace(source, style)
	}
	return made
}

func TestRespaceKeepsTokens(t *testing.T) {
	for _, c := range harness.Cases() {
		want := tokens(c.Source)
		for style, variant := range variants(c.Source) {
			got := tokens(variant)
			if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
				t.Errorf("formula %d style %d lost tokens: %q", c.Number, style, variant)
			}
			if variant == c.Source && style > 0 {
				t.Errorf("formula %d style %d did not change the spacing", c.Number, style)
			}
		}
	}
}
