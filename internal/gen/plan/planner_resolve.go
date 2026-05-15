package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initResolve() {
	p.entityResolverByEntity = make(map[semantic.EntityID]EntityResolverID, len(p.model.Entities))
	p.plan.ResolveMethods = make([]ResolveMethod, 0, len(p.plan.Relations))

	resolveMethodByRelation := make(map[RelationID]ResolveMethodID, len(p.plan.Relations))

	var resolveMethodID counter[ResolveMethodID]

	for _, rel := range p.plan.Relations {
		if rel.ReversedBy.Set {
			continue
		}

		rmid := resolveMethodID.Next()
		p.plan.ResolveMethods = append(p.plan.ResolveMethods, ResolveMethod{
			ID:       rmid,
			Relation: rel.ID,
		})
		resolveMethodByRelation[rel.ID] = rmid
	}

	p.plan.EntityResolvers = make([]EntityResolver, 0, len(p.relationsByEntity))

	var entityRelationsID counter[EntityResolverID]

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

		erid := entityRelationsID.Next()
		p.plan.EntityResolvers = append(p.plan.EntityResolvers, EntityResolver{
			ID:          erid,
			Entity:      eid,
			FetchParent: p.fetchParentByEntity[eid],
			Resolutions: resolutions,
		})
		p.entityResolverByEntity[eid] = erid
	}
}
