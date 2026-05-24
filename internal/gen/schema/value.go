package schema

import "golang.org/x/exp/constraints"

type Scalar interface {
	constraints.Integer | constraints.Float | constraints.Complex | ~string | ~bool
}

type Value[T Scalar] struct {
	V          T
	Definition Definition
}

func NewValue[T Scalar](v T, d Definition) Value[T] {
	return Value[T]{V: v, Definition: d}
}
