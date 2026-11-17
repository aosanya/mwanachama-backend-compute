package own

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

const oneCharSymbols = "+-*/%<>!?:,()[]"

var twoCharSymbols = []string{"<=", ">=", "==", "!=", "&&", "||"}

type lexer struct {
	runes []rune
	at    int
	pos   Pos
}

func tokenize(source string) ([]token, error) {
	l := &lexer{runes: []rune(source), pos: Pos{Line: 1, Column: 1}}
	var tokens []token
	for {
		l.skipSpace()
		next, err := l.next()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, next)
		if next.kind == tokenEnd {
			return tokens, nil
		}
	}
}

func (l *lexer) next() (token, error) {
	start := l.pos
	if l.done() {
		return token{kind: tokenEnd, pos: start}, nil
	}
	r := l.peek(0)
	switch {
	case isDigit(r):
		return l.number(start)
	case r == '"':
		return l.text(start)
	case isNameStart(r):
		return l.name(start)
	}
	return l.symbol(start)
}

func (l *lexer) done() bool { return l.at >= len(l.runes) }

func (l *lexer) peek(offset int) rune {
	i := l.at + offset
	if i >= len(l.runes) {
		return 0
	}
	return l.runes[i]
}

func (l *lexer) advance() rune {
	r := l.runes[l.at]
	l.at++
	if r == '\n' {
		l.pos.Line++
		l.pos.Column = 1
	} else {
		l.pos.Column++
	}
	return r
}

func (l *lexer) skipSpace() {
	for !l.done() && unicode.IsSpace(l.peek(0)) {
		l.advance()
	}
}

func (l *lexer) take(accept func(rune) bool) string {
	start := l.at
	for !l.done() && accept(l.peek(0)) {
		l.advance()
	}
	return string(l.runes[start:l.at])
}

func (l *lexer) takeCount(count int) string {
	start := l.at
	for range count {
		l.advance()
	}
	return string(l.runes[start:l.at])
}

func (l *lexer) number(start Pos) (token, error) {
	whole := l.take(isDigit)
	if len(whole) == 4 && l.atDateTail() {
		return token{tokenDate, whole + l.takeCount(6), start}, nil
	}
	kind, text := tokenInteger, whole
	if l.peek(0) == '.' && isDigit(l.peek(1)) {
		l.advance()
		kind, text = tokenDecimal, whole+"."+l.take(isDigit)
	}
	if l.peek(0) == 'f' && !isNameRune(l.peek(1)) {
		l.advance()
		kind = tokenFloat
	}
	if isNameRune(l.peek(0)) {
		return token{}, readError(l.pos, "a number cannot be followed by letters")
	}
	return token{kind, text, start}, nil
}

func (l *lexer) atDateTail() bool {
	shape := "-dd-dd"
	for i, want := range shape {
		got := l.peek(i)
		if (want == 'd' && !isDigit(got)) || (want == '-' && got != '-') {
			return false
		}
	}
	return !isDigit(l.peek(len(shape)))
}

func (l *lexer) text(start Pos) (token, error) {
	l.advance()
	var value strings.Builder
	for !l.done() {
		switch r := l.advance(); r {
		case '"':
			return token{tokenText, value.String(), start}, nil
		case '\\':
			escaped, err := l.escape()
			if err != nil {
				return token{}, err
			}
			value.WriteRune(escaped)
		default:
			value.WriteRune(r)
		}
	}
	return token{}, readError(start, "the text is not closed")
}

func (l *lexer) escape() (rune, error) {
	at := l.pos
	if l.done() {
		return 0, readError(at, "the text ends inside an escape")
	}
	switch r := l.advance(); r {
	case '"', '\\':
		return r, nil
	case 'n':
		return '\n', nil
	case 't':
		return '\t', nil
	default:
		return 0, readError(at, fmt.Sprintf("the escape \\%c is not known", r))
	}
}

func (l *lexer) name(start Pos) (token, error) {
	text := l.take(isNameRune)
	if l.peek(0) != '@' {
		return token{tokenName, text, start}, nil
	}
	if !isDigit(l.peek(1)) {
		return token{}, readError(l.pos, "a version number must follow @")
	}
	l.advance()
	return token{tokenName, text + "@" + l.take(isDigit), start}, nil
}

func (l *lexer) symbol(start Pos) (token, error) {
	pair := string([]rune{l.peek(0), l.peek(1)})
	if slices.Contains(twoCharSymbols, pair) {
		l.takeCount(2)
		return token{tokenSymbol, pair, start}, nil
	}
	r := l.peek(0)
	if strings.ContainsRune(oneCharSymbols, r) {
		l.advance()
		return token{tokenSymbol, string(r), start}, nil
	}
	return token{}, readError(start, fmt.Sprintf("the character %q is not allowed", r))
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func isNameStart(r rune) bool { return r == '_' || unicode.IsLetter(r) }

func isNameRune(r rune) bool { return isNameStart(r) || isDigit(r) }
