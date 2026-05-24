package depquery

type BFSPlanner struct{}

func (p BFSPlanner) Plan(candidates []Candidate) [][]Candidate {
	childMap := make(map[ID][]Candidate)
	nestedChildMap := make(map[ID][]Candidate)

	var currentLevel []Candidate

	for _, c := range candidates {
		if !c.HasChildren {
			continue
		}

		if c.ParentID == 0 {
			currentLevel = append(currentLevel, c)
		} else if c.Nested {
			nestedChildMap[c.ParentID] = append(nestedChildMap[c.ParentID], c)
		} else {
			childMap[c.ParentID] = append(childMap[c.ParentID], c)
		}
	}

	var consumeNested func(id ID) []Candidate

	consumeNested = func(id ID) []Candidate {
		var cc []Candidate
		for _, c := range nestedChildMap[id] {
			cc = append(cc, c)
			cc = append(cc, consumeNested(c.ID)...)
		}

		nestedChildMap[id] = nil

		return cc
	}

	var result [][]Candidate

	for len(currentLevel) > 0 {
		result = append(result, currentLevel)

		var nested []Candidate
		var nextLevel []Candidate

		for _, c := range currentLevel {
			nested = append(nested, consumeNested(c.ID)...)
		}

		nextLevel = nested

		for _, c := range nested {
			nextLevel = append(nextLevel, childMap[c.ID]...)
			childMap[c.ID] = nil
		}

		for _, c := range currentLevel {
			nextLevel = append(nextLevel, childMap[c.ID]...)
			childMap[c.ID] = nil
		}

		currentLevel = nextLevel
	}

	return result
}
