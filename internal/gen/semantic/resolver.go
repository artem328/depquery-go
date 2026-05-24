package semantic

import (
	"github.com/artem328/depquery-go/internal/gen/schema"
)

type Resolver struct {
	schema schema.Schema
	model  Model
	errors []error

	types         typeIndex
	seenEntities  map[string]entityEntry
	seenVariants  map[variantKey]variantEntry
	seenMembers   map[memberKey]memberEntry
	seenRelations map[relationKey]struct{}
	seenTypes     map[[16]byte]TypeID
}

func NewResolver(schema schema.Schema) *Resolver {
	return &Resolver{
		schema:        schema,
		types:         newTypeIndex(),
		seenEntities:  make(map[string]entityEntry, len(schema)),
		seenVariants:  make(map[variantKey]variantEntry, len(schema)*16),
		seenMembers:   make(map[memberKey]memberEntry, len(schema)*16),
		seenRelations: make(map[relationKey]struct{}, len(schema)*16),
		seenTypes:     make(map[[16]byte]TypeID, len(schema)*16),
	}
}

func (r *Resolver) Build() (Model, []error) {
	r.initTypeIndex()
	if len(r.errors) > 0 {
		return Model{}, r.errors
	}

	r.initEntities()
	if len(r.errors) > 0 {
		return Model{}, r.errors
	}

	r.initVariants()
	if len(r.errors) > 0 {
		return Model{}, r.errors
	}

	r.initNested()
	if len(r.errors) > 0 {
		return Model{}, r.errors
	}

	r.initRelations()
	if len(r.errors) > 0 {
		return Model{}, r.errors
	}

	return r.model, nil
}
