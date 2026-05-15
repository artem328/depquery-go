package yaml

import (
	"github.com/artem328/depquery-go/internal/gen/schema"
	"gopkg.in/yaml.v3"
)

func (p *Parser) parseType(n *yaml.Node) schema.Type {
	switch {
	case p.isMap(n):
		return p.parseTypeObject(n)
	case p.isString(n):
		return p.parseTypeString(n)
	default:
		p.err(n, "type must be a string or extended type config object")
		return schema.Type{}
	}
}

func (p *Parser) parseTypeString(n *yaml.Node) schema.Type {
	return schema.Type{
		Base:       p.consumeStringValue(n),
		Definition: p.def(n),
	}
}

func (p *Parser) parseTypeObject(n *yaml.Node) schema.Type {
	var t schema.Type
	t.Definition = p.def(n)

	for key, value := range p.consumeMap(n, "expected extended type config object") {
		switch key.Value {
		case "base":
			t.Base = p.consumeStringValue(value, "expected type path")
		case "params":
			for _, param := range p.consumeSeq(value, "expected type params sequence") {
				t.Params = append(t.Params, p.parseType(param))
			}
		default:
			p.err(key, "unknown type key: "+key.Value)
		}
	}

	return t
}
