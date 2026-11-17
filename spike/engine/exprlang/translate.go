package exprlang

import (
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func Translations(c harness.Case) (string, bool) {
	text := translate(c.Source)
	return text, text != c.Source
}

func translate(source string) string {
	var out strings.Builder
	for i := 0; i < len(source); {
		end := i + 1
		switch ch := source[i]; {
		case ch == '"' || ch == '\'' || ch == '`':
			end = endOfString(source, i)
			out.WriteString(source[i:end])
		case isIdentifierStart(ch):
			end = endOfRun(source, i, isIdentifierPart)
			out.WriteString(source[i:end])
		case isDigit(ch):
			end = endOfNumber(source, i)
			out.WriteString(numberLiteral(source[i:end]))
		case ch == '@':
			out.WriteString("_v")
		default:
			out.WriteByte(ch)
		}
		i = end
	}
	return out.String()
}

func numberLiteral(text string) string {
	if body, isFloat := strings.CutSuffix(text, "f"); isFloat {
		return body
	}
	if strings.Contains(text, ".") {
		return `dec("` + text + `")`
	}
	return text
}

func endOfString(source string, start int) int {
	quote := source[start]
	for i := start + 1; i < len(source); i++ {
		switch {
		case source[i] == '\\' && quote != '`':
			i++
		case source[i] == quote:
			return i + 1
		}
	}
	return len(source)
}

func endOfNumber(source string, start int) int {
	end := endOfRun(source, start, isDigit)
	if end+1 < len(source) && source[end] == '.' && isDigit(source[end+1]) {
		end = endOfRun(source, end+1, isDigit)
	}
	if end < len(source) && source[end] == 'f' {
		end++
	}
	return end
}

func endOfRun(source string, start int, member func(byte) bool) int {
	end := start
	for end < len(source) && member(source[end]) {
		end++
	}
	return end
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentifierStart(ch byte) bool {
	return ch == '_' || ch == '$' || ch >= 0x80 || (ch|0x20 >= 'a' && ch|0x20 <= 'z')
}

func isIdentifierPart(ch byte) bool {
	return isIdentifierStart(ch) || isDigit(ch)
}
