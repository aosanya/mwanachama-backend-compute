package exprlang

import (
	"reflect"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

var (
	decimalType  = reflect.TypeFor[numeric.Decimal]()
	decimalsType = reflect.TypeFor[[]numeric.Decimal]()
	int64sType   = reflect.TypeFor[[]int64]()
	anysType     = reflect.TypeFor[[]any]()
	anyType      = reflect.TypeFor[any]()
	intType      = reflect.TypeFor[int]()
	int64Type    = reflect.TypeFor[int64]()
	floatType    = reflect.TypeFor[float64]()
	boolType     = reflect.TypeFor[bool]()
	stringType   = reflect.TypeFor[string]()
	timeType     = reflect.TypeFor[time.Time]()

	integerTypes = []reflect.Type{intType, int64Type}

	kindTypes = map[harness.Kind]reflect.Type{
		harness.KindDecimal:     decimalType,
		harness.KindFloat:       floatType,
		harness.KindInteger:     int64Type,
		harness.KindText:        stringType,
		harness.KindBoolean:     boolType,
		harness.KindDate:        timeType,
		harness.KindDecimalList: decimalsType,
		harness.KindIntegerList: int64sType,
	}
)

func signature(out reflect.Type, in ...reflect.Type) any {
	return reflect.Zero(reflect.FuncOf(in, []reflect.Type{out}, false)).Interface()
}

func asInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int:
		return int64(n), true
	}
	return 0, false
}

func asDecimal(v any) (numeric.Decimal, bool) {
	if d, ok := v.(numeric.Decimal); ok {
		return d, true
	}
	n, ok := asInt64(v)
	return numeric.FromInt(n), ok
}
