package formula

import (
	"fmt"
	"unicode"

	"github.com/aosanya/mwanachama-backend-compute/limit"
)

type tokenKind int

const (
	tokenEnd tokenKind = iota
	tokenInteger
	tokenDecimal
	tokenText
	tokenName
	tokenSymbol
)

type token struct {
	kind   tokenKind
	text   string
	start  int
	finish int
}

type syntaxFailure struct {
	offset  int
	code    string
	message string
}

type lexer struct {
	runes    []rune
	position int
	settings limit.Settings
}

func newLexer(text string, settings limit.Settings) *lexer {
	return &lexer{runes: []rune(text), settings: settings}
}

func (l *lexer) end() int {
	return len(l.runes)
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isNameStart(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isNamePart(r rune) bool {
	return isNameStart(r) || isDigit(r) || r == '_'
}

func describeCharacter(r rune) string {
	switch {
	case r >= 0x80:
		return fmt.Sprintf("U+%04X", r)
	default:
		return string(r)
	}
}

func unexpected(r rune, offset int) *syntaxFailure {
	if unicode.IsControl(r) {
		return &syntaxFailure{offset: offset, code: codeBadSyntax, message: "unexpected control character"}
	}
	return &syntaxFailure{offset: offset, code: codeBadSyntax, message: "unexpected character " + describeCharacter(r)}
}

func refusal(offset int, name limit.Name, value int) *syntaxFailure {
	return &syntaxFailure{offset: offset, code: limitCode(name), message: limit.Refusal{Limit: name, Value: value}.Sentence()}
}

func (l *lexer) skipSpace() {
	for l.position < len(l.runes) {
		switch l.runes[l.position] {
		case ' ', '\t', '\n', '\r':
			l.position++
		default:
			return
		}
	}
}

func (l *lexer) next() (token, *syntaxFailure) {
	l.skipSpace()
	if l.position >= len(l.runes) {
		return token{kind: tokenEnd, start: len(l.runes), finish: len(l.runes)}, nil
	}
	start := l.position
	r := l.runes[start]
	switch {
	case isDigit(r):
		return l.number()
	case isNameStart(r):
		return l.name()
	case r == '"':
		return l.text()
	}
	return l.symbol()
}

func (l *lexer) number() (token, *syntaxFailure) {
	start := l.position
	for l.position < len(l.runes) && isDigit(l.runes[l.position]) {
		l.position++
	}
	whole := l.position - start
	if whole > 1 && l.runes[start] == '0' {
		return token{}, &syntaxFailure{offset: start, code: codeBadSyntax, message: "a number cannot have a leading zero"}
	}
	hasFraction := l.position+1 < len(l.runes) && l.runes[l.position] == '.' && isDigit(l.runes[l.position+1])
	if !hasFraction {
		if l.position < len(l.runes) && l.runes[l.position] == '.' {
			return token{}, unexpected('.', l.position)
		}
		if whole > l.settings.IntegerDigits {
			return token{}, refusal(start, limit.NumberMagnitudeLimit, l.settings.IntegerDigits)
		}
		return token{kind: tokenInteger, text: string(l.runes[start:l.position]), start: start, finish: l.position}, nil
	}
	l.position++
	fractionStart := l.position
	for l.position < len(l.runes) && isDigit(l.runes[l.position]) {
		l.position++
	}
	places := l.position - fractionStart
	if whole+places > l.settings.DecimalDigits {
		return token{}, refusal(start, limit.NumberMagnitudeLimit, l.settings.DecimalDigits)
	}
	if places > l.settings.DecimalPlaces {
		return token{}, refusal(start, limit.DecimalPlacesLimit, l.settings.DecimalPlaces)
	}
	return token{kind: tokenDecimal, text: string(l.runes[start:l.position]), start: start, finish: l.position}, nil
}

func (l *lexer) name() (token, *syntaxFailure) {
	start := l.position
	for l.position < len(l.runes) && isNamePart(l.runes[l.position]) {
		l.position++
		if l.position-start > l.settings.NameLength {
			return token{}, unexpected(l.runes[l.position-1], l.position-1)
		}
	}
	return token{kind: tokenName, text: string(l.runes[start:l.position]), start: start, finish: l.position}, nil
}

func (l *lexer) text() (token, *syntaxFailure) {
	start := l.position
	l.position++
	var value []rune
	for {
		if l.position >= len(l.runes) {
			return token{}, &syntaxFailure{offset: start, code: codeBadSyntax, message: "text is not closed"}
		}
		r := l.runes[l.position]
		switch {
		case r == '"':
			l.position++
			return token{kind: tokenText, text: string(value), start: start, finish: l.position}, nil
		case r == '\\':
			if l.position+1 >= len(l.runes) {
				return token{}, &syntaxFailure{offset: l.position, code: codeBadSyntax, message: "text ends inside an escape"}
			}
			escaped := l.runes[l.position+1]
			if escaped != '"' && escaped != '\\' {
				return token{}, &syntaxFailure{offset: l.position, code: codeBadSyntax, message: "text has an unknown escape"}
			}
			value = append(value, escaped)
			l.position += 2
		case unicode.IsControl(r):
			return token{}, unexpected(r, l.position)
		default:
			value = append(value, r)
			l.position++
		}
		if len(value) > l.settings.TextLiteral {
			return token{}, refusal(start, limit.TextLengthLimit, l.settings.TextLiteral)
		}
	}
}

var twoCharacterSymbols = map[string]bool{"<=": true, ">=": true, "==": true, "!=": true, "&&": true, "||": true}

const oneCharacterSymbols = "+-*/%<>!?:()[],.@"

func (l *lexer) symbol() (token, *syntaxFailure) {
	start := l.position
	r := l.runes[start]
	if start+1 < len(l.runes) {
		pair := string(l.runes[start : start+2])
		if twoCharacterSymbols[pair] {
			l.position += 2
			return token{kind: tokenSymbol, text: pair, start: start, finish: l.position}, nil
		}
	}
	for _, s := range oneCharacterSymbols {
		if r == s {
			l.position++
			return token{kind: tokenSymbol, text: string(r), start: start, finish: l.position}, nil
		}
	}
	return token{}, unexpected(r, start)
}
