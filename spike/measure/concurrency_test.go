package measure

import (
	"fmt"
	"sync"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const (
	goroutineCount = 64
	runsPerRoutine = 2000
)

var concurrentFormulas = []int{3, 5, 6, 7, 8, 15}

type mismatch struct {
	goroutine int
	run       int
	detail    string
}

func TestRunDoesNotChangeItsInputs(t *testing.T) {
	for _, e := range engines() {
		t.Run(e.Name(), func(t *testing.T) {
			set := mustCompiledFor(t, e)
			for number, c := range set.cases {
				for _, example := range c.Examples {
					before := harness.InputKey(example.Inputs)
					if _, err := set.programs[number].Run(example.Inputs); err != nil {
						t.Fatalf("formula %d example %q: %v", number, example.Name, err)
					}
					if after := harness.InputKey(example.Inputs); after != before {
						t.Errorf("formula %d example %q: inputs changed from %s to %s", number, example.Name, before, after)
					}
				}
			}
		})
	}
}

func TestOneProgramFromManyGoroutines(t *testing.T) {
	for _, e := range engines() {
		for _, number := range concurrentFormulas {
			t.Run(fmt.Sprintf("%s/%s", e.Name(), formulaName(number)), func(t *testing.T) {
				set := mustCompiledFor(t, e)
				program := set.programs[number]
				inputs := make([][]map[string]any, goroutineCount)
				want := make([][]any, goroutineCount)
				for g := range inputs {
					inputs[g] = makeInputs(number, int64(1000*number+g), runsPerRoutine)
					want[g] = make([]any, runsPerRoutine)
					for i, in := range inputs[g] {
						got, err := program.Run(in)
						if err != nil {
							t.Fatalf("single-threaded run %d/%d: %v", g, i, err)
						}
						want[g][i] = got
					}
				}
				start := make(chan struct{})
				var group sync.WaitGroup
				failures := make(chan mismatch, goroutineCount)
				for g := 0; g < goroutineCount; g++ {
					group.Add(1)
					go func() {
						defer group.Done()
						<-start
						for i, in := range inputs[g] {
							got, err := program.Run(in)
							switch {
							case err != nil:
								failures <- mismatch{g, i, "run error: " + err.Error()}
								return
							case !harness.Equal(want[g][i], got):
								failures <- mismatch{g, i, fmt.Sprintf("want %s, got %s", harness.Describe(want[g][i]), harness.Describe(got))}
								return
							}
						}
					}()
				}
				close(start)
				group.Wait()
				close(failures)
				for f := range failures {
					t.Errorf("goroutine %d run %d: %s", f.goroutine, f.run, f.detail)
				}
			})
		}
	}
}

func TestSharedLibraryProgramsFromManyGoroutines(t *testing.T) {
	for _, e := range engines() {
		t.Run(e.Name(), func(t *testing.T) {
			set := mustCompiledFor(t, e)
			for name, shared := range set.library {
				number := publishedNumber(set, name)
				inputs := make([][]map[string]any, goroutineCount)
				want := make([][]any, goroutineCount)
				for g := range inputs {
					inputs[g] = makeInputs(number, int64(7000*number+g), runsPerRoutine)
					want[g] = make([]any, runsPerRoutine)
					for i, in := range inputs[g] {
						got, err := shared.Run(in)
						if err != nil {
							t.Fatalf("%s single-threaded run %d/%d: %v", name, g, i, err)
						}
						want[g][i] = got
					}
				}
				callers := set.programs[callerOf(set, name)]
				callerNumber := callerOf(set, name)
				callerInputs := make([][]map[string]any, goroutineCount)
				callerWant := make([][]any, goroutineCount)
				for g := range callerInputs {
					callerInputs[g] = makeInputs(callerNumber, int64(9000*callerNumber+g), runsPerRoutine)
					callerWant[g] = make([]any, runsPerRoutine)
					for i, in := range callerInputs[g] {
						got, err := callers.Run(in)
						if err != nil {
							t.Fatalf("caller of %s single-threaded run %d/%d: %v", name, g, i, err)
						}
						callerWant[g][i] = got
					}
				}
				var group sync.WaitGroup
				failures := make(chan mismatch, 2*goroutineCount)
				for g := 0; g < goroutineCount; g++ {
					group.Add(2)
					go func() {
						defer group.Done()
						for i, in := range inputs[g] {
							got, err := shared.Run(in)
							if err != nil || !harness.Equal(want[g][i], got) {
								failures <- mismatch{g, i, fmt.Sprintf("%s direct: want %s, got %s, err %v", name, harness.Describe(want[g][i]), harness.Describe(got), err)}
								return
							}
						}
					}()
					go func() {
						defer group.Done()
						for i, in := range callerInputs[g] {
							got, err := callers.Run(in)
							if err != nil || !harness.Equal(callerWant[g][i], got) {
								failures <- mismatch{g, i, fmt.Sprintf("%s via caller: want %s, got %s, err %v", name, harness.Describe(callerWant[g][i]), harness.Describe(got), err)}
								return
							}
						}
					}()
				}
				group.Wait()
				close(failures)
				for f := range failures {
					t.Errorf("goroutine %d run %d: %s", f.goroutine, f.run, f.detail)
				}
			}
		})
	}
}

func publishedNumber(set *compiledSet, name string) int {
	for number, c := range set.cases {
		if c.PublishedAs == name {
			return number
		}
	}
	return 0
}

func callerOf(set *compiledSet, name string) int {
	for number, c := range set.cases {
		for _, used := range c.Uses {
			if used == name {
				return number
			}
		}
	}
	return 0
}
