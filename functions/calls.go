package functions

import (
	"strconv"
	"strings"
	"sync"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

type Calls struct {
	mu       sync.Mutex
	prepared map[string]Callable
	building map[string]bool
}

func NewCalls() *Calls {
	return &Calls{prepared: map[string]Callable{}, building: map[string]bool{}}
}

func (c *Calls) callable(library string, name string, version int) (Callable, bool) {
	key := Function{Library: library, Name: name, Version: version}.Key()
	c.mu.Lock()
	if ready, found := c.prepared[key]; found {
		c.mu.Unlock()
		return ready, true
	}
	if c.building[key] {
		c.mu.Unlock()
		return Callable{}, false
	}
	c.building[key] = true
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.building, key)
		c.mu.Unlock()
	}()
	candidate, found := Find(key)
	if !found {
		return Callable{}, false
	}
	ready, err := candidate.PrepareIn(c)
	if err != nil {
		return Callable{}, false
	}
	c.mu.Lock()
	c.prepared[key] = ready
	c.mu.Unlock()
	return ready, true
}

func (c *Calls) Function(library string, name string, version int) (formula.Signature, bool) {
	ready, found := c.callable(library, name, version)
	if !found {
		return formula.Signature{}, false
	}
	return ready.signature(c), true
}

func (c *Calls) Versions(library string, name string) []int {
	var versions []int
	for _, candidate := range All() {
		if candidate.Library == library && candidate.Name == name {
			versions = append(versions, candidate.Version)
		}
	}
	return versions
}

func (c *Calls) Apply(library string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	return c.ApplyMetered(library, name, version, args, formula.NewMeter(formula.DefaultOptions().Limits.Cost))
}

func (c *Calls) ApplyMetered(library string, name string, version int, args []formula.Value, meter *formula.Meter) (formula.Value, *formula.Problem) {
	ready, found := c.callable(library, name, version)
	if !found {
		return formula.Value{}, &formula.Problem{Code: "unknown-function", Sentence: "`" + library + "." + name + "@" + strconv.Itoa(version) + "` is not in the library."}
	}
	names := ready.function.ParamOrder()
	if len(args) != len(names) {
		return formula.Value{}, &formula.Problem{Code: "bad-input", Sentence: ready.function.Key() + " takes " + strconv.Itoa(len(names)) + " values and was given " + strconv.Itoa(len(args)) + "."}
	}
	row := function.Row{}
	for i, input := range names {
		row[input] = receivedOf(args[i])
	}
	if problems := ready.validator.Validate(row, nil); len(problems) > 0 {
		sentences := make([]string, len(problems))
		for i, refused := range problems {
			sentences[i] = refused.Sentence
		}
		return formula.Value{}, &formula.Problem{Code: problems[0].Code, Sentence: strings.Join(sentences, " ")}
	}
	return ready.program.ApplyMetered(args, meter)
}

func (c Callable) signature(calls *Calls) formula.Signature {
	params := make([]formula.Param, len(c.function.Definition.Inputs))
	for i, input := range c.function.Definition.Inputs {
		params[i] = formula.Param{Name: input.Name, Type: input.Type}
	}
	return formula.Signature{Params: params, Result: c.function.Result, Cost: c.program.Cost(), CallDepth: c.callDepth(calls)}
}

func (c Callable) callDepth(calls *Calls) int {
	tree, problems := formula.Parse(c.function.Body)
	if len(problems) > 0 {
		return 0
	}
	deepest := 0
	var walk func(node formula.Node)
	walk = func(node formula.Node) {
		if call, isCall := node.(formula.Call); isCall && call.Library != nil && call.Version != nil {
			if callee, found := calls.callable(*call.Library, call.Name, *call.Version); found {
				if depth := callee.callDepth(calls) + 1; depth > deepest {
					deepest = depth
				}
			}
		}
		for _, child := range node.Children() {
			walk(child)
		}
	}
	walk(tree.Root)
	return deepest
}

func receivedOf(value formula.Value) numtext.Received {
	if value.Type() != formula.TypeListInteger && value.Type() != formula.TypeListDecimal {
		return numtext.One(value.String())
	}
	items := make([]string, value.Len())
	for i := range items {
		items[i] = value.Item(i).String()
	}
	return numtext.Many(items...)
}
