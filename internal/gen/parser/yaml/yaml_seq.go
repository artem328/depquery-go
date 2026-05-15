package yaml

import (
	"iter"

	"gopkg.in/yaml.v3"
)

func (p *Parser) isSeq(n *yaml.Node) bool {
	return n.Kind == yaml.SequenceNode
}

func (p *Parser) consumeSeq(s *yaml.Node, notASeq ...string) iter.Seq2[int, *yaml.Node] {
	if !p.isSeq(s) {
		errMsg := "expected sequence"
		if len(notASeq) > 0 && notASeq[0] != "" {
			errMsg = notASeq[0]
		}

		p.err(s, errMsg)

		return nil
	}

	return func(yield func(int, *yaml.Node) bool) {
		for i, n := range s.Content {
			if !yield(i, n) {
				return
			}
		}
	}
}
