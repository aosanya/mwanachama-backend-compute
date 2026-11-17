package harness

import (
	"fmt"
	"maps"
	"sort"
	"strings"
	"testing"
)

type Result struct {
	Number   int
	Title    string
	Examples int
	Passed   bool
	Reason   string
}

type Report struct {
	Engine  string
	Results []Result
}

func NewReport(e Engine) Report {
	return Report{Engine: e.Name(), Results: Run(e)}
}

func (r Report) PassCount() int {
	count := 0
	for _, result := range r.Results {
		if result.Passed {
			count++
		}
	}
	return count
}

func (r Report) AllPassed() bool {
	return r.PassCount() == len(r.Results)
}

func (r Report) String() string {
	summary := fmt.Sprintf("%s: %d of %d formulas pass", r.Engine, r.PassCount(), len(r.Results))
	return Format(r.Results) + summary + "\n"
}

func Run(e Engine) []Result {
	library := Library{}
	var results []Result
	for _, c := range inDependencyOrder(Cases()) {
		results = append(results, runCase(e, c, library))
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Number < results[j].Number })
	return results
}

func Format(results []Result) string {
	var out strings.Builder
	for _, r := range results {
		if r.Passed {
			fmt.Fprintf(&out, "PASS %2d  %s  (%d examples)\n", r.Number, r.Title, r.Examples)
			continue
		}
		fmt.Fprintf(&out, "FAIL %2d  %s  %s\n", r.Number, r.Title, r.Reason)
	}
	return out.String()
}

func RunTest(t *testing.T, e Engine) {
	t.Helper()
	for _, r := range Run(e) {
		t.Run(fmt.Sprintf("formula_%02d", r.Number), func(t *testing.T) {
			if !r.Passed {
				t.Errorf("%s: %s", r.Title, r.Reason)
			}
		})
	}
}

func runCase(e Engine, c Case, library Library) Result {
	result := Result{Number: c.Number, Title: c.Title, Examples: len(c.Examples)}
	if missing := missingDependency(c, library); missing != "" {
		result.Reason = fmt.Sprintf("needs %s, which is not in the library", missing)
		return result
	}
	program, err := compile(e, c, library)
	if err != nil {
		result.Reason = "compile: " + err.Error()
		return result
	}
	if c.PublishedAs != "" {
		library[c.PublishedAs] = program
	}
	failures := failedExamples(program, c)
	result.Passed = len(failures) == 0
	result.Reason = strings.Join(failures, "; ")
	return result
}

func failedExamples(program Program, c Case) []string {
	var failures []string
	for _, example := range c.Examples {
		got, err := run(program, example.Inputs)
		switch {
		case err != nil:
			failures = append(failures, fmt.Sprintf("example %q: run error: %v", example.Name, err))
		case !Equal(example.Want, got):
			failures = append(failures, fmt.Sprintf("example %q: want %s, got %s", example.Name, Describe(example.Want), Describe(got)))
		}
	}
	return failures
}

func compile(e Engine, c Case, library Library) (program Program, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	program, err = e.Compile(c, library)
	if err == nil && program == nil {
		err = fmt.Errorf("engine returned no program")
	}
	return program, err
}

func run(program Program, inputs map[string]any) (got any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return program.Run(maps.Clone(inputs))
}

func missingDependency(c Case, library Library) string {
	for _, name := range c.Uses {
		if _, ok := library[name]; !ok {
			return name
		}
	}
	return ""
}

func inDependencyOrder(cases []Case) []Case {
	published := map[string]bool{}
	for _, c := range cases {
		if c.PublishedAs != "" {
			published[c.PublishedAs] = true
		}
	}
	placed := map[string]bool{}
	var ordered []Case
	remaining := cases
	for len(remaining) > 0 {
		var ready, blocked []Case
		for _, c := range remaining {
			if dependenciesPlaced(c, published, placed) {
				ready = append(ready, c)
			} else {
				blocked = append(blocked, c)
			}
		}
		if len(ready) == 0 {
			return append(ordered, blocked...)
		}
		for _, c := range ready {
			placed[c.PublishedAs] = true
		}
		ordered = append(ordered, ready...)
		remaining = blocked
	}
	return ordered
}

func dependenciesPlaced(c Case, published, placed map[string]bool) bool {
	for _, name := range c.Uses {
		if published[name] && !placed[name] {
			return false
		}
	}
	return true
}
