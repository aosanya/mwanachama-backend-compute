package numtext

import (
	"strconv"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

func allDigits(text string) bool {
	if text == "" {
		return false
	}
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

func readInteger(text string) (Value, Reason) {
	if text == "" {
		return Value{}, ReasonEmpty
	}
	if text[0] == '+' {
		return Value{}, ReasonSign
	}
	body := strings.TrimPrefix(text, "-")
	if !allDigits(body) {
		return Value{}, ReasonNotDigits
	}
	if text == "-0" {
		return Value{}, ReasonNegativeZero
	}
	if len(body) > 1 && body[0] == '0' {
		return Value{}, ReasonLeadingZero
	}
	if len(body) > maxIntegerDigits {
		return Value{}, ReasonTooManyDigits
	}
	whole, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return Value{}, ReasonTooManyDigits
	}
	return Value{Type: Integer, Text: text, Integer: whole, Decimal: number.DecimalFromInt(whole)}, ""
}

func checkDecimalText(text string) Reason {
	if text == "" {
		return ReasonEmpty
	}
	if text[0] == '+' {
		return ReasonSign
	}
	body := strings.TrimPrefix(text, "-")
	whole, fraction, hasPoint := strings.Cut(body, ".")
	if whole == "" {
		if hasPoint {
			return ReasonMissingWholePart
		}
		return ReasonNotDigits
	}
	if !allDigits(whole) {
		return ReasonNotDigits
	}
	if hasPoint {
		if fraction == "" {
			return ReasonMissingFraction
		}
		if !allDigits(fraction) {
			return ReasonNotDigits
		}
	}
	if len(whole) > 1 && whole[0] == '0' {
		return ReasonLeadingZero
	}
	if len(whole)+len(fraction) > maxDecimalDigits {
		return ReasonTooManyDigits
	}
	return ""
}

func readDecimal(text string) (Value, Reason) {
	if reason := checkDecimalText(text); reason != "" {
		return Value{}, reason
	}
	decimal, err := number.ParseDecimal(text)
	if err != nil {
		return Value{}, ReasonNotDigits
	}
	return Value{Type: Decimal, Text: text, Decimal: decimal}, ""
}

func readFloat(text string) (Value, Reason) {
	if reason := checkDecimalText(text); reason != "" {
		return Value{}, reason
	}
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return Value{}, ReasonTooManyDigits
	}
	approximate, err := number.NewFloat(parsed)
	if err != nil {
		return Value{}, ReasonTooManyDigits
	}
	return Value{Type: Float, Text: text, Float: approximate}, ""
}
