package jen

import (
	"strings"

	"github.com/artem328/depquery-go/internal/gen/plan"
)

type prefetchResolverNaming struct {
	Interface string
	Method    []string
}

func (n *prefetchResolverNaming) warmUp(p plan.Plan) {
	n.Interface = "PrefetchResolver"

	n.Method = make([]string, len(p.ResolveMethods))

	for id := range p.ResolveMethods {
		rmid := plan.ResolveMethodID(id)

		n.Method[rmid] = resolvePrefetchResolverMethodName(p, rmid)
	}
}

func resolvePrefetchResolverMethodName(p plan.Plan, rmid plan.ResolveMethodID) string {
	var sb strings.Builder

	rm := p.ResolveMethods[rmid]
	rel := p.Relations[rm.Relation]
	f := p.Model.Entities[rel.From]

	sb.WriteString("Resolve")
	sb.WriteString(sanitizeID(f.Name, sanitizeRawCapitalized))

	if rel.Variant.Set {
		v := p.Model.Variants[rel.Variant.V]

		sb.WriteString("V")
		sb.WriteString(sanitizeID(v.Name, sanitizeRawCapitalized))
	}

	sb.WriteString(sanitizeID(rel.Name, sanitizeRawCapitalized))

	return sb.String()
}
