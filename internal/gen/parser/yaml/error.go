package yaml

import (
	"gopkg.in/yaml.v3"
)

type parseError struct {
	def Definition
	msg string
}

func (e parseError) Error() string {
	def := e.def.DefinedAt()
	if def == "" {
		return e.msg
	}

	return "parse " + def + ": " + e.msg
}

func (p *Parser) err(n *yaml.Node, msg string) {
	p.errors = append(p.errors, parseError{def: p.def(n), msg: msg})
}

func (p *Parser) def(n *yaml.Node) Definition {
	return Definition{Filename: p.filename, Node: n}
}
