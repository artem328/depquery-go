package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type ResolveMethodID int

func (id ResolveMethodID) String() string {
	return "plan.ResolveMethodID(" + strconv.Itoa(int(id)) + ")"
}

type ResolveMethod interface {
	GetID() ResolveMethodID
	isResolveMethod()
}

type RelationResolveMethod struct {
	ID       ResolveMethodID
	Relation RelationID
}

func (m RelationResolveMethod) GetID() ResolveMethodID {
	return m.ID
}

func (RelationResolveMethod) isResolveMethod() {}

type NestedResolveMethod struct {
	ID     ResolveMethodID
	Nested NestedID
}

func (m NestedResolveMethod) GetID() ResolveMethodID {
	return m.ID
}

func (NestedResolveMethod) isResolveMethod() {}

type EntityResolverID int

func (id EntityResolverID) String() string {
	return "plan.EntityResolverID(" + strconv.Itoa(int(id)) + ")"
}

type EntityResolver struct {
	ID                EntityResolverID
	Entity            semantic.EntityID
	ParentFetchGetter ParentFetchGetterID
	Resolutions       []EntityResolution
}

type EntityResolution interface {
	isEntityResolution()
}

type EntityResolutionRelation struct {
	Relation      RelationID
	ResolveMethod ResolveMethodID
	EntityFetch   EntityFetchID
}

func (e EntityResolutionRelation) isEntityResolution() {}

type EntityResolutionVariant struct {
	Variant     semantic.VariantID
	Resolutions []EntityResolution
}

func (EntityResolutionVariant) isEntityResolution() {}

type NestedResolverID int

func (id NestedResolverID) String() string {
	return "plan.NestedResolverID(" + strconv.Itoa(int(id)) + ")"
}

type NestedResolver struct {
	ID                   NestedResolverID
	Entity               semantic.EntityID
	ParentFetchGetter    ParentFetchGetterID
	Resolutions          []NestedResolution
	SyntheticIDNamespace uint64
}

type NestedResolution struct {
	Nested            NestedID
	Synthetic         bool
	ResolveMethod     ResolveMethodID
	NestedEntityFetch NestedEntityFetchID
}
