package jen

import (
	"strings"

	"github.com/artem328/depquery-go/internal/gen/plan"
)

type fetchContextNaming struct {
	Struct                  string
	Constructor             string
	ConstructorForEntity    []string
	FieldByParentID         []string
	FieldSyntheticState     []string
	FieldChildSeen          []string
	FieldPending            []string
	EnqueueMethod           []string
	ParentGetterMethod      []string
	AddNestedMethod         []string
	FlushMethod             string
	SyntheticNamespaceConst string
}

func (n *fetchContextNaming) warmUp(p plan.Plan) {
	n.Struct = "fetchContext"
	n.Constructor = "newFetchContext"
	n.FlushMethod = "flush"
	n.SyntheticNamespaceConst = "syntheticIDNamespace"

	n.FieldByParentID = make([]string, len(p.FetchParents))

	for id := range p.FetchParents {
		fpid := plan.FetchParentID(id)

		n.FieldByParentID[fpid] = resolveFetchContextFieldByParentIDName(p, fpid)
	}

	n.FieldChildSeen = make([]string, len(p.FetchChildren))
	n.FieldPending = make([]string, len(p.FetchChildren))

	for id := range p.FetchChildren {
		fcid := plan.FetchChildID(id)

		n.FieldChildSeen[fcid] = resolveFetchContextFieldChildSeenName(p, fcid)
		n.FieldPending[fcid] = resolveFetchContextFieldPendingName(p, fcid)
	}

	n.ConstructorForEntity = make([]string, len(p.FetchContextRoots))

	for id := range p.FetchContextRoots {
		fcrid := plan.FetchContextRootID(id)

		n.ConstructorForEntity[fcrid] = resolveFetchContextConstructorForEntityName(p, fcrid)
	}

	n.EnqueueMethod = make([]string, len(p.EntityFetches))

	for id := range p.EntityFetches {
		efid := plan.EntityFetchID(id)

		n.EnqueueMethod[efid] = resolveFetchContextEnqueueMethodName(p, efid)
	}

	n.FieldSyntheticState = make([]string, len(p.SyntheticStateContainers))
	for id := range p.SyntheticStateContainers {
		sscid := plan.SyntheticStateContainerID(id)

		n.FieldSyntheticState[sscid] = resolveFetchContextFieldSyntheticStateName(p, sscid)
	}

	n.ParentGetterMethod = make([]string, len(p.FetchParents))

	for id := range p.ParentFetchGetters {
		pfgid := plan.ParentFetchGetterID(id)

		n.ParentGetterMethod[pfgid] = resolveFetchContextParentGetterMethodName(p, pfgid)
	}

	n.AddNestedMethod = make([]string, len(p.NestedEntityFetches))

	for id := range p.NestedEntityFetches {
		nefid := plan.NestedEntityFetchID(id)

		n.AddNestedMethod[nefid] = resolveFetchContextAddNestedMethodName(p, nefid)
	}
}

func resolveFetchContextFieldByParentIDName(p plan.Plan, fpid plan.FetchParentID) string {
	fp := p.FetchParents[fpid]

	var sb strings.Builder

	sb.WriteString(sanitizeID(p.Model.Entities[fp.Entity].Name, sanitizeUnexported))

	if fp.Reversed {
		sb.WriteString("By")
		sb.WriteString(sanitizeID(p.Model.Entities[fp.ReversedByEntity].Name, sanitizeRawCapitalized))
	}

	sb.WriteString("ByID")

	return sb.String()
}

func resolveFetchContextParentGetterMethodName(p plan.Plan, pfgif plan.ParentFetchGetterID) string {
	pfg := p.ParentFetchGetters[pfgif]
	fp := p.FetchParents[pfg.FetchParent]

	return "get" + sanitizeID(p.Model.Entities[fp.Entity].Name, sanitizeRawCapitalized)
}

func resolveFetchContextFieldChildSeenName(p plan.Plan, fcid plan.FetchChildID) string {
	fc := p.FetchChildren[fcid]

	var sb strings.Builder

	sb.WriteString("seen")
	sb.WriteString(sanitizeID(p.Model.Entities[fc.Entity].Name, sanitizeRawCapitalized))

	if fc.Reversed {
		sb.WriteString("By")
		sb.WriteString(sanitizeID(p.Model.Entities[fc.ReversedByEntity].Name, sanitizeRawCapitalized))
	}

	return sb.String()
}

func resolveFetchContextFieldPendingName(p plan.Plan, fcid plan.FetchChildID) string {
	fc := p.FetchChildren[fcid]

	var sb strings.Builder

	sb.WriteString("pending")
	sb.WriteString(sanitizeID(p.Model.Entities[fc.Entity].Name, sanitizeRawCapitalized))

	if fc.Reversed {
		sb.WriteString("By")
		sb.WriteString(sanitizeID(p.Model.Entities[fc.ReversedByEntity].Name, sanitizeRawCapitalized))
	}

	return sb.String()
}

func resolveFetchContextConstructorForEntityName(p plan.Plan, fcrid plan.FetchContextRootID) string {
	fcr := p.FetchContextRoots[fcrid]

	return "new" + sanitizeID(p.Model.Entities[fcr.Entity].Name, sanitizeRawCapitalized) + "FetchContext"
}

func resolveFetchContextEnqueueMethodName(p plan.Plan, efid plan.EntityFetchID) string {
	ef := p.EntityFetches[efid]

	var sb strings.Builder

	sb.WriteString("addPrefetch")
	sb.WriteString(sanitizeID(p.Model.Entities[ef.Entity].Name, sanitizeRawCapitalized))

	if ef.Reversed {
		fc := p.FetchChildren[ef.Child]

		sb.WriteString("By")
		sb.WriteString(sanitizeID(p.Model.Entities[fc.ReversedByEntity].Name, sanitizeRawCapitalized))
	}

	return sb.String()
}

func resolveFetchContextFieldSyntheticStateName(p plan.Plan, sscid plan.SyntheticStateContainerID) string {
	ssc := p.SyntheticStateContainers[sscid]

	return "synth" + sanitizeID(p.Model.Entities[ssc.Entity].Name, sanitizeRawCapitalized)
}

func resolveFetchContextAddNestedMethodName(p plan.Plan, nefid plan.NestedEntityFetchID) string {
	nef := p.NestedEntityFetches[nefid]

	return "addNested" + sanitizeID(p.Model.Entities[nef.Entity].Name, sanitizeRawCapitalized)
}
