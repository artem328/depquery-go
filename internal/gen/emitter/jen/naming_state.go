package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type stateNaming struct {
	Interface                     string
	Struct                        string
	Constructor                   string
	StateContainerField           []string
	StateContainerReversedByField []string
	ResolveMethod                 []string
	ShouldResolveMethod           []string
	ResolveReversedByMethod       []string
	ShouldResolveReversedByMethod []string
	AdderMethod                   []string
}

func (n *stateNaming) warmUp(p plan.Plan) {
	n.Interface = "State"
	n.Struct = "state"
	n.Constructor = "newState"

	n.StateContainerField = make([]string, len(p.StateContainers))
	n.ResolveMethod = make([]string, len(p.StateContainers))
	n.ShouldResolveMethod = make([]string, len(p.StateContainers))
	n.AdderMethod = make([]string, len(p.StateContainers))

	for id := range p.StateContainers {
		scid := plan.StateContainerID(id)

		n.StateContainerField[scid] = resolveStateContainerFieldName(p, scid)
		n.ResolveMethod[scid] = resolveStateResolveMethodName(p, scid)
		n.ShouldResolveMethod[scid] = resolveStateShouldResolveMethodName(p, scid)
		n.AdderMethod[scid] = resolveStateAdderMethodName(p, scid)
	}

	n.StateContainerReversedByField = make([]string, len(p.ReversedStateContainers))
	n.ResolveReversedByMethod = make([]string, len(p.ReversedStateContainers))
	n.ShouldResolveReversedByMethod = make([]string, len(p.ReversedStateContainers))

	for id := range p.ReversedStateContainers {
		rscid := plan.ReversedStateContainerID(id)

		n.StateContainerReversedByField[rscid] = resolveStateContainerReversedByFieldName(p, rscid)
		n.ResolveReversedByMethod[rscid] = resolveStateResolveReversedByMethodName(p, rscid)
		n.ShouldResolveReversedByMethod[rscid] = resolveStateShouldResolveReversedByMethodName(p, rscid)
	}
}

func resolveStateContainerFieldName(p plan.Plan, scid plan.StateContainerID) string {
	sc := p.StateContainers[scid]
	e := p.Model.Entities[sc.Entity]

	return sanitizeID(e.Name, sanitizeUnexported)
}

func resolveStateContainerReversedByFieldName(p plan.Plan, rscid plan.ReversedStateContainerID) string {
	rsc := p.ReversedStateContainers[rscid]
	sc := p.StateContainers[rsc.StateContainerID]

	return sanitizeID(p.Model.Entities[sc.Entity].Name, sanitizeUnexported) + "By" + sanitizeID(p.Model.Entities[rsc.HolderEntity].Name, sanitizeRawCapitalized)
}

func resolveStateResolveMethodName(p plan.Plan, scid plan.StateContainerID) string {
	sc := p.StateContainers[scid]

	return sanitizeID(p.Model.Entities[sc.Entity].Name, sanitizeExported)
}

func resolveStateShouldResolveMethodName(p plan.Plan, scid plan.StateContainerID) string {
	sc := p.StateContainers[scid]

	return "Should" + sanitizeID(p.Model.Entities[sc.Entity].Name, sanitizeRawCapitalized)
}

func resolveStateResolveReversedByMethodName(p plan.Plan, rscid plan.ReversedStateContainerID) string {
	rsc := p.ReversedStateContainers[rscid]
	sc := p.StateContainers[rsc.StateContainerID]

	return sanitizeID(p.Model.Entities[sc.Entity].Name, sanitizeExported) + "By" + sanitizeID(p.Model.Entities[rsc.HolderEntity].Name, sanitizeRawCapitalized)
}

func resolveStateShouldResolveReversedByMethodName(p plan.Plan, rscid plan.ReversedStateContainerID) string {
	rsc := p.ReversedStateContainers[rscid]
	sc := p.StateContainers[rsc.StateContainerID]

	return "Should" + sanitizeID(p.Model.Entities[sc.Entity].Name, sanitizeRawCapitalized) + "By" + sanitizeID(p.Model.Entities[rsc.HolderEntity].Name, sanitizeRawCapitalized)
}

func resolveStateAdderMethodName(p plan.Plan, scid plan.StateContainerID) string {
	sc := p.StateContainers[scid]

	return "add" + sanitizeID(p.Model.Entities[sc.Entity].Name, sanitizeRawCapitalized)
}
