package jen

import (
	"strings"

	"github.com/artem328/depquery-go/internal/gen/plan"
)

type relationNaming struct {
	ConstantName []string
	ChildBuilder []string
}

func (n *relationNaming) warmUp(p plan.Plan) {
	n.ConstantName = make([]string, len(p.Relations))
	n.ChildBuilder = make([]string, len(p.Relations))

	for id := range p.Relations {
		rid := plan.RelationID(id)
		n.ConstantName[rid] = resolveResolverOptionConstantName(p, rid)
		n.ChildBuilder[rid] = resolveChildBuilderName(p, rid)
	}
}

func resolveResolverOptionConstantName(p plan.Plan, rid plan.RelationID) string {
	var sb strings.Builder

	sb.WriteString("rel")

	rel := p.Model.Relations[rid]
	from := p.Model.Entities[rel.From]

	sb.WriteString(sanitizeID(from.Name, sanitizeRawCapitalized))
	if rel.Variant.Set {
		v := p.Model.Variants[rel.Variant.V]
		sb.WriteString("V")
		sb.WriteString(sanitizeID(v.Name, sanitizeRawCapitalized))
	}

	sb.WriteString(sanitizeID(rel.Name, sanitizeRawCapitalized))

	return sb.String()
}

func resolveChildBuilderName(p plan.Plan, rid plan.RelationID) string {
	return sanitizeID(p.Model.Relations[rid].Name, sanitizeUnexported)
}
