package cel

import (
	"strings"
	"text/scanner"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const decimalFunction = "decimal"

func translate(source string) string {
	var s scanner.Scanner
	s.Init(strings.NewReader(source))
	s.Mode = scanner.ScanIdents | scanner.ScanFloats | scanner.ScanStrings
	s.Whitespace = 0
	var out strings.Builder
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		out.WriteString(translateToken(&s, tok))
	}
	return out.String()
}

func translateToken(s *scanner.Scanner, tok rune) string {
	text := s.TokenText()
	switch {
	case (tok == scanner.Int || tok == scanner.Float) && s.Peek() == 'f':
		s.Next()
		return asDouble(text)
	case tok == scanner.Float:
		return decimalFunction + `("` + text + `")`
	case tok == '@':
		return "_v"
	}
	return text
}

func asDouble(text string) string {
	if strings.ContainsAny(text, ".eE") {
		return text
	}
	return text + ".0"
}

func Translations(c harness.Case) (string, bool) {
	text := translate(c.Source)
	return text, text != c.Source
}
