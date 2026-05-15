package parser

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/artem328/depquery-go/internal/gen/parser/yaml"
	"github.com/artem328/depquery-go/internal/gen/schema"
)

type constructor func(filename string, _ io.Reader) (Parser, error)

var parsers = map[string]constructor{
	"yaml": func(filename string, r io.Reader) (Parser, error) { return yaml.NewParser(filename, r) },
	"yml":  func(filename string, r io.Reader) (Parser, error) { return yaml.NewParser(filename, r) },
}

type Parser interface {
	Parse() (schema.Schema, []error)
}

func FromReader(filename string, r io.Reader, typ string) (Parser, error) {
	if fn, ok := parsers[typ]; ok {
		return fn(filename, r)
	}

	return nil, fmt.Errorf("schema of type %q is not supported", typ)
}

func FromFile(path string) (Parser, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	return FromReader(path, bytes.NewReader(data), ext)
}
