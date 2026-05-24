package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initState() {
	var (
		stateContainerID         counter[StateContainerID]
		reversedStateContainerID counter[ReversedStateContainerID]
	)

	p.containerByEntity = make(map[semantic.EntityID]StateContainerID, len(p.model.Entities))
	p.reversedContainersByRelation = make(map[RelationID]ReversedStateContainerID, len(p.model.Relations))

	p.plan.StateContainers = make([]StateContainer, 0, len(p.model.Entities))
	p.plan.ReversedStateContainers = make([]ReversedStateContainer, 0, len(p.model.Relations))

	for id, e := range p.model.Entities {
		eid := semantic.EntityID(id)

		if (p.referencedEntities[eid] == 0 && p.referencingEntities[eid] == 0 && len(p.nestedsByEntity[eid]) == 0) || e.Synthetic {
			// TODO: warning
			continue
		}

		scid := stateContainerID.Next()

		p.plan.StateContainers = append(p.plan.StateContainers, StateContainer{
			ID:     scid,
			Entity: eid,
		})
		p.containerByEntity[eid] = scid
	}

	for _, r := range p.plan.Relations {
		if !r.ReversedBy.Set {
			continue
		}
		rscid := reversedStateContainerID.Next()
		scid := p.containerByEntity[r.To]

		p.plan.ReversedStateContainers = append(p.plan.ReversedStateContainers, ReversedStateContainer{
			ID:                 rscid,
			StateContainerID:   scid,
			HolderEntity:       r.From,
			HolderEntityMember: r.ReversedBy.V,
		})
		p.plan.StateContainers[scid].ReversedBy = append(p.plan.StateContainers[scid].ReversedBy, rscid)
		p.reversedContainersByRelation[r.ID] = rscid
	}
}
