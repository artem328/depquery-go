package depquery

import (
	"iter"
)

func SliceIter[T any](s []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range s {
			if !yield(s[i]) {
				return
			}
		}
	}
}

func SingleIter[T any](s T) iter.Seq[T] {
	return func(yield func(T) bool) { yield(s) }
}

func NoIter[T any]() iter.Seq[T] {
	return func(func(T) bool) {}
}
