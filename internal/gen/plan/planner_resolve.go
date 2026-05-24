package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initResolve() {
	p.initEntityResolvers()
	p.initNestedResolvers()
}

func (p *Planner) initEntityResolvers() {
	p.entityResolverByEntity = make(map[semantic.EntityID]EntityResolverID, len(p.model.Entities))

	resolveMethodByRelation := make(map[RelationID]ResolveMethodID, len(p.plan.Relations))

	var resolveMethodID counter[ResolveMethodID]

	for _, rel := range p.plan.Relations {
		if rel.ReversedBy.Set {
			continue
		}

		rmid := resolveMethodID.Next()
		p.plan.ResolveMethods = append(p.plan.ResolveMethods, RelationResolveMethod{
			ID:       rmid,
			Relation: rel.ID,
		})
		resolveMethodByRelation[rel.ID] = rmid
	}

	p.plan.EntityResolvers = make([]EntityResolver, 0, len(p.relationsByEntity))

	var entityResolverID counter[EntityResolverID]

	for id := range p.model.Entities {
		eid := semantic.EntityID(id)
		rids := p.relationsByEntity[eid]
		if len(rids) == 0 {
			continue
		}

		resolutions := make([]EntityResolution, 0, len(rids))
		variantResolutions := make(map[semantic.VariantID][]EntityResolution)
		variants := make([]semantic.VariantID, 0)
		for _, rid := range rids {
			rel := p.plan.Relations[rid]

			efid := p.fetchByEntity[rel.To]
			if rel.ReversedBy.Set {
				efid = p.reversedFetchByRelation[rel.ID]
			}

			if rel.Variant.Set {
				if _, ok := variantResolutions[rel.Variant.V]; !ok {
					variants = append(variants, rel.Variant.V)
				}

				variantResolutions[rel.Variant.V] = append(variantResolutions[rel.Variant.V], EntityResolutionRelation{
					Relation:      rid,
					ResolveMethod: resolveMethodByRelation[rid],
					EntityFetch:   efid,
				})
				continue
			}

			resolutions = append(resolutions, EntityResolutionRelation{
				Relation:      rid,
				ResolveMethod: resolveMethodByRelation[rid],
				EntityFetch:   efid,
			})
		}

		for _, vid := range variants {
			resolutions = append(resolutions, EntityResolutionVariant{
				Variant:     vid,
				Resolutions: variantResolutions[vid],
			})
		}

		erid := entityResolverID.Next()
		p.plan.EntityResolvers = append(p.plan.EntityResolvers, EntityResolver{
			ID:                erid,
			Entity:            eid,
			ParentFetchGetter: p.parentFetchGetterByFetchParent[p.fetchParentByEntity[eid]],
			Resolutions:       resolutions,
		})
		p.entityResolverByEntity[eid] = erid
	}
}

func (p *Planner) initNestedResolvers() {
	p.plan.NestedResolvers = make([]NestedResolver, 0, len(p.plan.Nesteds))
	p.nestedResolverByEntity = make(map[semantic.EntityID]NestedResolverID, len(p.model.Entities))
	resolveMethodByNested := make(map[NestedID]ResolveMethodID, len(p.plan.Nesteds))

	var (
		resolveMethodID  counter[ResolveMethodID]
		nestedResolverID counter[NestedResolverID]
	)

	resolveMethodID.val = ResolveMethodID(len(p.plan.ResolveMethods))

	for _, n := range p.plan.Nesteds {
		nrmid := resolveMethodID.Next()

		p.plan.ResolveMethods = append(p.plan.ResolveMethods, NestedResolveMethod{
			ID:     nrmid,
			Nested: n.ID,
		})
		resolveMethodByNested[n.ID] = nrmid
	}

	for id := range p.model.Entities {
		eid := semantic.EntityID(id)
		nids := p.nestedsByEntity[eid]
		if len(nids) == 0 {
			continue
		}

		resolutions := make([]NestedResolution, 0, len(nids))

		for _, nid := range nids {
			n := p.plan.Nesteds[nid]
			ne := p.model.Entities[n.To]

			resolutions = append(resolutions, NestedResolution{
				Nested:            n.ID,
				Synthetic:         ne.Synthetic,
				ResolveMethod:     resolveMethodByNested[n.ID],
				NestedEntityFetch: p.nestedEntityFetchByNested[n.ID],
			})
		}

		nrid := nestedResolverID.Next()
		p.plan.NestedResolvers = append(p.plan.NestedResolvers, NestedResolver{
			ID:                   nrid,
			Entity:               eid,
			ParentFetchGetter:    p.parentFetchGetterByFetchParent[p.fetchParentByEntity[eid]],
			Resolutions:          resolutions,
			SyntheticIDNamespace: p.idNamespace.Next(),
		})
		p.nestedResolverByEntity[eid] = nrid
	}
}
