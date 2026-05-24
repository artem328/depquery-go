package plan

import (
	"strconv"

	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type BuilderID int

func (id BuilderID) String() string {
	return "plan.BuilderID(" + strconv.Itoa(int(id)) + ")"
}

type Builder interface {
	GetID() BuilderID
	GetMethods() []BuilderMethod
	GetChildBuilders() []ChildBuilder

	isBuilder()
}

type ChildBuilder interface {
	GetBuilderID() BuilderID
	isChildBuilder()
}

type RegularChildBuilder struct {
	Builder BuilderID
}

func (cb RegularChildBuilder) GetBuilderID() BuilderID {
	return cb.Builder
}

func (RegularChildBuilder) isChildBuilder() {}

type RelationChildBuilder struct {
	Builder  BuilderID
	Relation RelationID
}

func (cb RelationChildBuilder) GetBuilderID() BuilderID {
	return cb.Builder
}

func (RelationChildBuilder) isChildBuilder() {}

type NestedChildBuilder struct {
	Builder BuilderID
	Nested  NestedID
}

func (cb NestedChildBuilder) GetBuilderID() BuilderID {
	return cb.Builder
}

func (NestedChildBuilder) isChildBuilder() {}

type CommonBuilder struct {
	ID            BuilderID
	Methods       []BuilderMethod
	ChildBuilders []ChildBuilder
}

func (b CommonBuilder) GetID() BuilderID {
	return b.ID
}

func (b CommonBuilder) GetMethods() []BuilderMethod {
	return b.Methods
}

func (b CommonBuilder) GetChildBuilders() []ChildBuilder {
	return b.ChildBuilders
}

type RootBuilder struct {
	Entity            semantic.EntityID
	FetchContextRoot  FetchContextRootID
	EntityResolver    EntityResolverID
	NestedResolver    NestedResolverID
	IsRelationBuilder bool
	IsNestedBuilder   bool
	CommonBuilder
}

func (RootBuilder) isBuilder() {}

type VariantBuilder struct {
	Variant semantic.VariantID
	Parent  BuilderID
	CommonBuilder
}

func (VariantBuilder) isBuilder() {}

type BuilderMethodID int

func (id BuilderMethodID) String() string {
	return "plan.BuilderMethodID(" + strconv.Itoa(int(id)) + ")"
}

type BuilderMethod interface {
	GetID() BuilderMethodID

	isBuilderMethod()
}

type EnableBuilderMethod struct {
	ID       BuilderMethodID
	Relation RelationID
}

func (m EnableBuilderMethod) GetID() BuilderMethodID {
	return m.ID
}

func (EnableBuilderMethod) isBuilderMethod() {}

type DeepBuilderMethod struct {
	ID             BuilderMethodID
	EnableMethodID BuilderMethodID
	Relation       RelationID
	ChildBuilder   BuilderID
}

func (m DeepBuilderMethod) GetID() BuilderMethodID {
	return m.ID
}

func (DeepBuilderMethod) isBuilderMethod() {}

type VariantBuilderMethod struct {
	ID           BuilderMethodID
	Variant      semantic.VariantID
	ChildBuilder BuilderID
}

func (m VariantBuilderMethod) GetID() BuilderMethodID {
	return m.ID
}

func (VariantBuilderMethod) isBuilderMethod() {}

type NestedBuilderMethod struct {
	ID           BuilderMethodID
	Nested       NestedID
	ChildBuilder BuilderID
}

func (m NestedBuilderMethod) GetID() BuilderMethodID {
	return m.ID
}

func (NestedBuilderMethod) isBuilderMethod() {}
