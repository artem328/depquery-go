package semantic

import (
	"strings"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

const definitionAtPrefix = ". definition at "

type wrappedError struct {
	err error
	def schema.Definition
}

func (e wrappedError) Error() string {
	if e.def == nil {
		return e.err.Error()
	}

	def := e.def.DefinedAt()
	if def == "" {
		return e.err.Error()
	}

	return e.err.Error() + definitionAtPrefix + def
}

type entityError struct {
	entity string
	msg    string
	def    schema.Definition
}

func (e entityError) Error() string {
	var sb strings.Builder

	if e.entity != "" {
		sb.WriteString("entity ")
		sb.WriteString(e.entity)
		sb.WriteString(": ")
	}

	sb.WriteString(e.msg)

	if e.def == nil {
		return sb.String()
	}

	def := e.def.DefinedAt()
	if def != "" {
		sb.WriteString(definitionAtPrefix)
		sb.WriteString(def)
	}

	return sb.String()
}

func (r *Resolver) err(err error) {
	r.errors = append(r.errors, err)
}

func (r *Resolver) werr(err error, def schema.Definition) {
	r.err(wrappedError{err: err, def: def})
}

func (r *Resolver) eerr(entity, msg string, def schema.Definition) {
	r.err(entityError{entity: entity, msg: msg, def: def})
}
