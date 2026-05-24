package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type EntityFetchID int

func (id EntityFetchID) String() string {
	return "plan.EntityFetchID(" + strconv.Itoa(int(id)) + ")"
}

type EntityFetch struct {
	ID                     EntityFetchID
	Entity                 semantic.EntityID
	StateContainer         StateContainerID
	ReversedStateContainer ReversedStateContainerID
	PrefetchMethod         PrefetchMethodID
	Parent                 FetchParentID
	Child                  FetchChildID
	IsParent               bool
	Reversed               bool
}

type NestedEntityFetchID int

func (id NestedEntityFetchID) String() string {
	return "plan.NestedEntityFetchID(" + strconv.Itoa(int(id)) + ")"
}

type NestedEntityFetch struct {
	ID                        NestedEntityFetchID
	Entity                    semantic.EntityID
	Parent                    FetchParentID
	Synthetic                 bool
	StateContainer            StateContainerID
	SyntheticStateContainerID SyntheticStateContainerID
}

type FetchParentID int

func (id FetchParentID) String() string {
	return "plan.FetchParentID(" + strconv.Itoa(int(id)) + ")"
}

type FetchParent struct {
	ID               FetchParentID
	Entity           semantic.EntityID
	ReversedByEntity semantic.EntityID
	Reversed         bool
}

type FetchChildID int

func (id FetchChildID) String() string {
	return "plan.FetchChildID(" + strconv.Itoa(int(id)) + ")"
}

type FetchChild struct {
	ID               FetchChildID
	Entity           semantic.EntityID
	ReversedByEntity semantic.EntityID
	Reversed         bool
}

type ReversedFetchParent struct {
	FetchParent          FetchParentID
	StateContainer       StateContainerID
	ReversedFetchParents []FetchParentReverse
}

type FetchParentReverse struct {
	FetchParent    FetchParentID
	StateContainer ReversedStateContainerID
}

type ParentFetchGetterID int

func (id ParentFetchGetterID) String() string {
	return "plan.ParentFetchGetterID(" + strconv.Itoa(int(id)) + ")"
}

type ParentFetchGetter struct {
	ID                      ParentFetchGetterID
	FetchParent             FetchParentID
	Synthetic               bool
	StateContainer          StateContainerID
	SyntheticStateContainer SyntheticStateContainerID
}

type FetchContextRootID int

func (id FetchContextRootID) String() string {
	return "plan.FetchContextRootID(" + strconv.Itoa(int(id)) + ")"
}

type FetchContextRoot struct {
	ID                      FetchContextRootID
	Entity                  semantic.EntityID
	Synthetic               bool
	FetchParent             FetchParentID
	StateContainer          StateContainerID
	SyntheticStateContainer SyntheticStateContainerID
}
