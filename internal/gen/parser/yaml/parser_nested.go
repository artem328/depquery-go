package yaml

import (
	"github.com/artem328/depquery-go/internal/gen/schema"
	"gopkg.in/yaml.v3"
)

func (p *Parser) parseNestedFields(n *yaml.Node) []schema.Nested {
	var nn []schema.Nested

	for _, nest := range p.consumeSeq(n, "expected a list of nested fields") {
		nn = append(nn, p.parseNestedField(nest))
	}

	return nn
}

func (p *Parser) parseNestedField(n *yaml.Node) schema.Nested {
	switch {
	case p.isMap(n):
		return p.parseNestedFieldObject(n)
	case p.isString(n):
		return p.parseNestedFieldString(n)
	default:
		p.err(n, "entity nested field must be either entity name or extended config object")
		return schema.Nested{}
	}
}

func (p *Parser) parseNestedFieldObject(n *yaml.Node) schema.Nested {
	var nest schema.Nested
	nest.Definition = p.def(n)

	for key, value := range p.consumeMap(n, "expected nested config object") {
		switch key.Value {
		case "name":
			nest.Name = p.consumeStringValue(value)
		case "entity":
			nest.Entity = p.consumeStringValue(value)
		default:
			p.err(key, "unknown nested field key: "+key.Value)
		}
	}

	return nest
}

func (p *Parser) parseNestedFieldString(n *yaml.Node) schema.Nested {
	name := p.consumeStringValue(n, "expected entity name")

	return schema.Nested{
		Name:       name,
		Entity:     name,
		Definition: p.def(n),
	}
}
