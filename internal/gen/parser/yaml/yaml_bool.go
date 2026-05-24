package yaml

import (
	"github.com/artem328/depquery-go/internal/gen/schema"
	"gopkg.in/yaml.v3"
)

func (p *Parser) isBool(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!bool"
}

func (p *Parser) consumeBool(n *yaml.Node, notABool ...string) bool {
	if p.isBool(n) {
		return n.Value == "true"
	}

	errMsg := "expected boolean"
	if len(notABool) > 0 && notABool[0] != "" {
		errMsg = notABool[0]
	}

	p.err(n, errMsg)

	return false
}

func (p *Parser) consumeBoolValue(n *yaml.Node, notABool ...string) schema.Value[bool] {
	return schema.NewValue(p.consumeBool(n, notABool...), p.def(n))
}
