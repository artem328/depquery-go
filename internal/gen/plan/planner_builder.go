package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initBuilders() {
	type (
		rootBuilderEntry struct {
			ID      BuilderID
			Builder RootBuilder
		}
		variantBuilderEntry struct {
			ID      BuilderID
			Builder VariantBuilder
		}
	)

	builderByEntity := make(map[semantic.EntityID]*rootBuilderEntry)
	builderByVariant := make(map[semantic.VariantID]*variantBuilderEntry)

	var (
		builderID       counter[BuilderID]
		builderMethodID counter[BuilderMethodID]
	)

	for id := range p.model.Entities {
		eid := semantic.EntityID(id)
		if p.referencingEntities[eid] == 0 && len(p.nestedsByEntity[eid]) == 0 {
			continue
		}

		bid := builderID.Next()

		builderByEntity[eid] = &rootBuilderEntry{
			ID: bid,
			Builder: RootBuilder{
				Entity:            eid,
				FetchContextRoot:  p.fetchContextRootByEntity[eid],
				EntityResolver:    p.entityResolverByEntity[eid],
				NestedResolver:    p.nestedResolverByEntity[eid],
				IsRelationBuilder: p.referencingEntities[eid] > 0,
				IsNestedBuilder:   len(p.nestedsByEntity[eid]) > 0,
				CommonBuilder:     CommonBuilder{ID: bid},
			},
		}
	}

	for id, v := range p.model.Variants {
		vid := semantic.VariantID(id)
		if p.referencingVariants[vid] == 0 {
			continue
		}

		bid := builderID.Next()

		parent := builderByEntity[v.Entity]
		builderByVariant[vid] = &variantBuilderEntry{
			ID:      bid,
			Builder: VariantBuilder{Variant: vid, Parent: parent.ID, CommonBuilder: CommonBuilder{ID: bid}},
		}
		bmid := builderMethodID.Next()
		parent.Builder.Methods = append(parent.Builder.Methods, VariantBuilderMethod{
			ID:           bmid,
			Variant:      vid,
			ChildBuilder: bid,
		})
		parent.Builder.ChildBuilders = append(parent.Builder.ChildBuilders, RegularChildBuilder{
			Builder: bid,
		})
	}

	for id, r := range p.model.Relations {
		rid := RelationID(id)

		if r.Variant.Set {
			b, ok := builderByVariant[r.Variant.V]
			if !ok {
				continue
			}

			enableMethodID := builderMethodID.Next()

			b.Builder.Methods = append(b.Builder.Methods, EnableBuilderMethod{
				ID:       enableMethodID,
				Relation: rid,
			})

			if cb, ok := builderByEntity[r.To]; ok {
				b.Builder.Methods = append(b.Builder.Methods, DeepBuilderMethod{
					ID:             builderMethodID.Next(),
					EnableMethodID: enableMethodID,
					Relation:       rid,
					ChildBuilder:   cb.ID,
				})
				b.Builder.ChildBuilders = append(b.Builder.ChildBuilders, RelationChildBuilder{
					Builder:  cb.ID,
					Relation: rid,
				})
			}

			continue
		}

		b := builderByEntity[r.From]
		enableMethodID := builderMethodID.Next()

		b.Builder.Methods = append(b.Builder.Methods, EnableBuilderMethod{
			ID:       enableMethodID,
			Relation: rid,
		})

		if cb, ok := builderByEntity[r.To]; ok {
			b.Builder.Methods = append(b.Builder.Methods, DeepBuilderMethod{
				ID:             builderMethodID.Next(),
				EnableMethodID: enableMethodID,
				Relation:       rid,
				ChildBuilder:   cb.ID,
			})
			b.Builder.ChildBuilders = append(b.Builder.ChildBuilders, RelationChildBuilder{
				Builder:  cb.ID,
				Relation: rid,
			})
		}
	}

	for _, n := range p.plan.Nesteds {
		b := builderByEntity[n.From]

		if cb, ok := builderByEntity[n.To]; ok {
			b.Builder.Methods = append(b.Builder.Methods, NestedBuilderMethod{
				ID:           builderMethodID.Next(),
				Nested:       n.ID,
				ChildBuilder: cb.ID,
			})
			b.Builder.ChildBuilders = append(b.Builder.ChildBuilders, NestedChildBuilder{
				Builder: cb.ID,
				Nested:  n.ID,
			})
		}
	}

	p.plan.Builders = make([]Builder, builderID.Current())

	for _, b := range builderByEntity {
		p.plan.Builders[b.ID] = b.Builder
	}
	for _, b := range builderByVariant {
		p.plan.Builders[b.ID] = b.Builder
	}
}
