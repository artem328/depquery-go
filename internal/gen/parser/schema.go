package parser

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

type entityVariant struct {
	Name      string
	Type      *goType
	Relations []relation
}

type entity struct {
	Name      string
	Type      *goType
	ID        id
	Relations []relation
	Variants  []entityVariant

	source *yaml.Node
}

type id struct {
	Member     string
	MemberType schema.MemberType
	Type       *goType

	source *yaml.Node
}

type goType struct {
	Raw     string
	Package string
	Name    string
	Wrapper string
	Params  []*goType

	source *yaml.Node
}

func (t *goType) IsSet() bool {
	return t.Name != ""
}

func (t *goType) IsRaw() bool {
	return t.Raw != ""
}

func (t *goType) String() string {
	if t.Package == "" {
		return t.Name
	}

	return t.Package + "." + t.Name
}

func (t *goType) Parse() error {
	parts := strings.Split(t.Raw, ".")

	switch len(parts) {
	case 1:
		t.Name = parts[0]
	case 2:
		t.Package = parts[0]
		t.Name = parts[1]
	default:
		return fmt.Errorf("invalid type format: %s. should be either pkgname.Type or Type", t.Raw)
	}

	return nil
}

type relation struct {
	Entity   string
	Name     string
	Reversed *id

	source *yaml.Node
}
