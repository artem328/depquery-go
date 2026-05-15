package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

type Planner struct {
	model                            semantic.Model
	plan                             Plan
	relationsByEntity                map[semantic.EntityID][]RelationID
	referencedEntities               map[semantic.EntityID]int
	reverseReferencedEntities        map[semantic.EntityID]int
	referencingEntities              map[semantic.EntityID]int
	referencingVariants              map[semantic.VariantID]int
	containerByEntity                map[semantic.EntityID]StateContainerID
	reversedContainersByRelation     map[RelationID]ReversedStateContainerID
	prefetchMethodByEntity           map[semantic.EntityID]PrefetchMethodID
	reversedPrefetchMethodByRelation map[RelationID]PrefetchMethodID
	fetchByEntity                    map[semantic.EntityID]EntityFetchID
	fetchContextRootByEntity         map[semantic.EntityID]FetchContextRootID
	reversedFetchByRelation          map[RelationID]EntityFetchID
	fetchParentByEntity              map[semantic.EntityID]FetchParentID
	entityResolverByEntity           map[semantic.EntityID]EntityResolverID
}

func NewPlanner(m semantic.Model) *Planner {
	return &Planner{
		model: m,
		plan:  Plan{Model: m},
	}
}

func (p *Planner) Plan() (Plan, error) {
	p.initRelations()
	p.initState()
	p.initPrefetch()
	p.initFetch()
	p.initResolve()
	p.initBuilders()

	return p.plan, nil
}
