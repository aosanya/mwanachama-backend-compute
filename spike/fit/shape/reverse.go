package shape

import (
	"regexp"
	"strconv"
	"strings"
)

func Reverse(text, decimalWrapper string) string {
	pattern := regexp.MustCompile(
		`"(?:[^"\\]|\\.)*"` +
			`|` + regexp.QuoteMeta(decimalWrapper) + `\("(\d+\.\d+)"\)` +
			`|\b\d+\.\d+\b` +
			`|\b[A-Za-z_][A-Za-z0-9_]*_v\d+\b`)
	return pattern.ReplaceAllStringFunc(text, func(match string) string {
		switch {
		case strings.HasPrefix(match, `"`):
			return match
		case strings.HasPrefix(match, decimalWrapper+"("):
			return strings.TrimSuffix(strings.TrimPrefix(match, decimalWrapper+`("`), `")`)
		case match[0] >= '0' && match[0] <= '9':
			return FloatLiteral(match)
		}
		at := strings.LastIndex(match, "_v")
		return match[:at] + "@" + match[at+2:]
	})
}

func FloatLiteral(text string) string {
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return text + "f"
	}
	return strconv.FormatFloat(value, 'f', -1, 64) + "f"
}
