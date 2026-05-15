package semantic

import (
	"fmt"
	"go/types"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

type relationKey struct {
	EntityID EntityID
	Variant  Optional[VariantID]
	Name     string
}

func (r *Resolver) initRelations() {
	for _, e := range r.schema {
		entityName := e.Name.V
		entityID := r.seenEntities[entityName].ID

		for _, rel := range e.Relations {
			relationName := rel.Name.V
			key := relationKey{
				EntityID: entityID,
				Name:     rel.Name.V,
			}

			if _, ok := r.seenRelations[key]; ok {
				r.eerr(entityName, fmt.Sprintf("relation `%s` of the entity was already defined", relationName), rel.Definition)
				continue
			}

			relationEntity := r.seenEntities[rel.Entity.V]

			var reversedBy Optional[MemberID]

			if rel.ReversedBy.V != "" {
				reversedBy.V = r.findOrAddReversedByMember(e, relationEntity.Schema, rel.ReversedBy)
				reversedBy.Set = true
			}

			r.seenRelations[key] = struct{}{}
			r.model.Relations = append(r.model.Relations, Relation{
				Name:       relationName,
				From:       entityID,
				To:         relationEntity.ID,
				ReversedBy: reversedBy,
			})
		}

		for _, v := range e.Variants {
			variantName := r.resolveVariantName(entityName, v)
			variant := r.seenVariants[variantKey{EntityID: entityID, Name: variantName}]

			for _, rel := range v.Relations {
				relationName := rel.Name.V
				key := relationKey{
					EntityID: entityID,
					Variant:  Optional[VariantID]{V: variant.ID, Set: true},
					Name:     rel.Name.V,
				}

				if _, ok := r.seenRelations[key]; ok {
					r.eerr(entityName, fmt.Sprintf("relation `%s` of the entity's variant `%s` was already defined", relationName, variantName), rel.Definition)
					continue
				}

				relationEntity := r.seenEntities[rel.Entity.V]

				if rel.ReversedBy.V != "" {
					r.eerr(entityName, fmt.Sprintf("relation `%s` of the entity's variant `%s` cannot be reversed. consider moving reversed relation to the root entity", relationName, variantName), rel.ReversedBy.Definition)
				}

				r.seenRelations[key] = struct{}{}
				r.model.Relations = append(r.model.Relations, Relation{
					Name:    relationName,
					From:    entityID,
					Variant: Optional[VariantID]{V: variant.ID, Set: true},
					To:      relationEntity.ID,
				})
			}
		}
	}
}

func (r *Resolver) findOrAddReversedByMember(from, ref schema.Entity, reversedByMember schema.Value[string]) MemberID {
	fromEntityName := from.Name.V
	fromEntity := r.seenEntities[fromEntityName]

	refEntityName := ref.Name.V
	refEntity := r.seenEntities[refEntityName]

	_, ft := r.findOrCreateMember(fromEntityName, fromEntity.Type, from.ID)
	if ft == nil {
		panic("failed to find referencing entity ID member. at this point it is a bug")
	}

	mid, rt := r.findOrCreateMember(refEntityName, refEntity.Type, reversedByMember)
	if rt == nil {
		return 0
	}

	if !types.Identical(ft, rt) {
		r.eerr(
			fromEntityName,
			fmt.Sprintf(
				"the entity cannot be resolved by %s's member `%s` because its type (%s) isn't same as ID member's type (%s) of the entity",
				refEntity.Type.String(),
				reversedByMember.V,
				rt.String(),
				ft.String(),
			),
			reversedByMember.Definition,
		)
		return 0
	}

	return mid
}
