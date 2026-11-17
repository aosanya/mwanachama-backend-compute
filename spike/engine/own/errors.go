package own

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/spike/builtins"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

var (
	errDivideByZero = errors.New("The formula divides by zero.")
	errEmptyList    = errors.New("The formula takes a value from an empty list.")
	errOverflow     = errors.New("The formula produced a whole number that is too large.")
	errNotFinite    = errors.New("The formula produced a number that is not finite.")
)

func readError(pos Pos, what string) error {
	return fmt.Errorf("The formula cannot be read: %s at line %d, column %d.", what, pos.Line, pos.Column)
}

func typeError(pos Pos, what string) error {
	return fmt.Errorf("The formula cannot be used: %s at line %d, column %d.", what, pos.Line, pos.Column)
}

func unknownFunction(name string) error {
	return fmt.Errorf("The function %s is not known.", name)
}

func unknownName(name string) error {
	return fmt.Errorf("The name %s is not known.", name)
}

func plain(err error) error {
	switch {
	case errors.Is(err, numeric.ErrDivideByZero):
		return errDivideByZero
	case errors.Is(err, builtins.ErrEmptyList):
		return errEmptyList
	case errors.Is(err, numeric.ErrNotFinite):
		return errNotFinite
	}
	return err
}

func withArticle(kind harness.Kind) string {
	name := kind.String()
	if strings.HasPrefix(name, "i") {
		return "an " + name
	}
	return "a " + name
}
