package plan

import "golang.org/x/exp/constraints"

type counter[T constraints.Integer] struct {
	val T
}

func (c *counter[T]) Next() T {
	n := c.val
	c.val++

	return n
}

func (c *counter[T]) Current() T {
	return c.val
}
