package jen

import (
	"strings"

	"github.com/artem328/depquery-go/internal/gen/plan"
)

type entityPrefetcherNaming struct {
	Interface string
	Method    []string
}

func (n *entityPrefetcherNaming) warmUp(p plan.Plan) {
	n.Interface = "EntityPrefetcher"

	n.Method = make([]string, len(p.PrefetchMethods))

	for id, pm := range p.PrefetchMethods {
		pmid := plan.PrefetchMethodID(id)

		n.Method[pmid] = resolveEntityPrefetcherMethodName(p, pm)
	}
}

func resolveEntityPrefetcherMethodName(p plan.Plan, pm plan.PrefetchMethod) string {
	var sb strings.Builder

	sb.WriteString("Prefetch")
	sb.WriteString(sanitizeID(p.Model.Entities[pm.Entity].Name, sanitizeRawCapitalized))

	if pm.Reversed {
		sb.WriteString("By")
		sb.WriteString(sanitizeID(p.Model.Entities[pm.ReversedByEntity].Name, sanitizeRawCapitalized))
	}

	return sb.String()
}
