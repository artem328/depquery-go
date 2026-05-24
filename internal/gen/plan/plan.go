package plan

import (
	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type Plan struct {
	Model                    semantic.Model
	Relations                []Relation
	Nesteds                  []Nested
	Builders                 []Builder
	EntityResolvers          []EntityResolver
	NestedResolvers          []NestedResolver
	StateContainers          []StateContainer
	ReversedStateContainers  []ReversedStateContainer
	SyntheticStateContainers []SyntheticStateContainer
	PrefetchMethods          []PrefetchMethod
	ResolveMethods           []ResolveMethod
	FetchParents             []FetchParent
	ReversedFetchParents     []ReversedFetchParent
	FetchChildren            []FetchChild
	EntityFetches            []EntityFetch
	NestedEntityFetches      []NestedEntityFetch
	FetchContextRoots        []FetchContextRoot
	ParentFetchGetters       []ParentFetchGetter
}
