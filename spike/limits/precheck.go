package limits

import (
	"strings"
	"unicode/utf8"
)

var placeFunctions = map[string]bool{"round": true, "to_decimal": true}

type bracket struct {
	function string
	commas   int
	tokens   int
	literal  string
}

type scanner struct {
	source   string
	limits   Limits
	at       int
	brackets []bracket
	previous string
}

func Precheck(source string, limits Limits) error {
	if limits.FormulaLength > 0 && len(source) > limits.FormulaLength && utf8.RuneCountInString(source) > limits.FormulaLength {
		return refusal(limits, LimitFormulaLength, LayerPrecheck, nil)
	}
	if strings.TrimSpace(source) == "" {
		return unreadable("the formula is empty", 1, 1, LayerPrecheck, nil)
	}
	s := &scanner{source: source, limits: limits}
	if err := s.scan(); err != nil {
		return err
	}
	return nil
}

func (s *scanner) scan() error {
	for s.at < len(s.source) {
		c := s.source[s.at]
		var err error
		switch {
		case c == '"':
			err = s.text()
		case isDigit(c) && !isNameByte(s.previousByte()):
			err = s.number()
		case isNameStart(c):
			s.name()
		case c == '(' || c == '[':
			err = s.open(c)
		case c == ')' || c == ']':
			err = s.close()
		case c == ',':
			s.comma()
			s.at++
		default:
			if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
				s.token("")
			}
			s.at++
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *scanner) previousByte() byte {
	if s.at == 0 {
		return ' '
	}
	return s.source[s.at-1]
}

func (s *scanner) top() *bracket {
	if len(s.brackets) == 0 {
		return nil
	}
	return &s.brackets[len(s.brackets)-1]
}

func (s *scanner) token(literal string) {
	s.previous = ""
	if b := s.top(); b != nil {
		b.tokens++
		b.literal = literal
	}
}

func (s *scanner) comma() {
	if b := s.top(); b != nil {
		b.commas++
		b.tokens = 0
		b.literal = ""
	}
}

func (s *scanner) name() {
	start := s.at
	for s.at < len(s.source) && isNameByte(s.source[s.at]) {
		s.at++
	}
	name := s.source[start:s.at]
	s.token("")
	s.previous = name
}

func (s *scanner) open(c byte) error {
	function := ""
	if c == '(' {
		function = s.previous
	}
	s.previous = ""
	s.brackets = append(s.brackets, bracket{function: function})
	s.at++
	if s.limits.Depth > 0 && len(s.brackets) >= s.limits.Depth {
		return refusal(s.limits, LimitDepth, LayerPrecheck, nil)
	}
	return nil
}

func (s *scanner) close() error {
	b := s.top()
	if b == nil {
		s.at++
		return nil
	}
	s.brackets = s.brackets[:len(s.brackets)-1]
	s.at++
	s.token("")
	if placeFunctions[b.function] && b.commas == 1 && b.tokens == 1 && b.literal != "" {
		if s.limits.DecimalPlaces > 0 && tooLong(strings.TrimLeft(b.literal, "0"), s.limits.DecimalPlaces) {
			return refusal(s.limits, LimitDecimalPlaces, LayerPrecheck, nil)
		}
	}
	return nil
}

func tooLong(digits string, most int) bool {
	if len(digits) > 9 {
		return true
	}
	value := 0
	for i := 0; i < len(digits); i++ {
		value = value*10 + int(digits[i]-'0')
	}
	return value > most
}

func (s *scanner) text() error {
	s.at++
	length := 0
	for s.at < len(s.source) {
		c := s.source[s.at]
		switch {
		case c == '\\':
			s.at += 2
			length++
		case c == '"':
			s.at++
			s.token("")
			return nil
		default:
			if c&0xC0 != 0x80 {
				length++
			}
			s.at++
		}
		if s.limits.TextLength > 0 && length > s.limits.TextLength {
			return refusal(s.limits, LimitTextLength, LayerPrecheck, nil)
		}
	}
	return nil
}

func (s *scanner) number() error {
	whole := s.digits()
	fraction, exponent := "", ""
	if s.peek(0) == '.' && isDigit(s.peek(1)) {
		s.at++
		fraction = s.digits()
	}
	if (s.peek(0) == 'e' || s.peek(0) == 'E') && (isDigit(s.peek(1)) || ((s.peek(1) == '+' || s.peek(1) == '-') && isDigit(s.peek(2)))) {
		s.at++
		sign := byte('+')
		if s.peek(0) == '+' || s.peek(0) == '-' {
			sign = s.peek(0)
			s.at++
		}
		exponent = s.digits()
		if sign == '-' {
			exponent = ""
		}
	}
	literal := ""
	if fraction == "" && exponent == "" {
		literal = whole
	}
	s.token(literal)
	if s.limits.NumberMagnitude > 0 && len(whole)+exponentValue(exponent) > s.limits.NumberMagnitude && !isDate(whole, s) {
		return refusal(s.limits, LimitNumberMagnitude, LayerPrecheck, nil)
	}
	if s.limits.DecimalPlaces > 0 && len(fraction) > s.limits.DecimalPlaces {
		return refusal(s.limits, LimitDecimalPlaces, LayerPrecheck, nil)
	}
	return nil
}

func isDate(whole string, s *scanner) bool {
	return len(whole) == 4 && s.peek(0) == '-'
}

func exponentValue(digits string) int {
	digits = strings.TrimLeft(digits, "0")
	if len(digits) > 9 {
		return 1 << 30
	}
	value := 0
	for i := 0; i < len(digits); i++ {
		value = value*10 + int(digits[i]-'0')
	}
	return value
}

func (s *scanner) digits() string {
	start := s.at
	for s.at < len(s.source) && isDigit(s.source[s.at]) {
		s.at++
	}
	return s.source[start:s.at]
}

func (s *scanner) peek(offset int) byte {
	if s.at+offset >= len(s.source) {
		return 0
	}
	return s.source[s.at+offset]
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isNameStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c >= 0x80
}

func isNameByte(c byte) bool { return isNameStart(c) || isDigit(c) || c == '@' }
