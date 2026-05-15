package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type StateContainerID int

func (id StateContainerID) String() string {
	return "plan.StateContainerID(" + strconv.Itoa(int(id)) + ")"
}

type StateContainer struct {
	ID         StateContainerID
	Entity     semantic.EntityID
	ReversedBy []ReversedStateContainerID
}

type ReversedStateContainerID int

func (id ReversedStateContainerID) String() string {
	return "plan.ReversedStateContainerID(" + strconv.Itoa(int(id)) + ")"
}

type ReversedStateContainer struct {
	ID                 ReversedStateContainerID
	StateContainerID   StateContainerID
	HolderEntity       semantic.EntityID
	HolderEntityMember semantic.MemberID
}
