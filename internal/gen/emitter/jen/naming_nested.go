package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type nestedNaming struct {
	ConstantName []string
	ChildBuilder []string
}

func (n *nestedNaming) warmUp(p plan.Plan) {
	n.ConstantName = make([]string, len(p.Nesteds))
	n.ChildBuilder = make([]string, len(p.Nesteds))

	for id := range p.Nesteds {
		nid := plan.NestedID(id)

		n.ConstantName[nid] = resolveNestedConstantName(p, nid)
		n.ChildBuilder[nid] = resolveNestedChildBuilderName(p, nid)
	}
}

func resolveNestedConstantName(p plan.Plan, nid plan.NestedID) string {
	n := p.Nesteds[nid]

	return "nes" + sanitizeID(p.Model.Entities[n.From].Name, sanitizeRawCapitalized) + sanitizeID(n.Name, sanitizeRawCapitalized)
}

func resolveNestedChildBuilderName(p plan.Plan, nid plan.NestedID) string {
	n := p.Nesteds[nid]

	return "n" + sanitizeID(n.Name, sanitizeRawCapitalized)
}
