package own

import "github.com/aosanya/mwanachama-backend-compute/spike/harness"

const inlineSlots = 8

type frame struct {
	slots   []any
	backing [inlineSlots]any
}

func newFrame(size int) *frame {
	f := &frame{}
	if size <= inlineSlots {
		f.slots = f.backing[:size]
	} else {
		f.slots = make([]any, size)
	}
	return f
}

type thunk[T any] func(*frame) (T, error)

type code struct {
	kind harness.Kind
	fn   any
}

func as[T any](c code) thunk[T] { return c.fn.(thunk[T]) }

func constant[T any](value T) thunk[T] {
	return func(*frame) (T, error) { return value, nil }
}

func failed[T any](err error) (T, error) {
	var zero T
	return zero, err
}

func certain1[A, R any](f func(A) R) func(A) (R, error) {
	return func(a A) (R, error) { return f(a), nil }
}

func certain2[A, B, R any](f func(A, B) R) func(A, B) (R, error) {
	return func(a A, b B) (R, error) { return f(a, b), nil }
}

func apply1[A, R any](x thunk[A], f func(A) (R, error)) thunk[R] {
	return func(fr *frame) (R, error) {
		a, err := x(fr)
		if err != nil {
			return failed[R](err)
		}
		out, err := f(a)
		if err != nil {
			return failed[R](plain(err))
		}
		return out, nil
	}
}

func apply2[A, B, R any](x thunk[A], y thunk[B], f func(A, B) (R, error)) thunk[R] {
	return func(fr *frame) (R, error) {
		a, err := x(fr)
		if err != nil {
			return failed[R](err)
		}
		b, err := y(fr)
		if err != nil {
			return failed[R](err)
		}
		out, err := f(a, b)
		if err != nil {
			return failed[R](plain(err))
		}
		return out, nil
	}
}

func conditional[T any](test thunk[bool], then, other thunk[T]) thunk[T] {
	return func(fr *frame) (T, error) {
		chosen, err := test(fr)
		if err != nil {
			return failed[T](err)
		}
		if chosen {
			return then(fr)
		}
		return other(fr)
	}
}

func collect[T any](items []code) thunk[[]T] {
	parts := make([]thunk[T], len(items))
	for i, item := range items {
		parts[i] = as[T](item)
	}
	return func(fr *frame) ([]T, error) {
		out := make([]T, len(parts))
		for i, part := range parts {
			value, err := part(fr)
			if err != nil {
				return nil, err
			}
			out[i] = value
		}
		return out, nil
	}
}
