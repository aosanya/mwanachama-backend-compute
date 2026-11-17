package own

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

const dateLayout = "2006-01-02"

func literalNode(t token) (Node, bool, error) {
	switch t.kind {
	case tokenInteger:
		return integerNode(t)
	case tokenDecimal:
		return decimalNode(t)
	case tokenFloat:
		return floatNode(t)
	case tokenDate:
		return dateNode(t)
	case tokenText:
		return TextLiteral{Pos: t.pos, Value: t.text}, true, nil
	}
	return nil, false, nil
}

func integerNode(t token) (Node, bool, error) {
	value, err := strconv.ParseInt(t.text, 10, 64)
	if err != nil {
		return nil, true, readError(t.pos, fmt.Sprintf("the number %s is too large", t.text))
	}
	return IntegerLiteral{Pos: t.pos, Value: value}, true, nil
}

func decimalNode(t token) (Node, bool, error) {
	value, err := numeric.Parse(t.text)
	if err != nil {
		return nil, true, readError(t.pos, fmt.Sprintf("the number %s cannot be read", t.text))
	}
	return DecimalLiteral{Pos: t.pos, Value: value}, true, nil
}

func floatNode(t token) (Node, bool, error) {
	value, err := strconv.ParseFloat(t.text, 64)
	if err != nil {
		return nil, true, readError(t.pos, fmt.Sprintf("the float %sf is too large", t.text))
	}
	return FloatLiteral{Pos: t.pos, Value: value}, true, nil
}

func dateNode(t token) (Node, bool, error) {
	value, err := time.Parse(dateLayout, t.text)
	if err != nil {
		return nil, true, readError(t.pos, fmt.Sprintf("%s is not a real date", t.text))
	}
	return DateLiteral{Pos: t.pos, Value: value.UTC()}, true, nil
}
