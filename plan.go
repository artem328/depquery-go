package depquery

type Candidate struct {
	ID          ID
	ParentID    ID
	HasChildren bool
	Nested      bool
}

type Planner interface {
	Plan([]Candidate) [][]Candidate
}
