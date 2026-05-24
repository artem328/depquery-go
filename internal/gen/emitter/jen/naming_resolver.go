package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type resolverNaming struct {
	Type              string
	EntityConstructor []string
	NestedConstructor []string
}

func (n *resolverNaming) warmUp(p plan.Plan) {
	n.Type = "resolver"

	n.EntityConstructor = make([]string, len(p.EntityResolvers))

	for id := range p.EntityResolvers {
		erid := plan.EntityResolverID(id)

		n.EntityConstructor[erid] = resolveResolverEntityConstructorName(p, erid)
	}

	n.NestedConstructor = make([]string, len(p.NestedResolvers))
	for id := range p.NestedResolvers {
		nrid := plan.NestedResolverID(id)

		n.NestedConstructor[nrid] = resolveResolverNestedConstructorName(p, nrid)
	}
}

func resolveResolverEntityConstructorName(p plan.Plan, erid plan.EntityResolverID) string {
	er := p.EntityResolvers[erid]

	return "resolve" + sanitizeID(p.Model.Entities[er.Entity].Name, sanitizeRawCapitalized) + "Relations"
}

func resolveResolverNestedConstructorName(p plan.Plan, nrid plan.NestedResolverID) string {
	nr := p.NestedResolvers[nrid]

	return "resolve" + sanitizeID(p.Model.Entities[nr.Entity].Name, sanitizeRawCapitalized) + "Nested"
}
