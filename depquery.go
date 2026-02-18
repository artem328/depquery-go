package depquery

import "fmt"

type MaybeFetched[T any] struct {
	val     T
	fetched bool
}

func Fetched[T any](v T) MaybeFetched[T] {
	return MaybeFetched[T]{val: v, fetched: true}
}

func (mf MaybeFetched[T]) Value() T {
	return mf.val
}

func (mf MaybeFetched[T]) Fetched() bool {
	return mf.fetched
}

func (mf MaybeFetched[T]) Maybe() (T, bool) {
	return mf.val, mf.fetched
}

type ID uint

type Builder struct {
	ID        ID
	ParentID  ID
	Relations uint64
}

type NotInStateError[T comparable] struct {
	id   T
	name string
	msg  string
}

func NewNotInStateError[T comparable](name string, id T) NotInStateError[T] {
	return NotInStateError[T]{id: id, name: name, msg: fmt.Sprintf(": not found %s (%v) in state", name, id)}
}

func (e NotInStateError[T]) Error() string {
	return "not in state" + e.msg
}
