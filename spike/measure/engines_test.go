package measure

import (
	"fmt"
	"sync"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/cel"
	"github.com/aosanya/mwanachama-backend-compute/spike/engine/exprlang"
	"github.com/aosanya/mwanachama-backend-compute/spike/engine/own"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func engines() []harness.Engine {
	return []harness.Engine{cel.New(), exprlang.New(), own.New()}
}

type compiledSet struct {
	library  harness.Library
	programs map[int]harness.Program
	cases    map[int]harness.Case
}

var (
	setMutex sync.Mutex
	sets     = map[string]*compiledSet{}
)

func compiledFor(e harness.Engine) (*compiledSet, error) {
	setMutex.Lock()
	defer setMutex.Unlock()
	if set, ok := sets[e.Name()]; ok {
		return set, nil
	}
	set := &compiledSet{
		library:  harness.Library{},
		programs: map[int]harness.Program{},
		cases:    map[int]harness.Case{},
	}
	for _, wave := range [][]harness.Case{casesWithout(true), casesWithout(false)} {
		for _, c := range wave {
			program, err := e.Compile(c, set.library)
			if err != nil {
				return nil, fmt.Errorf("%s formula %d: %w", e.Name(), c.Number, err)
			}
			set.programs[c.Number] = program
			set.cases[c.Number] = c
			if c.PublishedAs != "" {
				set.library[c.PublishedAs] = program
			}
		}
	}
	sets[e.Name()] = set
	return set, nil
}

func casesWithout(independent bool) []harness.Case {
	var picked []harness.Case
	for _, c := range harness.Cases() {
		if (len(c.Uses) == 0) == independent {
			picked = append(picked, c)
		}
	}
	return picked
}

func mustCompiledFor(tb testing.TB, e harness.Engine) *compiledSet {
	tb.Helper()
	set, err := compiledFor(e)
	if err != nil {
		tb.Fatal(err)
	}
	return set
}

func formulaName(number int) string {
	return fmt.Sprintf("f%02d", number)
}
