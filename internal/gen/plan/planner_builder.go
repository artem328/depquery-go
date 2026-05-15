package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

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

func (p *Planner) initBuilders() {
	builderByEntity := make(map[semantic.EntityID]*rootBuilderEntry)
	builderByVariant := make(map[semantic.VariantID]*variantBuilderEntry)

	var (
		builderID       counter[BuilderID]
		builderMethodID counter[BuilderMethodID]
	)

	for id := range p.model.Entities {
		eid := semantic.EntityID(id)
		if p.referencingEntities[eid] == 0 {
			continue
		}

		bid := builderID.Next()

		builderByEntity[eid] = &rootBuilderEntry{
			ID: bid,
			Builder: RootBuilder{
				Entity:           eid,
				FetchContextRoot: p.fetchContextRootByEntity[eid],
				EntityResolver:   p.entityResolverByEntity[eid],
				CommonBuilder:    CommonBuilder{ID: bid},
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
			CommonBuilderMethod: CommonBuilderMethod{ID: bmid},
			Variant:             vid,
			ChildBuilder:        bid,
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
				CommonBuilderMethod: CommonBuilderMethod{ID: enableMethodID},
				Relation:            rid,
			})

			if cb, ok := builderByEntity[r.To]; ok {
				b.Builder.Methods = append(b.Builder.Methods, DeepBuilderMethod{
					CommonBuilderMethod: CommonBuilderMethod{ID: builderMethodID.Next()},
					EnableMethodID:      enableMethodID,
					Relation:            rid,
					ChildBuilder:        cb.ID,
				})
				b.Builder.ChildBuilders = append(b.Builder.ChildBuilders, RelationChildBuilder{
					Relation:            rid,
					RegularChildBuilder: RegularChildBuilder{Builder: cb.ID},
				})
			}

			continue
		}

		b := builderByEntity[r.From]
		enableMethodID := builderMethodID.Next()

		b.Builder.Methods = append(b.Builder.Methods, EnableBuilderMethod{
			CommonBuilderMethod: CommonBuilderMethod{ID: enableMethodID},
			Relation:            rid,
		})

		if cb, ok := builderByEntity[r.To]; ok {
			b.Builder.Methods = append(b.Builder.Methods, DeepBuilderMethod{
				CommonBuilderMethod: CommonBuilderMethod{ID: builderMethodID.Next()},
				EnableMethodID:      enableMethodID,
				Relation:            rid,
				ChildBuilder:        cb.ID,
			})
			b.Builder.ChildBuilders = append(b.Builder.ChildBuilders, RelationChildBuilder{
				Relation:            rid,
				RegularChildBuilder: RegularChildBuilder{Builder: cb.ID},
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
