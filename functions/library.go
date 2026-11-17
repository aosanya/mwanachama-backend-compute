package functions

import (
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
)

type Function struct {
	Library    string
	Name       string
	Version    int
	Definition function.Definition
	Body       string
	Result     formula.Type
}

func (f Function) Key() string {
	return f.Library + "." + f.Name + "@" + strconv.Itoa(f.Version)
}

func (f Function) ParamOrder() []string {
	names := make([]string, len(f.Definition.Inputs))
	for i, input := range f.Definition.Inputs {
		names[i] = input.Name
	}
	return names
}

func All() []Function {
	return nil
}

func Find(key string) (Function, bool) {
	for _, candidate := range All() {
		if candidate.Key() == key {
			return candidate, true
		}
	}
	return Function{}, false
}
