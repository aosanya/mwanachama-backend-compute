package cel

import (
	"cel.dev/cel-go/cel"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

var _ harness.TreeEngine = engine{}

func parse(c harness.Case) (*cel.Ast, error) {
	env, err := cel.NewEnv()
	if err != nil {
		return nil, err
	}
	parsed, issues := env.Parse(translate(c.Source))
	if err := issues.Err(); err != nil {
		return nil, err
	}
	return parsed, nil
}

func (engine) Tree(c harness.Case) (any, error) {
	parsed, err := parse(c)
	if err != nil {
		return nil, err
	}
	return parsed.NativeRep().Expr(), nil
}
