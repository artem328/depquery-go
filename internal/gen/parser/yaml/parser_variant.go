package yaml

import (
	"github.com/artem328/depquery-go/internal/gen/schema"
	"gopkg.in/yaml.v3"
)

func (p *Parser) parseVariants(n *yaml.Node) []schema.Variant {
	var vv []schema.Variant

	for _, v := range p.consumeSeq(n, "expected a list of variants") {
		vv = append(vv, p.parseVariant(v))
	}

	return vv
}

func (p *Parser) parseVariant(n *yaml.Node) schema.Variant {
	var v schema.Variant
	v.Definition = p.def(n)

	for key, value := range p.consumeMap(n, "expected variant config object") {
		switch key.Value {
		case "name":
			v.Name = p.consumeStringValue(value, "expected variant name")
		case "type":
			v.Type = p.parseType(value)
		case "relations":
			v.Relations = p.parseRelations(value)
		default:
			p.err(key, "unknown variant key: "+key.Value)
		}
	}

	return v
}
