package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type candidateNaming struct {
	Interface       string
	CandidateMethod string
	ResolverMethod  string
	RelationStruct  []string
	NestedStruct    []string
}

func (n *candidateNaming) warmUp(p plan.Plan) {
	n.Interface = "candidate"
	n.CandidateMethod = "Candidate"
	n.ResolverMethod = "Resolver"

	n.RelationStruct = make([]string, len(p.Builders))
	n.NestedStruct = make([]string, len(p.Builders))

	for id := range p.Builders {
		bid := plan.BuilderID(id)

		n.RelationStruct[bid] = resolveCandidateRelationBuilderStructName(p, bid)
		n.NestedStruct[bid] = resolveCandidateNestedBuilderStructName(p, bid)
	}
}

func resolveCandidateRelationBuilderStructName(p plan.Plan, bid plan.BuilderID) string {
	b := p.Builders[bid]

	switch bb := b.(type) {
	case plan.RootBuilder:
		return sanitizeID(p.Model.Entities[bb.Entity].Name, sanitizeUnexported) + "BuilderRelationsCandidate"
	default:
		return ""
	}
}

func resolveCandidateNestedBuilderStructName(p plan.Plan, bid plan.BuilderID) string {
	b := p.Builders[bid]

	switch bb := b.(type) {
	case plan.RootBuilder:
		return sanitizeID(p.Model.Entities[bb.Entity].Name, sanitizeUnexported) + "BuilderNestedCandidate"
	default:
		return ""
	}
}
