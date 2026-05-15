package yaml

import (
	"github.com/artem328/depquery-go/internal/gen/schema"
	"gopkg.in/yaml.v3"
)

func (p *Parser) parseEntity(name, entity *yaml.Node) schema.Entity {
	var e schema.Entity

	e.Name = schema.NewValue(name.Value, p.def(name))
	e.Definition = p.def(name)

	for key, value := range p.consumeMap(entity, "expected entity configuration") {
		switch key.Value {
		case "type":
			e.Type = p.parseType(value)
		case "id":
			e.ID = p.consumeStringValue(value)
		case "relations":
			e.Relations = p.parseRelations(value)
		case "variants":
			e.Variants = p.parseVariants(value)
		default:
			p.err(key, "unknown entity key: "+key.Value)
		}
	}

	return e
}
