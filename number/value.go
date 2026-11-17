package number

import "fmt"

type Kind int

const (
	KindDecimal Kind = iota
	KindFloat
)

func (k Kind) String() string {
	if k == KindFloat {
		return "float"
	}
	return "decimal"
}

type Value struct {
	kind Kind
	d    Decimal
	f    Float
}

func DecimalValue(d Decimal) Value { return Value{kind: KindDecimal, d: d} }

func FloatValue(f Float) Value { return Value{kind: KindFloat, f: f} }

func (v Value) Kind() Kind { return v.kind }

func (v Value) Decimal() Decimal { return v.d }

func (v Value) Float() Float { return v.f }

func (v Value) String() string {
	if v.kind == KindFloat {
		return fmt.Sprint(v.f.Value())
	}
	return v.d.String()
}

type KindMixError struct {
	Left  Kind
	Right Kind
}

func (e *KindMixError) Error() string {
	return fmt.Sprintf("A %s and a %s cannot be combined; convert one with to_float or to_decimal.", e.Left, e.Right)
}

func sameKind(a, b Value) error {
	if a.kind != b.kind {
		return &KindMixError{Left: a.kind, Right: b.kind}
	}
	return nil
}

func Add(a, b Value) (Value, error) {
	if err := sameKind(a, b); err != nil {
		return Value{}, err
	}
	if a.kind == KindDecimal {
		return DecimalValue(a.d.Add(b.d)), nil
	}
	sum, err := a.f.Add(b.f)
	return FloatValue(sum), err
}

func Sub(a, b Value) (Value, error) {
	if err := sameKind(a, b); err != nil {
		return Value{}, err
	}
	if a.kind == KindDecimal {
		return DecimalValue(a.d.Sub(b.d)), nil
	}
	difference, err := a.f.Sub(b.f)
	return FloatValue(difference), err
}

func Mul(a, b Value) (Value, error) {
	if err := sameKind(a, b); err != nil {
		return Value{}, err
	}
	if a.kind == KindDecimal {
		return DecimalValue(a.d.Mul(b.d)), nil
	}
	product, err := a.f.Mul(b.f)
	return FloatValue(product), err
}

func Compare(a, b Value) (int, error) {
	if err := sameKind(a, b); err != nil {
		return 0, err
	}
	if a.kind == KindDecimal {
		return a.d.Cmp(b.d), nil
	}
	switch {
	case a.f.Value() < b.f.Value():
		return -1, nil
	case a.f.Value() > b.f.Value():
		return 1, nil
	}
	return 0, nil
}
