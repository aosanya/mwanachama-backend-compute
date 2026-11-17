package cel

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types/ref"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

var kindTypes = map[harness.Kind]*cel.Type{
	harness.KindDecimal:     decimalType,
	harness.KindFloat:       cel.DoubleType,
	harness.KindInteger:     cel.IntType,
	harness.KindText:        cel.StringType,
	harness.KindBoolean:     cel.BoolType,
	harness.KindDate:        cel.TimestampType,
	harness.KindDecimalList: cel.ListType(decimalType),
	harness.KindIntegerList: cel.ListType(cel.IntType),
}

func typeOf(kind harness.Kind) *cel.Type {
	return kindTypes[kind]
}

func variables(c harness.Case) []cel.EnvOption {
	options := make([]cel.EnvOption, len(c.Params))
	for i, p := range c.Params {
		options[i] = cel.Variable(p.Name, typeOf(p.Kind))
	}
	return options
}

func nativeOf(v ref.Val, kind harness.Kind) any {
	switch kind {
	case harness.KindDecimalList:
		items, _ := decimalsOf(v)
		return items
	case harness.KindIntegerList:
		items, _ := intsOf(v)
		return items
	}
	return v.Value()
}
