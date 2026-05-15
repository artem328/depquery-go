package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type PrefetchMethodID int

func (id PrefetchMethodID) String() string {
	return "plan.PrefetchMethodID(" + strconv.Itoa(int(id)) + ")"
}

type PrefetchMethod struct {
	ID               PrefetchMethodID
	Entity           semantic.EntityID
	ReversedByEntity semantic.EntityID
	Reversed         bool
}
