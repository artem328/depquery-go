package yaml

import (
	"github.com/artem328/depquery-go/internal/gen/schema"
	"gopkg.in/yaml.v3"
)

func (p *Parser) parseRelations(n *yaml.Node) []schema.Relation {
	var rr []schema.Relation

	for _, r := range p.consumeSeq(n, "expected a list of relations") {
		rr = append(rr, p.parseRelation(r))
	}

	return rr
}

func (p *Parser) parseRelation(n *yaml.Node) schema.Relation {
	switch {
	case p.isMap(n):
		return p.parseRelationObject(n)
	case p.isString(n):
		return p.parseRelationString(n)
	default:
		p.err(n, "entity relation must be either entity name or extended config object")
		return schema.Relation{}
	}
}

func (p *Parser) parseRelationObject(n *yaml.Node) schema.Relation {
	var r schema.Relation
	r.Definition = p.def(n)

	for key, value := range p.consumeMap(n, "expected relation config object") {
		switch key.Value {
		case "name":
			r.Name = p.consumeStringValue(value)
		case "entity":
			r.Entity = p.consumeStringValue(value)
		case "reversedBy":
			r.ReversedBy = p.consumeStringValue(value)
		default:
			p.err(key, "unknown relation key: "+key.Value)
		}
	}

	return r
}

func (p *Parser) parseRelationString(n *yaml.Node) schema.Relation {
	name := p.consumeStringValue(n, "expected entity name")

	return schema.Relation{
		Name:       name,
		Entity:     name,
		Definition: p.def(n),
	}
}
