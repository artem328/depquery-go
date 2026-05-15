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
	RegularChildBuilder
	Relation RelationID
}

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
	Entity           semantic.EntityID
	FetchContextRoot FetchContextRootID
	EntityResolver   EntityResolverID
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

type CommonBuilderMethod struct {
	ID BuilderMethodID
}

func (m CommonBuilderMethod) GetID() BuilderMethodID {
	return m.ID
}

type EnableBuilderMethod struct {
	CommonBuilderMethod
	Relation RelationID
}

func (EnableBuilderMethod) isBuilderMethod() {}

type DeepBuilderMethod struct {
	CommonBuilderMethod
	EnableMethodID BuilderMethodID
	Relation       RelationID
	ChildBuilder   BuilderID
}

func (DeepBuilderMethod) isBuilderMethod() {}

type VariantBuilderMethod struct {
	CommonBuilderMethod
	Variant      semantic.VariantID
	ChildBuilder BuilderID
}

func (VariantBuilderMethod) isBuilderMethod() {}
