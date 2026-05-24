package depquery

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBFSPlanner_Plan(t *testing.T) {
	t.Parallel()

	t.Run("FlattensNestedCandidates", func(t *testing.T) {
		t.Parallel()

		// A
		// ├── B (nested)
		// │   └── C (nested)
		// │       └── D
		// └── E

		candidates := []Candidate{
			{ID: 2, ParentID: 1, Nested: true, HasChildren: true}, // B
			{ID: 1, HasChildren: true},                            // A
			{ID: 4, ParentID: 3, HasChildren: true},               // D
			{ID: 3, ParentID: 2, Nested: true, HasChildren: true}, // C
			{ID: 5, ParentID: 1, HasChildren: true},               // E
		}

		planner := BFSPlanner{}

		got := planner.Plan(candidates)

		want := [][]Candidate{
			{
				{ID: 1, HasChildren: true}, // A
			},
			{
				{ID: 2, ParentID: 1, Nested: true, HasChildren: true}, // B
				{ID: 3, ParentID: 2, Nested: true, HasChildren: true}, // C
				{ID: 4, ParentID: 3, HasChildren: true},               // D
				{ID: 5, ParentID: 1, HasChildren: true},               // E
			},
		}

		require.Equal(t, want, got)
	})
}
