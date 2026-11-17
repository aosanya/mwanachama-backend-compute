package examplecases

import (
	"fmt"
	"strings"
)

func Mismatch(c Case, outcome Outcome) string {
	where := fmt.Sprintf("%s:%d", c.File, c.Line)
	if c.Expected.Result != nil {
		if outcome.Result == nil {
			return fmt.Sprintf("%s: expected %s, got problems %v: %v", where, c.Expected.Result, outcome.Problems, outcome.Where)
		}
		if outcome.Result.String() != c.Expected.Result.String() {
			return fmt.Sprintf("%s: expected %s, got %s", where, c.Expected.Result, outcome.Result)
		}
		return ""
	}
	if outcome.Result != nil {
		return fmt.Sprintf("%s: expected problems %v, got the result %s", where, c.Expected.Problems, outcome.Result)
	}
	if strings.Join(outcome.Problems, ", ") != strings.Join(c.Expected.Problems, ", ") {
		return fmt.Sprintf("%s: expected problems %v, got %v: %v", where, c.Expected.Problems, outcome.Problems, outcome.Where)
	}
	return ""
}

func Failing(cases []Case, held map[string]string, runtime *StandIn) []string {
	var failing []string
	for _, c := range cases {
		if _, isHeld := held[c.Name]; isHeld {
			continue
		}
		outcome, err := Evaluate(c, runtime)
		if err != nil || Mismatch(c, outcome) != "" {
			failing = append(failing, c.Name)
		}
	}
	return failing
}
