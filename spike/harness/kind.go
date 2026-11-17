package harness

import (
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type Kind int

const (
	KindDecimal Kind = iota + 1
	KindFloat
	KindInteger
	KindText
	KindBoolean
	KindDate
	KindDecimalList
	KindIntegerList
)

var kindNames = map[Kind]string{
	KindDecimal:     "decimal",
	KindFloat:       "float",
	KindInteger:     "integer",
	KindText:        "text",
	KindBoolean:     "boolean",
	KindDate:        "date",
	KindDecimalList: "decimal list",
	KindIntegerList: "integer list",
}

func (k Kind) String() string {
	if name, ok := kindNames[k]; ok {
		return name
	}
	return "unknown"
}

type Param struct {
	Name string
	Kind Kind
}

func KindOf(v any) (Kind, bool) {
	switch v.(type) {
	case numeric.Decimal:
		return KindDecimal, true
	case float64:
		return KindFloat, true
	case int64:
		return KindInteger, true
	case string:
		return KindText, true
	case bool:
		return KindBoolean, true
	case time.Time:
		return KindDate, true
	case []numeric.Decimal:
		return KindDecimalList, true
	case []int64:
		return KindIntegerList, true
	}
	return 0, false
}
