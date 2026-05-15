package yaml

import (
	"strconv"

	"gopkg.in/yaml.v3"
)

type Definition struct {
	Filename string
	*yaml.Node
}

func (d Definition) DefinedAt() string {
	if d.Node == nil {
		return d.Filename
	}

	return d.Filename + ":" + strconv.Itoa(d.Node.Line) + ":" + strconv.Itoa(d.Node.Column)
}
