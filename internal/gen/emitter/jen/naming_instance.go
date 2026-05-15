package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type instanceNaming struct {
	Interface             string
	ResolveMethod         string
	Struct                string
	FieldFetchContext     string
	FieldResolvers        string
	FieldPrefetchResolver string
	FieldEntityPrefetcher string
	PrefetchMethod        string
}

func (n *instanceNaming) warmUp(plan.Plan) {
	n.Interface = "Instance"
	n.ResolveMethod = "Resolve"
	n.Struct = "instance"
	n.FieldFetchContext = "ctx"
	n.FieldResolvers = "resolvers"
	n.FieldPrefetchResolver = "resolver"
	n.FieldEntityPrefetcher = "prefetcher"
	n.PrefetchMethod = "prefetch"
}
