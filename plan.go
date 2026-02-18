package depquery

type Candidate struct {
	ID        ID
	ParentID  ID
	Relations uint64
}

type Planner interface {
	Plan([]Candidate) [][]Candidate
}

type TopologicalPlanner struct{}

func (p TopologicalPlanner) Plan(candidates []Candidate) [][]Candidate {
	childMap := make(map[ID][]Candidate)
	idToCandidate := make(map[ID]Candidate)

	var currentLevel []Candidate

	for _, c := range candidates {
		idToCandidate[c.ID] = c

		if c.ParentID == 0 {
			currentLevel = append(currentLevel, c)
		} else {
			childMap[c.ParentID] = append(childMap[c.ParentID], c)
		}
	}

	var result [][]Candidate

	for len(currentLevel) > 0 {
		result = append(result, currentLevel)

		var nextLevel []Candidate
		for _, c := range currentLevel {
			nextLevel = append(nextLevel, childMap[c.ID]...)
		}

		currentLevel = nextLevel
	}

	return result
}
