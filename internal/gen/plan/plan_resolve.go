package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type ResolveMethodID int

func (id ResolveMethodID) String() string {
	return "plan.ResolveMethodID(" + strconv.Itoa(int(id)) + ")"
}

type ResolveMethod struct {
	ID       ResolveMethodID
	Relation RelationID
}

type EntityResolverID int

func (id EntityResolverID) String() string {
	return "plan.EntityResolverID(" + strconv.Itoa(int(id)) + ")"
}

type EntityResolver struct {
	ID          EntityResolverID
	Entity      semantic.EntityID
	FetchParent FetchParentID
	Resolutions []EntityResolution
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
