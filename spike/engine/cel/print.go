package cel

import (
	"cel.dev/cel-go/cel"

	"github.com/aosanya/mwanachama-backend-compute/spike/fit/shape"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func (engine) Print(c harness.Case) (string, error) {
	parsed, err := parse(c)
	if err != nil {
		return "", err
	}
	text, err := cel.AstToString(parsed)
	if err != nil {
		return "", err
	}
	return Reverse(text), nil
}

func Reverse(text string) string {
	return shape.Reverse(text, decimalFunction)
}
