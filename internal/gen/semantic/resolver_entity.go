package semantic

import (
	"fmt"
	"go/types"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

type entityEntry struct {
	ID     EntityID
	Schema schema.Entity
	Type   types.Type
}

func (r *Resolver) initEntities() {
	for _, e := range r.schema {
		entityName := e.Name.V

		if _, ok := r.seenEntities[entityName]; ok {
			r.eerr(entityName, "entity with this name was already defined", e.Definition)
			continue
		}

		tid, t := r.addEntityType(e)

		eid := EntityID(len(r.model.Entities))
		r.seenEntities[e.Name.V] = entityEntry{
			ID:     eid,
			Schema: e,
			Type:   t,
		}
		r.model.Entities = append(r.model.Entities, Entity{
			Name: entityName,
			Type: tid,
		})

		if t != nil {
			r.model.Entities[eid].IDMember = r.findOrAddEntityIDMember(e, t)
		}
	}
}

func (r *Resolver) addEntityType(e schema.Entity) (TypeID, types.Type) {
	entityName := e.Name.V

	t, ok := r.types.goTypes[e.Type.Base.V]
	if !ok {
		r.eerr(entityName, "entity type base must be existing non-scalar go type", e.Type.Definition)
		return 0, nil
	}

	if t.IsBasic {
		r.eerr(entityName, "entity type base cannot be scalar type", e.Type.Definition)
		return 0, nil
	}

	if !t.TypeName.Exported() {
		r.eerr(entityName, "entity type base must be exported ("+t.TypeName.Type().String()+" is not)", e.Type.Definition)
		return 0, nil
	}

	pt, rt := r.composeType(entityName, e.Type)
	if pt == nil {
		return 0, nil
	}

	tid := r.findOrCreateTypeRecursively(rt)

	return tid, pt
}

func (r *Resolver) findOrAddEntityIDMember(e schema.Entity, t types.Type) MemberID {
	entityName := e.Name.V

	mid, mt := r.findOrCreateMember(entityName, t, e.ID)
	if mt == nil {
		return 0
	}

	if !types.Comparable(mt) {
		r.eerr(entityName, fmt.Sprintf("type %s is not suitable for ID as it's incomparable", mt.String()), e.Type.Definition)

		return 0
	}

	return mid
}
