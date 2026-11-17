package batch

import (
	"strings"
	"sync"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

const notTrueOrFalseSentence = "The check did not give true or false."

type checkRunner struct {
	options  formula.Options
	mu       sync.RWMutex
	programs map[string]*formula.Program
}

func newCheckRunner(options formula.Options) *checkRunner {
	return &checkRunner{options: options, programs: map[string]*formula.Program{}}
}

func (r *checkRunner) Run(tree formula.Tree, values map[string]numtext.Value) function.Outcome {
	names := namesRead(tree.Root)
	inputs := make([]formula.Input, 0, len(names))
	converted := make(map[string]formula.Value, len(names))
	key := []string{formula.Format(tree)}
	for _, name := range names {
		value, found := values[name]
		kind, known := formulaTypes[value.Type]
		runValue, readable := formulaValue(value)
		if !found || !known || !readable {
			return function.Outcome{Failure: "The check cannot read `" + name + "`."}
		}
		inputs = append(inputs, formula.Input{Name: name, Type: kind})
		converted[name] = runValue
		key = append(key, name+":"+string(kind))
	}
	program, failure := r.program(strings.Join(key, "|"), tree, inputs)
	if failure != "" {
		return function.Outcome{Failure: failure}
	}
	result, problem := program.Run(converted)
	if problem != nil {
		return function.Outcome{Failure: problem.Sentence}
	}
	if result.Type() != formula.TypeBoolean {
		return function.Outcome{Failure: notTrueOrFalseSentence}
	}
	return function.Outcome{True: result.Boolean()}
}

func (r *checkRunner) program(key string, tree formula.Tree, inputs []formula.Input) (*formula.Program, string) {
	r.mu.RLock()
	program, found := r.programs[key]
	r.mu.RUnlock()
	if found {
		return program, ""
	}
	planned, problems := formula.Plan(tree, inputs, nil, r.options)
	if len(problems) > 0 {
		return nil, problems[0].Sentence
	}
	r.mu.Lock()
	r.programs[key] = planned
	r.mu.Unlock()
	return planned, ""
}

func namesRead(root formula.Node) []string {
	var names []string
	seen := map[string]bool{}
	var walk func(node formula.Node)
	walk = func(node formula.Node) {
		if name, isName := node.(formula.Name); isName {
			if !seen[name.Name] {
				seen[name.Name] = true
				names = append(names, name.Name)
			}
			return
		}
		for _, child := range node.Children() {
			walk(child)
		}
	}
	walk(root)
	return names
}
