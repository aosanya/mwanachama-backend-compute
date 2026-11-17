package basicruns

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

type Outcome struct {
	Codes []string
	Value formula.Value
	Done  bool
}

func codesOf(problems []formula.Problem) []string {
	codes := make([]string, len(problems))
	for i, problem := range problems {
		codes[i] = problem.Code
	}
	return codes
}

func execute(c Case, runtime formula.Runtime) (Outcome, error) {
	list, err := parseInputs(c.Inputs)
	if err != nil {
		return Outcome{}, err
	}
	tree, problems := formula.Parse(c.Formula)
	if len(problems) != 0 {
		return Outcome{Codes: codesOf(problems)}, nil
	}
	declared := make([]formula.Input, len(list))
	values := make(map[string]formula.Value, len(list))
	for i, item := range list {
		declared[i] = formula.Input{Name: item.name, Type: item.value.Type()}
		values[item.name] = item.value
	}
	program, problems := formula.Plan(tree, declared, runtime, formula.DefaultOptions())
	if len(problems) != 0 {
		return Outcome{Codes: codesOf(problems)}, nil
	}
	value, failure := program.Run(values)
	if failure != nil {
		return Outcome{Codes: []string{failure.Code}}, nil
	}
	return Outcome{Value: value, Done: true}, nil
}

func judge(c Case, outcome Outcome) error {
	switch {
	case strings.HasPrefix(c.Expect, "result: "):
		want := strings.TrimPrefix(c.Expect, "result: ")
		if !outcome.Done {
			return fmt.Errorf("want result %s, got problems %v", want, outcome.Codes)
		}
		if got := shown(outcome.Value); got != want {
			return fmt.Errorf("result = %s, want %s", got, want)
		}
	case strings.HasPrefix(c.Expect, "problems: "):
		want := splitTopLevel(strings.TrimPrefix(c.Expect, "problems: "), ", ")
		if outcome.Done {
			return fmt.Errorf("want problems %v, got result %s", want, shown(outcome.Value))
		}
		if !reflect.DeepEqual(outcome.Codes, want) {
			return fmt.Errorf("problems = %v, want %v", outcome.Codes, want)
		}
	case c.Expect == "accepted":
		if !outcome.Done {
			return fmt.Errorf("want the formula accepted, got problems %v", outcome.Codes)
		}
	default:
		return fmt.Errorf("expected cell %q is not result, problems or accepted", c.Expect)
	}
	return nil
}
