package own

import (
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func (c *compiler) list(n List) (code, error) {
	items := make([]code, len(n.Elements))
	for i, element := range n.Elements {
		item, err := c.expr(element)
		if err != nil {
			return code{}, err
		}
		items[i] = item
	}
	kind, err := listElementKind(n, items)
	if err != nil {
		return code{}, err
	}
	for i := range items {
		items[i] = promote(items[i], kind)
	}
	return listOf(n, kind, items)
}

func listElementKind(n List, items []code) (harness.Kind, error) {
	kind := items[0].kind
	for _, item := range items[1:] {
		common, err := commonKind(kind, item.kind)
		if err != nil {
			return 0, typeError(n.Pos, "the values in a list must agree, but "+err.Error())
		}
		kind = common
	}
	return kind, nil
}

func listOf(n List, kind harness.Kind, items []code) (code, error) {
	switch kind {
	case harness.KindDecimal:
		return code{harness.KindDecimalList, collect[numeric.Decimal](items)}, nil
	case harness.KindInteger:
		return code{harness.KindIntegerList, collect[int64](items)}, nil
	}
	return code{}, typeError(n.Pos, "a list can hold only integers or decimals, not "+kind.String()+" values")
}
