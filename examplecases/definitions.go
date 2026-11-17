package examplecases

import (
	"errors"

	fdef "github.com/aosanya/mwanachama-backend-compute/function"
)

const (
	TargetRow        = "row"
	TargetDefinition = "definition"
)

type Binding struct {
	Case       string
	Definition string
	Target     string
	Line       int
}

type Definitions struct {
	Definitions map[string]fdef.Definition
	Bindings    map[string]Binding
}

var errNotBuilt = errors.New("the definition loader is not built")

func LoadDefinitions(path string) (Definitions, error) {
	return Definitions{}, errNotBuilt
}

func ParseDefinitions(file string, text string) (Definitions, error) {
	return Definitions{}, errNotBuilt
}
