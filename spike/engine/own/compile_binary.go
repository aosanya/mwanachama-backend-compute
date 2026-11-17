package own

import (
	"fmt"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

var comparisonTests = map[string]func(int) bool{
	"==": func(order int) bool { return order == 0 },
	"!=": func(order int) bool { return order != 0 },
	"<":  func(order int) bool { return order < 0 },
	"<=": func(order int) bool { return order <= 0 },
	">":  func(order int) bool { return order > 0 },
	">=": func(order int) bool { return order >= 0 },
}

func (c *compiler) binary(n Binary) (code, error) {
	left, err := c.expr(n.Left)
	if err != nil {
		return code{}, err
	}
	right, err := c.expr(n.Right)
	if err != nil {
		return code{}, err
	}
	switch n.Op {
	case "&&", "||":
		return logical(n, left, right)
	case "==", "!=", "<", "<=", ">", ">=":
		return comparison(n, left, right)
	}
	return arithmetic(n, left, right)
}

func logical(n Binary, left, right code) (code, error) {
	for _, side := range []code{left, right} {
		if side.kind != harness.KindBoolean {
			return code{}, typeError(n.Pos, fmt.Sprintf("%s needs booleans but was given %s", n.Op, withArticle(side.kind)))
		}
	}
	if n.Op == "&&" {
		return code{harness.KindBoolean, both(as[bool](left), as[bool](right))}, nil
	}
	return code{harness.KindBoolean, either(as[bool](left), as[bool](right))}, nil
}

func both(left, right thunk[bool]) thunk[bool] {
	return func(fr *frame) (bool, error) {
		ok, err := left(fr)
		if err != nil || !ok {
			return false, err
		}
		return right(fr)
	}
}

func either(left, right thunk[bool]) thunk[bool] {
	return func(fr *frame) (bool, error) {
		ok, err := left(fr)
		if err != nil || ok {
			return ok, err
		}
		return right(fr)
	}
}

func comparison(n Binary, left, right code) (code, error) {
	left, right, err := align(n.Pos, left, right)
	if err != nil {
		return code{}, err
	}
	ops := kindTable[left.kind]
	if ops.compare == nil {
		return code{}, typeError(n.Pos, fmt.Sprintf("%s cannot compare %s values", n.Op, left.kind))
	}
	if left.kind == harness.KindBoolean && n.Op != "==" && n.Op != "!=" {
		return code{}, typeError(n.Pos, n.Op+" cannot put booleans in order")
	}
	return code{harness.KindBoolean, ops.compare(left.fn, right.fn, comparisonTests[n.Op])}, nil
}
