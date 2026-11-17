package stub

import (
	"errors"
	"fmt"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

var ErrDeliberate = errors.New("deliberate run failure")

type Engine struct {
	WrongAt        int
	RunErrorAt     int
	CompileErrorAt int
}

func (Engine) Name() string {
	return "stub"
}

func (e Engine) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	if c.Number == e.CompileErrorAt {
		return nil, fmt.Errorf("deliberate compile failure for formula %d", c.Number)
	}
	answers, ok := answersBySource()[c.Source]
	if !ok {
		return nil, fmt.Errorf("unknown source %q", c.Source)
	}
	return program{answers: answers, wrong: c.Number == e.WrongAt, failing: c.Number == e.RunErrorAt}, nil
}

type program struct {
	answers map[string]any
	wrong   bool
	failing bool
}

func (p program) Run(inputs map[string]any) (any, error) {
	if p.failing {
		return nil, ErrDeliberate
	}
	want, ok := p.answers[harness.InputKey(inputs)]
	if !ok {
		return nil, fmt.Errorf("no known answer for inputs %s", harness.InputKey(inputs))
	}
	if p.wrong {
		return wrongAnswer(want), nil
	}
	return want, nil
}

func answersBySource() map[string]map[string]any {
	bySource := map[string]map[string]any{}
	for _, c := range harness.Cases() {
		answers := map[string]any{}
		for _, example := range c.Examples {
			answers[harness.InputKey(example.Inputs)] = example.Want
		}
		bySource[c.Source] = answers
	}
	return bySource
}

func wrongAnswer(v any) any {
	switch x := v.(type) {
	case numeric.Decimal:
		return x.Add(numeric.FromInt(1))
	case float64:
		return x + 1
	case int64:
		return x + 1
	case string:
		return x + "?"
	case bool:
		return !x
	case time.Time:
		return x.AddDate(0, 0, 1)
	}
	return nil
}
