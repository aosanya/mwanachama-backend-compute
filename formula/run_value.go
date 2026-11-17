package formula

import (
	"strconv"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type Value struct {
	kind    Type
	whole   int64
	truth   bool
	real    number.Float
	exact   number.Decimal
	words   string
	day     datefn.Date
	members []Value
}

func IntegerValue(n int64) Value { return Value{kind: TypeInteger, whole: n} }

func DecimalValue(d number.Decimal) Value { return Value{kind: TypeDecimal, exact: d} }

func FloatValue(f number.Float) Value { return Value{kind: TypeFloat, real: f} }

func TextValue(text string) Value { return Value{kind: TypeText, words: text} }

func BooleanValue(truth bool) Value { return Value{kind: TypeBoolean, truth: truth} }

func DateValue(d datefn.Date) Value { return Value{kind: TypeDate, day: d} }

func IntegerListValue(items []int64) Value {
	members := make([]Value, len(items))
	for i, item := range items {
		members[i] = IntegerValue(item)
	}
	return Value{kind: TypeListInteger, members: members}
}

func DecimalListValue(items []number.Decimal) Value {
	members := make([]Value, len(items))
	for i, item := range items {
		members[i] = DecimalValue(item)
	}
	return Value{kind: TypeListDecimal, members: members}
}

func (v Value) Type() Type { return v.kind }

func (v Value) Integer() int64 { return v.whole }

func (v Value) Decimal() number.Decimal { return v.exact }

func (v Value) Float() number.Float { return v.real }

func (v Value) Text() string { return v.words }

func (v Value) Boolean() bool { return v.truth }

func (v Value) Date() datefn.Date { return v.day }

func (v Value) Len() int { return len(v.members) }

func (v Value) Item(index int) Value { return v.members[index] }

func (v Value) String() string {
	switch v.kind {
	case TypeInteger:
		return strconv.FormatInt(v.whole, 10)
	case TypeDecimal:
		return v.exact.String()
	case TypeFloat:
		return strconv.FormatFloat(v.real.Value(), 'f', -1, 64)
	case TypeText:
		return v.words
	case TypeBoolean:
		return strconv.FormatBool(v.truth)
	case TypeDate:
		return v.day.String()
	case TypeListInteger, TypeListDecimal:
		parts := make([]string, len(v.members))
		for i, member := range v.members {
			parts[i] = member.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	return ""
}

func (v Value) widened() Value {
	if v.kind == TypeInteger {
		return DecimalValue(number.DecimalFromInt(v.whole))
	}
	return v
}

func runIsList(t Type) bool {
	return t == TypeListInteger || t == TypeListDecimal
}
