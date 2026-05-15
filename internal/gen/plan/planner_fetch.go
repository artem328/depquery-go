package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initFetch() {
	p.fetchByEntity = make(map[semantic.EntityID]EntityFetchID)
	p.reversedFetchByRelation = make(map[RelationID]EntityFetchID)
	p.fetchParentByEntity = make(map[semantic.EntityID]FetchParentID)
	p.fetchContextRootByEntity = make(map[semantic.EntityID]FetchContextRootID)

	type reversedFetchParent struct {
		ID               FetchParentID
		StateContainerID ReversedStateContainerID
	}

	reversedFetchParents := make(map[semantic.EntityID][]reversedFetchParent)

	var (
		fetchContextRootID counter[FetchContextRootID]
		entityFetchID      counter[EntityFetchID]
		fetchParentID      counter[FetchParentID]
		fetchChildID       counter[FetchChildID]
	)

	for id := range p.model.Entities {
		eid := semantic.EntityID(id)

		isParent := p.referencingEntities[eid] > 0
		isChild := p.referencedEntities[eid] > p.reverseReferencedEntities[eid]

		if !isParent && !isChild {
			continue
		}

		var (
			parentID FetchParentID
			childID  FetchChildID
		)

		if isParent {
			parentID = fetchParentID.Next()
			p.plan.FetchParents = append(p.plan.FetchParents, FetchParent{
				ID:     parentID,
				Entity: eid,
			})
			p.fetchParentByEntity[eid] = parentID

			fcrid := fetchContextRootID.Next()
			p.plan.FetchContextRoots = append(p.plan.FetchContextRoots, FetchContextRoot{
				ID:             fcrid,
				Entity:         eid,
				FetchParent:    parentID,
				StateContainer: p.containerByEntity[eid],
			})
			p.fetchContextRootByEntity[eid] = fcrid
		}

		if isChild {
			childID = fetchChildID.Next()
			p.plan.FetchChildren = append(p.plan.FetchChildren, FetchChild{
				ID:     childID,
				Entity: eid,
			})
		}

		if !isChild {
			continue
		}

		efid := entityFetchID.Next()
		p.plan.EntityFetches = append(p.plan.EntityFetches, EntityFetch{
			ID:             efid,
			Entity:         eid,
			StateContainer: p.containerByEntity[eid],
			PrefetchMethod: p.prefetchMethodByEntity[eid],
			Parent:         parentID,
			Child:          childID,
			IsParent:       isParent,
			Reversed:       false,
		})
		p.fetchByEntity[eid] = efid
	}

	for _, rel := range p.plan.Relations {
		if !rel.ReversedBy.Set {
			continue
		}

		isParent := p.referencingEntities[rel.To] > 0
		isChild := p.referencedEntities[rel.To] > 0

		if !isParent && !isChild {
			continue
		}

		var (
			parentID FetchParentID
			childID  FetchChildID
		)

		if isParent {
			parentID = fetchParentID.Next()
			p.plan.FetchParents = append(p.plan.FetchParents, FetchParent{
				ID:               parentID,
				Entity:           rel.To,
				ReversedByEntity: rel.From,
				Reversed:         true,
			})
			reversedFetchParents[rel.To] = append(reversedFetchParents[rel.To], reversedFetchParent{
				ID:               parentID,
				StateContainerID: p.reversedContainersByRelation[rel.ID],
			})
		}

		if isChild {
			childID = fetchChildID.Next()
			p.plan.FetchChildren = append(p.plan.FetchChildren, FetchChild{
				ID:               childID,
				Entity:           rel.To,
				ReversedByEntity: rel.From,
				Reversed:         true,
			})
		}

		if !isChild {
			continue
		}

		efid := entityFetchID.Next()
		p.plan.EntityFetches = append(p.plan.EntityFetches, EntityFetch{
			ID:                     efid,
			Entity:                 rel.To,
			StateContainer:         p.containerByEntity[rel.To],
			ReversedStateContainer: p.reversedContainersByRelation[rel.ID],
			PrefetchMethod:         p.reversedPrefetchMethodByRelation[rel.ID],
			Parent:                 parentID,
			Child:                  childID,
			IsParent:               isParent,
			Reversed:               true,
		})
		p.reversedFetchByRelation[rel.ID] = efid
	}

	p.plan.ReversedFetchParents = make([]ReversedFetchParent, 0, len(p.plan.FetchParents))
	p.plan.ParentFetchGetters = make([]ParentFetchGetter, 0, len(p.plan.FetchParents))

	for _, fp := range p.plan.FetchParents {
		if fp.Reversed {
			continue
		}

		p.plan.ParentFetchGetters = append(p.plan.ParentFetchGetters, ParentFetchGetter{
			FetchParent:    fp.ID,
			StateContainer: p.containerByEntity[fp.Entity],
		})

		reversedBy := reversedFetchParents[fp.Entity]
		if len(reversedBy) == 0 {
			continue
		}

		reversed := make([]FetchParentReverse, 0, len(reversedBy))
		for _, r := range reversedBy {
			reversed = append(reversed, FetchParentReverse{
				FetchParent:    r.ID,
				StateContainer: r.StateContainerID,
			})
		}

		p.plan.ReversedFetchParents = append(p.plan.ReversedFetchParents, ReversedFetchParent{
			FetchParent:          fp.ID,
			StateContainer:       p.containerByEntity[fp.Entity],
			ReversedFetchParents: reversed,
		})
	}
}
