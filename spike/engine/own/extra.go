package own

import "github.com/aosanya/mwanachama-backend-compute/spike/harness"

type Function struct {
	Params []harness.Kind
	Result harness.Kind
	Call   func(meter Meter, args []any) (any, error)
}

func (f Function) overload() overload {
	return overload{
		params: f.Params,
		result: f.Result,
		build: func(args []code) any {
			boxed := make([]func(*frame) (any, error), len(args))
			for i, arg := range args {
				boxed[i] = kindTable[arg.kind].box(arg.fn)
			}
			return kindTable[f.Result].native(func(fr *frame) (any, error) {
				values := make([]any, len(boxed))
				for i, evaluate := range boxed {
					value, err := evaluate(fr)
					if err != nil {
						return nil, err
					}
					values[i] = value
				}
				return f.Call(Meter{fr.run}, values)
			})
		},
	}
}
