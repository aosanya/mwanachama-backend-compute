package numtext

import (
	"fmt"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

const RuleType = "type"

const (
	maxIntegerDigits = 18
	maxDecimalDigits = 30
	maxTextLength    = 200
)

type Type int

const (
	Integer Type = iota + 1
	Decimal
	Float
	Text
	Boolean
	Date
	IntegerList
	DecimalList
)

func (t Type) String() string {
	switch t {
	case Integer:
		return "integer"
	case Decimal:
		return "decimal"
	case Float:
		return "float"
	case Text:
		return "text"
	case Boolean:
		return "boolean"
	case Date:
		return "date"
	case IntegerList:
		return "list of integer"
	case DecimalList:
		return "list of decimal"
	}
	return fmt.Sprintf("type %d", int(t))
}

type Reason string

const (
	ReasonEmpty            Reason = "empty"
	ReasonSign             Reason = "sign"
	ReasonNegativeZero     Reason = "negative-zero"
	ReasonNotDigits        Reason = "not-digits"
	ReasonLeadingZero      Reason = "leading-zero"
	ReasonMissingWholePart Reason = "missing-whole-part"
	ReasonMissingFraction  Reason = "missing-fraction"
	ReasonTooManyDigits    Reason = "too-many-digits"
	ReasonTooLong          Reason = "too-long"
	ReasonControlCharacter Reason = "control-character"
	ReasonInvalidText      Reason = "invalid-text"
	ReasonNotBoolean       Reason = "not-boolean"
	ReasonNotDate          Reason = "not-date"
	ReasonNoSuchDay        Reason = "no-such-day"
	ReasonTextForList      Reason = "text-for-list"
	ReasonListForSingle    Reason = "list-for-single"
	ReasonUnknownType      Reason = "unknown-type"
)

type Received struct {
	Text  string
	Items []string
	List  bool
}

func One(text string) Received {
	return Received{Text: text}
}

func Many(items ...string) Received {
	return Received{Items: items, List: true}
}

type Value struct {
	Type    Type
	Text    string
	Integer int64
	Decimal number.Decimal
	Float   number.Float
	Boolean bool
	Date    datefn.Date
	Items   []Value
}

type Refusal struct {
	Rule     string
	Type     Type
	Reason   Reason
	Item     int
	Received string
}

func (r Refusal) Error() string {
	if r.Item > 0 {
		return fmt.Sprintf("item %d, %q, is refused as %s by the %s rule: %s", r.Item, r.Received, r.Type, r.Rule, r.Reason)
	}
	return fmt.Sprintf("%q is refused as %s by the %s rule: %s", r.Received, r.Type, r.Rule, r.Reason)
}
