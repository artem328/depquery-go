package plan

import (
	"github.com/artem328/depquery-go/internal/gen/semantic"
)

type Plan struct {
	Model                   semantic.Model
	Relations               []Relation
	Builders                []Builder
	EntityResolvers         []EntityResolver
	StateContainers         []StateContainer
	ReversedStateContainers []ReversedStateContainer
	PrefetchMethods         []PrefetchMethod
	ResolveMethods          []ResolveMethod
	FetchParents            []FetchParent
	ReversedFetchParents    []ReversedFetchParent
	FetchChildren           []FetchChild
	EntityFetches           []EntityFetch
	FetchContextRoots       []FetchContextRoot
	ParentFetchGetters      []ParentFetchGetter
}
