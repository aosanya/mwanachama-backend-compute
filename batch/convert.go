package batch

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

var numtextTypes = map[formula.Type]numtext.Type{
	formula.TypeInteger:     numtext.Integer,
	formula.TypeDecimal:     numtext.Decimal,
	formula.TypeFloat:       numtext.Float,
	formula.TypeText:        numtext.Text,
	formula.TypeBoolean:     numtext.Boolean,
	formula.TypeDate:        numtext.Date,
	formula.TypeListInteger: numtext.IntegerList,
	formula.TypeListDecimal: numtext.DecimalList,
}

var formulaTypes = map[numtext.Type]formula.Type{
	numtext.Integer:     formula.TypeInteger,
	numtext.Decimal:     formula.TypeDecimal,
	numtext.Float:       formula.TypeFloat,
	numtext.Text:        formula.TypeText,
	numtext.Boolean:     formula.TypeBoolean,
	numtext.Date:        formula.TypeDate,
	numtext.IntegerList: formula.TypeListInteger,
	numtext.DecimalList: formula.TypeListDecimal,
}

func formulaValue(value numtext.Value) (formula.Value, bool) {
	switch value.Type {
	case numtext.Integer:
		return formula.IntegerValue(value.Integer), true
	case numtext.Decimal:
		return formula.DecimalValue(value.Decimal), true
	case numtext.Float:
		return formula.FloatValue(value.Float), true
	case numtext.Text:
		return formula.TextValue(value.Text), true
	case numtext.Boolean:
		return formula.BooleanValue(value.Boolean), true
	case numtext.Date:
		return formula.DateValue(value.Date), true
	case numtext.IntegerList:
		items := make([]int64, len(value.Items))
		for i, item := range value.Items {
			items[i] = item.Integer
		}
		return formula.IntegerListValue(items), true
	case numtext.DecimalList:
		items := make([]number.Decimal, len(value.Items))
		for i, item := range value.Items {
			items[i] = item.Decimal
		}
		return formula.DecimalListValue(items), true
	}
	return formula.Value{}, false
}

func emptyList(kind formula.Type) (formula.Value, bool) {
	switch kind {
	case formula.TypeListInteger:
		return formula.IntegerListValue(nil), true
	case formula.TypeListDecimal:
		return formula.DecimalListValue(nil), true
	}
	return formula.Value{}, false
}
