package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type candidateNaming struct {
	Interface       string
	CandidateMethod string
	ResolverMethod  string
}

func (n *candidateNaming) warmUp(plan.Plan) {
	n.Interface = "candidate"
	n.CandidateMethod = "Candidate"
	n.ResolverMethod = "Resolver"
}
