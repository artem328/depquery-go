package depquery

type Candidate struct {
	ID          ID
	SubID       uint
	ParentID    ID
	HasChildren bool
	Nested      bool
}

type Planner interface {
	Plan([]Candidate) [][]Candidate
}
