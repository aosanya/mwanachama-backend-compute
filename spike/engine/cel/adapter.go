package cel

import (
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type adapter struct{}

func (a adapter) NativeToValue(value any) ref.Val {
	switch v := value.(type) {
	case numeric.Decimal:
		return decimalValue{v}
	case []numeric.Decimal:
		return types.NewDynamicList(a, v)
	}
	return types.DefaultTypeAdapter.NativeToValue(value)
}
