package cel

import (
	"fmt"
	"reflect"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

var decimalType = cel.OpaqueType("decimal")

var nativeDecimal = reflect.TypeOf(numeric.Decimal{})

type decimalValue struct {
	d numeric.Decimal
}

func (v decimalValue) ConvertToNative(typeDesc reflect.Type) (any, error) {
	if typeDesc == nativeDecimal {
		return v.d, nil
	}
	return nil, fmt.Errorf("decimal cannot become %s", typeDesc)
}

func (v decimalValue) ConvertToType(typeValue ref.Type) ref.Val {
	if typeValue == types.TypeType {
		return decimalType
	}
	if typeValue.TypeName() == decimalType.TypeName() {
		return v
	}
	return types.NewErr("decimal cannot become %s", typeValue.TypeName())
}

func (v decimalValue) Equal(other ref.Val) ref.Val {
	o, ok := asDecimal(other)
	return types.Bool(ok && v.d.Cmp(o) == 0)
}

func (v decimalValue) Type() ref.Type {
	return decimalType
}

func (v decimalValue) Value() any {
	return v.d
}

func asDecimal(v ref.Val) (numeric.Decimal, bool) {
	switch x := v.(type) {
	case decimalValue:
		return x.d, true
	case types.Int:
		return numeric.FromInt(int64(x)), true
	}
	return numeric.Decimal{}, false
}
