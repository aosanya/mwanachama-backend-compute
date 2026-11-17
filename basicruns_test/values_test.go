package basicruns

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type namedValue struct {
	name  string
	value formula.Value
}

func splitTopLevel(text string, separator string) []string {
	var parts []string
	depth := 0
	quoted := false
	start := 0
	for i := 0; i < len(text); i++ {
		switch {
		case quoted && text[i] == '\\':
			i++
		case text[i] == '"':
			quoted = !quoted
		case quoted:
		case text[i] == '[':
			depth++
		case text[i] == ']':
			depth--
		case depth == 0 && strings.HasPrefix(text[i:], separator):
			parts = append(parts, text[start:i])
			start = i + len(separator)
			i += len(separator) - 1
		}
	}
	return append(parts, text[start:])
}

func parseInputs(text string) ([]namedValue, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	var list []namedValue
	for _, pair := range splitTopLevel(text, ", ") {
		name, body, found := strings.Cut(strings.TrimSpace(pair), "=")
		if !found {
			return nil, fmt.Errorf("input %q has no =", pair)
		}
		value, err := parseValue(body)
		if err != nil {
			return nil, fmt.Errorf("input %s: %w", name, err)
		}
		list = append(list, namedValue{name: name, value: value})
	}
	return list, nil
}

func parseValue(text string) (formula.Value, error) {
	prefix, body, found := strings.Cut(text, ":")
	if !found {
		return formula.Value{}, fmt.Errorf("value %q has no type prefix", text)
	}
	switch prefix {
	case "int":
		n, err := strconv.ParseInt(body, 10, 64)
		return formula.IntegerValue(n), err
	case "dec":
		d, err := number.ParseDecimal(body)
		return formula.DecimalValue(d), err
	case "float":
		f, err := strconv.ParseFloat(body, 64)
		if err != nil {
			return formula.Value{}, err
		}
		n, err := number.NewFloat(f)
		return formula.FloatValue(n), err
	case "text":
		unquoted, err := strconv.Unquote(body)
		return formula.TextValue(unquoted), err
	case "bool":
		return formula.BooleanValue(body == "true"), nil
	case "date":
		d, err := datefn.Parse(body)
		return formula.DateValue(d), err
	case "list<int>":
		var items []int64
		for _, part := range listParts(body) {
			n, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return formula.Value{}, err
			}
			items = append(items, n)
		}
		return formula.IntegerListValue(items), nil
	case "list<dec>":
		var items []number.Decimal
		for _, part := range listParts(body) {
			d, err := number.ParseDecimal(part)
			if err != nil {
				return formula.Value{}, err
			}
			items = append(items, d)
		}
		return formula.DecimalListValue(items), nil
	}
	return formula.Value{}, fmt.Errorf("value %q has the unknown type %q", text, prefix)
}

func listParts(body string) []string {
	inner := strings.TrimSuffix(strings.TrimPrefix(body, "["), "]")
	if strings.TrimSpace(inner) == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func shortType(t formula.Type) string {
	switch t {
	case formula.TypeInteger:
		return "int"
	case formula.TypeDecimal:
		return "dec"
	case formula.TypeFloat:
		return "float"
	case formula.TypeText:
		return "text"
	case formula.TypeBoolean:
		return "bool"
	case formula.TypeDate:
		return "date"
	case formula.TypeListInteger:
		return "list<int>"
	case formula.TypeListDecimal:
		return "list<dec>"
	}
	return string(t)
}

func shown(v formula.Value) string {
	if v.Type() == formula.TypeText {
		return "text:" + strconv.Quote(v.String())
	}
	return shortType(v.Type()) + ":" + v.String()
}
