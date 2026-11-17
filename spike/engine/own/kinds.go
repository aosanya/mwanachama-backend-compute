package own

import (
	"cmp"
	"strings"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/builtins"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type setter func(from, to *frame) error

type kindOps struct {
	variable    func(index int) any
	conditional func(test thunk[bool], then, other any) any
	box         func(fn any) func(*frame) (any, error)
	store       func(fn any, index int) setter
	invoke      func(body any, setters []setter, size int) any
	compare     func(left, right any, test func(int) bool) any
}

var kindTable = map[harness.Kind]kindOps{
	harness.KindDecimal:     opsFor(numeric.Decimal.Cmp),
	harness.KindFloat:       opsFor(cmp.Compare[float64]),
	harness.KindInteger:     opsFor(cmp.Compare[int64]),
	harness.KindText:        opsFor(strings.Compare),
	harness.KindBoolean:     opsFor(compareBooleans),
	harness.KindDate:        opsFor(compareDates),
	harness.KindDecimalList: opsFor[[]numeric.Decimal](nil),
	harness.KindIntegerList: opsFor[[]int64](nil),
}

func opsFor[T any](order func(T, T) int) kindOps {
	ops := kindOps{
		variable: func(index int) any {
			return thunk[T](func(fr *frame) (T, error) { return fr.slots[index].(T), nil })
		},
		conditional: func(test thunk[bool], then, other any) any {
			return conditional(test, then.(thunk[T]), other.(thunk[T]))
		},
		box: func(fn any) func(*frame) (any, error) {
			body := fn.(thunk[T])
			return func(fr *frame) (any, error) {
				value, err := body(fr)
				if err != nil {
					return nil, err
				}
				return value, nil
			}
		},
		store: func(fn any, index int) setter {
			source := fn.(thunk[T])
			return func(from, to *frame) error {
				value, err := source(from)
				if err != nil {
					return err
				}
				to.slots[index] = value
				return nil
			}
		},
		invoke: func(body any, setters []setter, size int) any {
			run := body.(thunk[T])
			return thunk[T](func(from *frame) (T, error) {
				inner := newFrame(size)
				for _, set := range setters {
					if err := set(from, inner); err != nil {
						return failed[T](err)
					}
				}
				return run(inner)
			})
		},
	}
	if order != nil {
		ops.compare = func(left, right any, test func(int) bool) any {
			return apply2(left.(thunk[T]), right.(thunk[T]), func(a, b T) (bool, error) {
				return test(order(a, b)), nil
			})
		}
	}
	return ops
}

func compareBooleans(a, b bool) int {
	if a == b {
		return 0
	}
	return 1
}

func compareDates(a, b time.Time) int {
	return cmp.Compare(0, builtins.DaysBetween(a, b))
}
