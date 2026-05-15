package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type resolverNaming struct {
	Type        string
	Constructor []string
}

func (n *resolverNaming) warmUp(p plan.Plan) {
	n.Type = "resolver"

	n.Constructor = make([]string, len(p.EntityResolvers))

	for id := range p.EntityResolvers {
		erid := plan.EntityResolverID(id)

		n.Constructor[erid] = resolveResolverConstructorName(p, erid)
	}
}

func resolveResolverConstructorName(p plan.Plan, erid plan.EntityResolverID) string {
	er := p.EntityResolvers[erid]

	return "resolve" + sanitizeID(p.Model.Entities[er.Entity].Name, sanitizeRawCapitalized) + "Relations"
}
