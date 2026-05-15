package schema

import "golang.org/x/exp/constraints"

type Scalar interface {
	~string | constraints.Integer | constraints.Float | constraints.Complex
}

type Value[T Scalar] struct {
	V          T
	Definition Definition
}

func NewValue[T Scalar](v T, d Definition) Value[T] {
	return Value[T]{V: v, Definition: d}
}
