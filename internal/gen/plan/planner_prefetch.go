package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initPrefetch() {
	p.plan.PrefetchMethods = make([]PrefetchMethod, 0, len(p.model.Entities)+len(p.model.Relations))
	p.prefetchMethodByEntity = make(map[semantic.EntityID]PrefetchMethodID, len(p.model.Entities))
	p.reversedPrefetchMethodByRelation = make(map[RelationID]PrefetchMethodID, len(p.model.Relations))

	var prefetchMethodID counter[PrefetchMethodID]

	for id := range p.model.Entities {
		eid := semantic.EntityID(id)
		if p.referencedEntities[eid] == 0 || p.referencedEntities[eid] == p.reverseReferencedEntities[eid] {
			continue
		}

		pfid := prefetchMethodID.Next()
		p.plan.PrefetchMethods = append(p.plan.PrefetchMethods, PrefetchMethod{
			ID:     pfid,
			Entity: eid,
		})
		p.prefetchMethodByEntity[eid] = pfid
	}

	for _, r := range p.plan.Relations {
		if !r.ReversedBy.Set {
			continue
		}

		pfid := prefetchMethodID.Next()
		p.plan.PrefetchMethods = append(p.plan.PrefetchMethods, PrefetchMethod{
			ID:               pfid,
			Entity:           r.To,
			ReversedByEntity: r.From,
			Reversed:         true,
		})
		p.reversedPrefetchMethodByRelation[r.ID] = pfid
	}
}
