package harness

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

const floatTolerance = 1e-9

func Equal(want, got any) bool {
	switch w := want.(type) {
	case numeric.Decimal:
		g, ok := got.(numeric.Decimal)
		return ok && w.String() == g.String()
	case float64:
		g, ok := got.(float64)
		return ok && (w == g || math.Abs(w-g) <= floatTolerance)
	case time.Time:
		g, ok := got.(time.Time)
		return ok && w.Equal(g)
	case []numeric.Decimal:
		g, ok := got.([]numeric.Decimal)
		return ok && equalDecimalLists(w, g)
	}
	return reflect.DeepEqual(want, got)
}

func Describe(v any) string {
	kind, known := KindOf(v)
	if !known {
		return fmt.Sprintf("%v (%T)", v, v)
	}
	return fmt.Sprintf("%s (%s)", render(v), kind)
}

func InputKey(inputs map[string]any) string {
	names := make([]string, 0, len(inputs))
	for name := range inputs {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = name + "=" + Describe(inputs[name])
	}
	return strings.Join(parts, ";")
}

func equalDecimalLists(a, b []numeric.Decimal) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}
	return true
}

func render(v any) string {
	switch x := v.(type) {
	case numeric.Decimal:
		return x.String()
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	case string:
		return strconv.Quote(x)
	case time.Time:
		return x.UTC().Format("2006-01-02")
	case []numeric.Decimal:
		texts := make([]string, len(x))
		for i, d := range x {
			texts[i] = d.String()
		}
		return "[" + strings.Join(texts, ", ") + "]"
	}
	return fmt.Sprint(v)
}
