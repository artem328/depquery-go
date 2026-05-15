package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type EntityFetchID int

func (id EntityFetchID) String() string {
	return "plan.EntityFetchID(" + strconv.FormatInt(int64(id), 10) + ")"
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

type ParentFetchGetter struct {
	FetchParent    FetchParentID
	StateContainer StateContainerID
}

type FetchContextRootID int

func (id FetchContextRootID) String() string {
	return "plan.FetchContextRootID(" + strconv.FormatInt(int64(id), 10) + ")"
}

type FetchContextRoot struct {
	ID             FetchContextRootID
	Entity         semantic.EntityID
	FetchParent    FetchParentID
	StateContainer StateContainerID
}
