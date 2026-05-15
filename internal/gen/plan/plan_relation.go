package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type RelationID int

func (id RelationID) String() string {
	return "plan.RelationID(" + strconv.Itoa(int(id)) + ")"
}

type Relation struct {
	ID RelationID
	semantic.Relation
}
