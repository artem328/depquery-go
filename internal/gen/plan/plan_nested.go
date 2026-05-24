package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type NestedID int

func (n NestedID) String() string {
	return "plan.NestedID(" + strconv.Itoa(int(n)) + ")"
}

type Nested struct {
	ID NestedID
	semantic.Nested
}

type SyntheticStateContainerID int

func (n SyntheticStateContainerID) String() string {
	return "plan.SyntheticStateContainerID(" + strconv.Itoa(int(n)) + ")"
}

type SyntheticStateContainer struct {
	ID          SyntheticStateContainerID
	Entity      semantic.EntityID
	IDNamespace uint64
}
