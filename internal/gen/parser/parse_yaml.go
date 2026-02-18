package parser

import (
	"fmt"
	"iter"

	"gopkg.in/yaml.v3"
)

func (p *Parser) parseYAML() {
	if p.root.Kind != yaml.DocumentNode {
		p.yamlerr(p.root, "not a yaml document")

		return
	}

	for _, n := range p.root.Content {
		for key, value := range p.yamlMap(n, "expected entities configuration") {
			p.yamlEntity(key.Value, value)
		}
	}
}

func (p *Parser) yamlEntity(name string, n *yaml.Node) {
	var e entity

	if _, ok := p.entitiesByID[name]; ok {
		p.yamlerr(n, fmt.Sprintf("entity %s duplicated", name))

		return
	}

	for key, value := range p.yamlMap(n, "expected entity configuration") {
		switch key.Value {
		case "type":
			e.Type = p.yamlType(value)
		case "id":
			e.ID = p.yamlEntityID(value)
		case "relations":
			e.Relations = p.yamlRelations(value)
		case "variants":
			e.Variants = p.yamlEntityVariants(value)
		default:
			p.yamlerr(key, "unknown entity key: "+key.Value)
		}
	}

	e.Name = name
	e.source = n

	ee := &e

	p.entities = append(p.entities, ee)
	p.entitiesByID[name] = ee
}

func (p *Parser) yamlType(n *yaml.Node) *goType {
	var t goType

	switch {
	case p.yamlIsMap(n):
		return p.yamlTypeObject(n)
	case p.yamlIsString(n):
		t.Raw = p.yamlString(n, "")
		t.source = n
	default:
		p.yamlerr(n, "type must be a string or extended type config object")
	}

	return &t
}

func (p *Parser) yamlTypeObject(n *yaml.Node) *goType {
	var t goType

	for key, value := range p.yamlMap(n, "") {
		switch key.Value {
		case "base":
			t.Raw = p.yamlString(value, "expected type string")
		case "params":
			for _, param := range p.yamlSeq(value, "expected type params sequence") {
				t.Params = append(t.Params, p.yamlType(param))
			}
		}
	}

	t.source = n

	return &t
}

func (p *Parser) yamlEntityID(n *yaml.Node) id {
	var i id

	i.Member = p.yamlString(n, "expected entity member holding its ID")
	i.source = n

	return i
}

func (p *Parser) yamlRelations(n *yaml.Node) []relation {
	relations := make([]relation, 0)

	for _, rr := range p.yamlSeq(n, "expected a list of relations") {
		relations = append(relations, p.yamlRelation(rr))
	}

	return relations
}

func (p *Parser) yamlRelation(n *yaml.Node) relation {
	var r relation

	switch {
	case p.yamlIsMap(n):
		r = p.yamlRelationObject(n)
	case p.yamlIsString(n):
		r.Entity = p.yamlString(n, "expected entity name")
		r.Name = r.Entity
	default:
		p.yamlerr(n, "entity relation must be either entity name or extended config object")
	}

	r.source = n

	return r
}

func (p *Parser) yamlRelationObject(n *yaml.Node) relation {
	var r relation

	for key, value := range p.yamlMap(n, "expected relation config object") {
		switch key.Value {
		case "name":
			r.Name = p.yamlString(value, "")
		case "entity":
			r.Entity = p.yamlString(value, "")
		case "reversedBy":
			rev := p.yamlEntityID(value)
			r.Reversed = &rev
		default:
			p.yamlerr(key, "unknown relation key: "+key.Value)
		}
	}

	return r
}

func (p *Parser) yamlEntityVariants(n *yaml.Node) []entityVariant {
	variants := make([]entityVariant, 0)

	for _, v := range p.yamlSeq(n, "expected a list of entity variants") {
		variants = append(variants, p.yamlEntityVariant(v))
	}

	return variants
}

func (p *Parser) yamlEntityVariant(n *yaml.Node) entityVariant {
	var r entityVariant

	for key, value := range p.yamlMap(n, "expected variant config object") {
		switch key.Value {
		case "name":
			r.Name = p.yamlString(value, "expected entity variant name")
		case "type":
			r.Type = p.yamlType(value)
		case "relations":
			r.Relations = p.yamlRelations(value)
		default:
			p.yamlerr(key, "unknown entity variant key: "+key.Value)
		}
	}

	return r
}

func (p *Parser) yamlIsString(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!str"
}

func (p *Parser) yamlIsMap(n *yaml.Node) bool {
	return n.Kind == yaml.MappingNode && len(n.Content)%2 == 0
}

func (p *Parser) yamlString(n *yaml.Node, notAString string) string {
	if p.yamlIsString(n) {
		return n.Value
	}

	if notAString == "" {
		notAString = "expected string"
	}

	p.yamlerr(n, notAString)

	return ""
}

func (p *Parser) yamlMap(n *yaml.Node, notAMap string) iter.Seq2[*yaml.Node, *yaml.Node] {
	if !p.yamlIsMap(n) {
		if notAMap == "" {
			notAMap = "expected map"
		}

		p.yamlerr(n, notAMap)

		return nil
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

func (p *Parser) yamlSeq(s *yaml.Node, notASeq string) iter.Seq2[int, *yaml.Node] {
	if s.Kind != yaml.SequenceNode {
		if notASeq != "" {
			notASeq = "expected sequence"
		}

		p.yamlerr(s, notASeq)

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

func (p *Parser) yamlerr(n *yaml.Node, msg string) {
	p.err(parseError{filename: p.filename, msg: msg, line: n.Line, col: n.Column})
}
