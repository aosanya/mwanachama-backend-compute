package problem

import (
	"bytes"
	"encoding/json"
)

const (
	RuleRequired      = "required"
	RuleType          = "type"
	RuleMinItems      = "min-items"
	RuleMaxItems      = "max-items"
	RuleNoNegatives   = "no-negatives"
	RuleRange         = "range"
	RuleDecimalPlaces = "decimal-places"
	RuleAllowedValues = "allowed-values"
	RuleCheck         = "check"
	RuleKnownInput    = "known-input"
)

const (
	CodeRequired      = "required"
	CodeWrongType     = "wrong-type"
	CodeTooSmall      = "too-small"
	CodeTooLarge      = "too-large"
	CodeNoNegatives   = "no-negatives"
	CodeDecimalPlaces = "decimal-places"
	CodeNotAllowed    = "not-allowed"
	CodeListTooShort  = "list-too-short"
	CodeListTooLong   = "list-too-long"
	CodeCheckFailed   = "check-failed"
	CodeCheckError    = "check-error"
	CodeUnknownInput  = "unknown-input"
)

const (
	minCodeLength = 1
	maxCodeLength = 40
)

type Problem struct {
	Input    *string `json:"input"`
	Item     *int    `json:"item"`
	Given    *bool   `json:"given"`
	Value    *string `json:"value"`
	Rule     string  `json:"rule"`
	Sentence string  `json:"sentence"`
	Code     string  `json:"code"`
	Check    *string `json:"check"`
}

func AllCodes() []string {
	return []string{
		CodeRequired, CodeWrongType, CodeTooSmall, CodeTooLarge, CodeNoNegatives, CodeDecimalPlaces,
		CodeNotAllowed, CodeListTooShort, CodeListTooLong, CodeCheckFailed, CodeCheckError, CodeUnknownInput,
	}
}

func ValidCode(code string) bool {
	if len(code) < minCodeLength || len(code) > maxCodeLength {
		return false
	}
	for _, character := range code {
		letter := character >= 'a' && character <= 'z'
		digit := character >= '0' && character <= '9'
		if !letter && !digit && character != '-' {
			return false
		}
	}
	return true
}

func (p Problem) JSON() ([]byte, error) {
	return encode(p)
}

func ListJSON(problems []Problem) ([]byte, error) {
	if problems == nil {
		problems = []Problem{}
	}
	return encode(problems)
}

func encode(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

func text(value string) *string {
	return &value
}

func number(value int) *int {
	return &value
}

func flag(value bool) *bool {
	return &value
}
