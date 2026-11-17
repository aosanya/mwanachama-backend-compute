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

type LimitError struct {
	Name    string
	Allowed int
	Found   int
	Cause   error
	message string
}

func (e *LimitError) Error() string { return e.message }

func (e *LimitError) Unwrap() error { return e.Cause }

func limitError(name string, allowed, found int, format string) *LimitError {
	return &LimitError{Name: name, Allowed: allowed, Found: found, message: fmt.Sprintf(format, found, allowed)}
}

func depthError(allowed, found int) *LimitError {
	return limitError("depth", allowed, found, "The formula is nested too deeply: %d levels where %d are allowed.")
}

func nodeError(allowed, found int) *LimitError {
	return limitError("node", allowed, found, "The formula is too large: %d parts where %d are allowed.")
}

func costError(allowed int) *LimitError {
	return &LimitError{Name: "cost", Allowed: allowed, Found: allowed + 1, message: fmt.Sprintf("The formula needs more than the %d steps allowed.", allowed)}
}

func timeError(cause error) *LimitError {
	return &LimitError{Name: "time", Cause: cause, message: "The formula ran out of time."}
}

func formulaLengthError(allowed, found int) *LimitError {
	return limitError("formula length", allowed, found, "The formula has %d characters where %d are allowed.")
}

func listLengthError(allowed, found int) *LimitError {
	return limitError("list length", allowed, found, "A list has %d entries where %d are allowed.")
}

func textLengthError(allowed, found int) *LimitError {
	return limitError("text length", allowed, found, "A text has %d characters where %d are allowed.")
}

func magnitudeError(allowed, found int) *LimitError {
	return limitError("number magnitude", allowed, found, "A number has %d digits where %d are allowed.")
}

func decimalPlacesError(allowed, found int) *LimitError {
	return limitError("decimal places", allowed, found, "A number has %d decimal places where %d are allowed.")
}

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
