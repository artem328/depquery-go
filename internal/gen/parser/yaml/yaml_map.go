package yaml

import (
	"iter"

	"gopkg.in/yaml.v3"
)

func (p *Parser) isMap(n *yaml.Node) bool {
	return n.Kind == yaml.MappingNode && len(n.Content)%2 == 0
}

func (p *Parser) consumeMap(n *yaml.Node, notAMap ...string) iter.Seq2[*yaml.Node, *yaml.Node] {
	if !p.isMap(n) {
		errMsg := "expected map"
		if len(notAMap) > 0 && notAMap[0] != "" {
			errMsg = notAMap[0]
		}

		p.err(n, errMsg)

		return func(func(*yaml.Node, *yaml.Node) bool) {}
	}

	return func(yield func(*yaml.Node, *yaml.Node) bool) {
		for i := 0; i < len(n.Content); i += 2 {
			key := n.Content[i]
			value := n.Content[i+1]

			if !yield(key, value) {
				return
			}
		}
	}
}
