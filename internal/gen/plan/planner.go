package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

type Planner struct {
	model                            semantic.Model
	plan                             Plan
	relationsByEntity                map[semantic.EntityID][]RelationID
	nestedsByEntity                  map[semantic.EntityID][]NestedID
	referencedEntities               map[semantic.EntityID]int
	reverseReferencedEntities        map[semantic.EntityID]int
	referencingEntities              map[semantic.EntityID]int
	referencingVariants              map[semantic.VariantID]int
	containerByEntity                map[semantic.EntityID]StateContainerID
	syntheticContainerByEntity       map[semantic.EntityID]SyntheticStateContainerID
	reversedContainersByRelation     map[RelationID]ReversedStateContainerID
	prefetchMethodByEntity           map[semantic.EntityID]PrefetchMethodID
	reversedPrefetchMethodByRelation map[RelationID]PrefetchMethodID
	fetchByEntity                    map[semantic.EntityID]EntityFetchID
	fetchContextRootByEntity         map[semantic.EntityID]FetchContextRootID
	reversedFetchByRelation          map[RelationID]EntityFetchID
	fetchParentByEntity              map[semantic.EntityID]FetchParentID
	parentFetchGetterByFetchParent   map[FetchParentID]ParentFetchGetterID
	entityResolverByEntity           map[semantic.EntityID]EntityResolverID
	nestedResolverByEntity           map[semantic.EntityID]NestedResolverID
	nestedEntityFetchByNested        map[NestedID]NestedEntityFetchID
	idNamespace                      counter[uint64]
}

func NewPlanner(m semantic.Model) *Planner {
	return &Planner{
		model: m,
		plan:  Plan{Model: m},
		idNamespace: counter[uint64]{
			val: 1,
		},
	}
}

func (p *Planner) Plan() (Plan, error) {
	p.initRelations()
	p.initNested()
	p.initState()
	p.initPrefetch()
	p.initFetch()
	p.initResolve()
	p.initBuilders()

	return p.plan, nil
}
