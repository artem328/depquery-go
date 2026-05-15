package yaml

import (
	"github.com/artem328/depquery-go/internal/gen/schema"
	"gopkg.in/yaml.v3"
)

func (p *Parser) isString(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!str"
}

func (p *Parser) consumeString(n *yaml.Node, notAString ...string) string {
	if p.isString(n) {
		return n.Value
	}

	errMsg := "expected string"
	if len(notAString) > 0 && notAString[0] != "" {
		errMsg = notAString[0]
	}

	p.err(n, errMsg)

	return ""
}

func (p *Parser) consumeStringValue(n *yaml.Node, notAString ...string) schema.Value[string] {
	return schema.NewValue(p.consumeString(n, notAString...), p.def(n))
}
