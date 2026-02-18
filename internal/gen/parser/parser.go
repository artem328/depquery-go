package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

var errNoEntities = errors.New("no entities found")

type parseError struct {
	filename string
	msg      string
	line     int
	col      int
}

func (e parseError) Error() string {
	if e.line == 0 {
		return e.msg
	}

	return fmt.Sprintf("parse %s:%d:%d: %s", e.filename, e.line, e.col, e.msg)
}

type genericError struct {
	filename string
	entity   string
	msg      string
	n        *yaml.Node
}

func (e genericError) Error() string {
	var sb strings.Builder

	if e.entity != "" {
		sb.WriteString("entity ")
		sb.WriteString(e.entity)
		sb.WriteString(": ")
	}

	sb.WriteString(e.msg)

	if e.n != nil {
		sb.WriteString(". definition at ")
		sb.WriteString(e.filename)
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(e.n.Line))
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(e.n.Column))
	}

	return sb.String()
}

type Parser struct {
	filename      string
	root          *yaml.Node
	entities      []*entity
	entitiesByID  map[string]*entity
	errs          []error
	schema        []*schema.Entity
	goTypes       map[string]goParsedType
	goEntityTypes map[string]goParsedEntity
}

func New(root *yaml.Node, filename string) *Parser {
	return &Parser{
		filename:      filename,
		root:          root,
		entitiesByID:  make(map[string]*entity),
		goTypes:       make(map[string]goParsedType),
		goEntityTypes: make(map[string]goParsedEntity),
	}
}

func FromFile(file string) (*Parser, error) {
	var err error

	file, err = filepath.Abs(file)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return FromReader(bufio.NewReader(f), file)
}

func FromReader(r io.Reader, filename string) (*Parser, error) {
	var root yaml.Node

	if err := yaml.NewDecoder(r).Decode(&root); err != nil {
		return nil, err
	}

	return New(&root, filename), nil
}

func (p *Parser) Parse() ([]*schema.Entity, []error) {
	p.parseYAML()

	if errs, ok := p.errors(); ok {
		return nil, errs
	}

	if len(p.entities) == 0 {
		return nil, []error{errNoEntities}
	}

	p.parseGo()

	if errs, ok := p.errors(); ok {
		return nil, errs
	}

	p.buildSchema()

	if errs, ok := p.errors(); ok {
		return nil, errs
	}

	return p.schema, nil
}

func (p *Parser) gerr(entity, msg string, n *yaml.Node) {
	p.err(genericError{filename: p.filename, entity: entity, msg: msg, n: n})
}

func (p *Parser) err(err error) {
	p.errs = append(p.errs, err)
}

func (p *Parser) errors() ([]error, bool) {
	return p.errs, len(p.errs) > 0
}
