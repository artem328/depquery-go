package parser

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (p *Parser) buildSchema() {
	p.schema = make([]*schema.Entity, 0, len(p.entities))
	schemaEntitiesByName := make(map[string]*schema.Entity)

	for _, e := range p.entities {
		se := p.buildEntity(e)
		schemaEntitiesByName[e.Name] = se

		p.schema = append(p.schema, se)
	}

	for _, e := range p.entities {
		schemaEntitiesByName[e.Name].Relations = p.buildRelations(e.Name, e.Relations, schemaEntitiesByName)
		schemaEntitiesByName[e.Name].Variants = p.buildVariants(e, schemaEntitiesByName)
	}
}

func (p *Parser) buildEntity(e *entity) *schema.Entity {
	return &schema.Entity{
		Name: e.Name,
		Type: p.buildType(e.Type),
		ID:   p.buildIDMember(e.ID),
	}
}

func (p *Parser) buildIDMember(id id) schema.IDMember {
	return schema.IDMember{
		Name:    id.Member,
		RcvType: id.MemberType,
		Type:    p.buildType(id.Type),
	}
}

func (p *Parser) buildType(t *goType) schema.Type {
	return schema.Type{
		Package: t.Package,
		Name:    t.Name,
		Wrapper: t.Wrapper,
		Params:  p.buildTypeList(t.Params),
	}
}

func (p *Parser) buildTypeList(tt []*goType) []schema.Type {
	if len(tt) == 0 {
		return nil
	}

	types := make([]schema.Type, len(tt))
	for i, t := range tt {
		types[i] = p.buildType(t)
	}

	return types
}

func (p *Parser) buildRelations(entityName string, rr []relation, idx map[string]*schema.Entity) []schema.Relation {
	if rr == nil {
		return nil
	}

	relations := make([]schema.Relation, 0, len(rr))

	for _, rel := range rr {
		ent, ok := idx[rel.Entity]
		if !ok {
			p.gerr(
				entityName,
				fmt.Sprintf("relation %s cannot be resolved. no entity %s found", rel.Name, rel.Entity),
				rel.source,
			)

			continue
		}

		var rev *schema.IDMember

		if rel.Reversed != nil {
			r := p.buildIDMember(*rel.Reversed)
			rev = &r
		}

		relations = append(relations, schema.Relation{
			Name:       rel.Name,
			Entity:     ent,
			ReversedBy: rev,
		})
	}

	return relations
}

func (p *Parser) buildVariants(e *entity, idx map[string]*schema.Entity) []schema.EntityVariant {
	if e.Variants == nil {
		return nil
	}

	variants := make([]schema.EntityVariant, 0, len(e.Variants))

	for _, v := range e.Variants {
		name := v.Name
		if name == "" {
			name = v.Type.Name
		}

		variants = append(variants, schema.EntityVariant{
			Name:      name,
			Type:      p.buildType(v.Type),
			Relations: p.buildRelations(e.Name, v.Relations, idx),
		})
	}

	return variants
}
