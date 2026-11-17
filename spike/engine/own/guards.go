package own

import "github.com/aosanya/mwanachama-backend-compute/spike/harness"

var placesArgument = map[string]int{"round": 1, "to_decimal": 1}

var listChargers = map[harness.Kind]func(any) any{
	harness.KindDecimalList: chargeDecimalList,
	harness.KindIntegerList: chargeIntegerList,
}

func (c *compiler) guardArguments(n Call, args []code) error {
	if c.metered {
		for i, arg := range args {
			if charge, isList := listChargers[arg.kind]; isList {
				args[i] = code{arg.kind, charge(arg.fn)}
			}
		}
	}
	index, hasPlaces := placesArgument[n.Name]
	if !hasPlaces || c.options.MaxDecimals <= 0 || index >= len(args) {
		return nil
	}
	if literal, ok := n.Args[index].(IntegerLiteral); ok && literal.Value > int64(c.options.MaxDecimals) {
		return decimalPlacesError(c.options.MaxDecimals, int(min(literal.Value, 1<<31)))
	}
	args[index] = code{args[index].kind, limitPlaces(args[index].fn, c.options.MaxDecimals)}
	return nil
}
