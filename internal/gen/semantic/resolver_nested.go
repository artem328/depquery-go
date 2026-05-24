package semantic

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (r *Resolver) initNested() {
	type nestedKey struct {
		Entity EntityID
		Name   string
	}

	seenNested := make(map[nestedKey]struct{})

	for _, e := range r.schema {
		if len(e.Nested) == 0 {
			continue
		}

		entityName := e.Name.V
		eid := r.seenEntities[entityName].ID

		for _, n := range e.Nested {
			nestedName := r.resolveNestedName(entityName, n)
			key := nestedKey{
				Entity: eid,
				Name:   nestedName,
			}

			if _, ok := seenNested[key]; ok {
				r.eerr(entityName, fmt.Sprintf("nested field `%s` for the entity was already defined", nestedName), n.Definition)
				continue
			}
			seenNested[key] = struct{}{}

			nestedEntity, ok := r.seenEntities[n.Entity.V]
			if !ok {
				r.eerr(entityName, fmt.Sprintf("nested field `%s` references unexisting entity `%s`", nestedName, n.Entity.V), n.Entity.Definition)
				continue
			}

			r.model.Nesteds = append(r.model.Nesteds, Nested{
				Name: nestedName,
				From: eid,
				To:   nestedEntity.ID,
			})
		}
	}
}

func (r *Resolver) resolveNestedName(entityName string, n schema.Nested) string {
	name := n.Name.V
	if name != "" {
		return name
	}

	e := r.seenEntities[n.Entity.V].Schema

	vt, ok := r.types.goTypes[e.Type.Base.V]
	if !ok {
		r.eerr(entityName, "entity nested field type base must be existing non-scalar go type", n.Entity.Definition)
		return ""
	}

	return vt.TypeName.Name()
}
