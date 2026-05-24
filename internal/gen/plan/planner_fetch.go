package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initFetch() {
	p.initSyntheticStateContainers()
	p.initFetchParents()
	p.initNestedEntityFetch()
}

func (p *Planner) initSyntheticStateContainers() {
	p.syntheticContainerByEntity = make(map[semantic.EntityID]SyntheticStateContainerID)
	p.plan.SyntheticStateContainers = make([]SyntheticStateContainer, 0, len(p.model.Entities))

	var syntheticStateContainerID counter[SyntheticStateContainerID]

	for id, e := range p.model.Entities {
		if !e.Synthetic {
			continue
		}

		eid := semantic.EntityID(id)

		if p.referencedEntities[eid] == 0 && p.referencingEntities[eid] == 0 && len(p.nestedsByEntity[eid]) == 0 {
			// TODO: warning
			continue
		}

		sscid := syntheticStateContainerID.Next()
		p.plan.SyntheticStateContainers = append(p.plan.SyntheticStateContainers, SyntheticStateContainer{
			ID:          sscid,
			Entity:      eid,
			IDNamespace: p.idNamespace.Next(),
		})
		p.syntheticContainerByEntity[eid] = sscid
	}
}

func (p *Planner) initFetchParents() {
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

	for id, e := range p.model.Entities {
		eid := semantic.EntityID(id)

		isParent := p.referencingEntities[eid] > 0 || len(p.nestedsByEntity[eid]) > 0
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
				ID:                      fcrid,
				Entity:                  eid,
				Synthetic:               e.Synthetic,
				FetchParent:             parentID,
				StateContainer:          p.containerByEntity[eid],
				SyntheticStateContainer: p.syntheticContainerByEntity[eid],
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

	p.parentFetchGetterByFetchParent = make(map[FetchParentID]ParentFetchGetterID)
	p.plan.ReversedFetchParents = make([]ReversedFetchParent, 0, len(p.plan.FetchParents))
	p.plan.ParentFetchGetters = make([]ParentFetchGetter, 0, len(p.plan.FetchParents))

	var parentFetchGetterID counter[ParentFetchGetterID]

	for _, fp := range p.plan.FetchParents {
		if fp.Reversed {
			continue
		}

		e := p.model.Entities[fp.Entity]
		pfgid := parentFetchGetterID.Next()

		p.plan.ParentFetchGetters = append(p.plan.ParentFetchGetters, ParentFetchGetter{
			ID:                      pfgid,
			FetchParent:             fp.ID,
			Synthetic:               e.Synthetic,
			StateContainer:          p.containerByEntity[fp.Entity],
			SyntheticStateContainer: p.syntheticContainerByEntity[fp.Entity],
		})
		p.parentFetchGetterByFetchParent[fp.ID] = pfgid

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

func (p *Planner) initNestedEntityFetch() {
	p.nestedEntityFetchByNested = make(map[NestedID]NestedEntityFetchID)
	nestedEntityFetchByEntity := make(map[semantic.EntityID]NestedEntityFetchID)

	var nestedEntityFetchID counter[NestedEntityFetchID]

	for _, n := range p.plan.Nesteds {
		e := p.model.Entities[n.To]

		if p.referencingEntities[n.To] == 0 && len(p.nestedsByEntity[n.To]) == 0 {
			continue
		}

		if nefid, ok := nestedEntityFetchByEntity[n.To]; ok {
			p.nestedEntityFetchByNested[n.ID] = nefid
			continue
		}

		nefid := nestedEntityFetchID.Next()

		p.plan.NestedEntityFetches = append(p.plan.NestedEntityFetches, NestedEntityFetch{
			ID:                        nefid,
			Entity:                    n.To,
			Parent:                    p.fetchParentByEntity[n.To],
			Synthetic:                 e.Synthetic,
			StateContainer:            p.containerByEntity[n.To],
			SyntheticStateContainerID: p.syntheticContainerByEntity[n.To],
		})
		p.nestedEntityFetchByNested[n.ID] = nefid
		nestedEntityFetchByEntity[n.To] = nefid
	}
}
