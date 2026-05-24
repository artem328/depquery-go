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

func SliceIter2[T any](s []T) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i := range s {
			if !yield(i, s[i]) {
				return
			}
		}
	}
}

func SingleIter2[T any](s T) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) { yield(0, s) }
}

func NoIter2[T any]() iter.Seq2[int, T] {
	return func(func(int, T) bool) {}
}
