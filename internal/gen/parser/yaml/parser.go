package yaml

import (
	"io"

	"github.com/artem328/depquery-go/internal/gen/schema"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	filename string
	root     *yaml.Node
	schema   schema.Schema
	errors   []error
}

func NewParser(filename string, r io.Reader) (*Parser, error) {
	var root yaml.Node

	if err := yaml.NewDecoder(r).Decode(&root); err != nil {
		return nil, err
	}

	return &Parser{filename: filename, root: &root}, nil
}

func (p *Parser) Parse() (schema.Schema, []error) {
	p.parse()

	return p.schema, p.errors
}

func (p *Parser) parse() {
	if p.root.Kind != yaml.DocumentNode {
		p.err(p.root, "not a yaml document")

		return
	}

	for _, n := range p.root.Content {
		for key, value := range p.consumeMap(n, "expected entities mapping") {
			p.schema = append(p.schema, p.parseEntity(key, value))
		}
	}
}
