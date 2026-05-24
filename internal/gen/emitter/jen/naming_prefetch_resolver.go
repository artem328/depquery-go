package jen

import (
	"fmt"
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

	sb.WriteString("Resolve")

	switch rmm := rm.(type) {
	case plan.RelationResolveMethod:
		rel := p.Relations[rmm.Relation]
		f := p.Model.Entities[rel.From]

		sb.WriteString(sanitizeID(f.Name, sanitizeRawCapitalized))

		if rel.Variant.Set {
			v := p.Model.Variants[rel.Variant.V]

			sb.WriteString("V")
			sb.WriteString(sanitizeID(v.Name, sanitizeRawCapitalized))
		}

		sb.WriteString(sanitizeID(rel.Name, sanitizeRawCapitalized))
	case plan.NestedResolveMethod:
		n := p.Nesteds[rmm.Nested]
		f := p.Model.Entities[n.From]

		sb.WriteString(sanitizeID(f.Name, sanitizeRawCapitalized))
		sb.WriteString("N")
		sb.WriteString(sanitizeID(n.Name, sanitizeRawCapitalized))
	default:
		panic(fmt.Errorf("unknown resolve method type: %T", rm))
	}

	return sb.String()
}
