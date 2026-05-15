package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initRelations() {
	p.referencedEntities = make(map[semantic.EntityID]int, len(p.model.Entities))
	p.reverseReferencedEntities = make(map[semantic.EntityID]int, len(p.model.Entities))
	p.referencingEntities = make(map[semantic.EntityID]int, len(p.model.Entities))
	p.referencingVariants = make(map[semantic.VariantID]int, len(p.model.Variants))
	p.relationsByEntity = make(map[semantic.EntityID][]RelationID, len(p.model.Relations))

	p.plan.Relations = make([]Relation, len(p.model.Relations))

	for id, r := range p.model.Relations {
		rid := RelationID(id)

		p.referencedEntities[r.To]++
		p.referencingEntities[r.From]++
		if r.Variant.Set {
			p.referencingVariants[r.Variant.V]++
		}

		if r.ReversedBy.Set {
			p.reverseReferencedEntities[r.To]++
		}

		p.plan.Relations[rid] = Relation{ID: rid, Relation: r}

		p.relationsByEntity[r.From] = append(p.relationsByEntity[r.From], rid)
	}
}
