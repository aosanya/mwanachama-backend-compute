package numtext

import "strings"

func Read(kind Type, received Received) (Value, []Refusal) {
	switch kind {
	case IntegerList:
		return readList(kind, Integer, received)
	case DecimalList:
		return readList(kind, Decimal, received)
	case Integer, Decimal, Float, Text, Boolean, Date:
		if received.List {
			return Value{}, []Refusal{newRefusal(kind, ReasonListForSingle, 0, "["+strings.Join(received.Items, ", ")+"]")}
		}
		value, reason := readOne(kind, received.Text)
		if reason != "" {
			return Value{}, []Refusal{newRefusal(kind, reason, 0, received.Text)}
		}
		return value, nil
	}
	return Value{}, []Refusal{newRefusal(kind, ReasonUnknownType, 0, received.Text)}
}

func readList(kind, element Type, received Received) (Value, []Refusal) {
	if !received.List {
		return Value{}, []Refusal{newRefusal(kind, ReasonTextForList, 0, received.Text)}
	}
	items := make([]Value, 0, len(received.Items))
	var refusals []Refusal
	for i, text := range received.Items {
		value, reason := readOne(element, text)
		if reason != "" {
			refusals = append(refusals, newRefusal(element, reason, i+1, text))
			continue
		}
		items = append(items, value)
	}
	if len(refusals) > 0 {
		return Value{}, refusals
	}
	return Value{Type: kind, Items: items}, nil
}

func readOne(kind Type, text string) (Value, Reason) {
	switch kind {
	case Integer:
		return readInteger(text)
	case Decimal:
		return readDecimal(text)
	case Float:
		return readFloat(text)
	case Text:
		return readText(text)
	case Boolean:
		return readBoolean(text)
	}
	return readDate(text)
}

func newRefusal(kind Type, reason Reason, item int, received string) Refusal {
	return Refusal{Rule: RuleType, Type: kind, Reason: reason, Item: item, Received: received}
}
