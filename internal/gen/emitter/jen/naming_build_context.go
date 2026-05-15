package jen

import "github.com/artem328/depquery-go/internal/gen/plan"

type buildContextNaming struct {
	Type                string
	Constructor         string
	FieldCandidates     string
	FieldCandidatesByID string
	FieldLastId         string
	NextIDMethod        string
	AddCandidateMethod  string
	PlanMethod          string
}

func (n *buildContextNaming) warmUp(plan.Plan) {
	n.Type = "buildContext"
	n.Constructor = "newBuildContext"
	n.FieldCandidates = "candidates"
	n.FieldCandidatesByID = "candidatesByID"
	n.FieldLastId = "lastID"
	n.NextIDMethod = "NextID"
	n.AddCandidateMethod = "AddCandidate"
	n.PlanMethod = "Plan"
}
