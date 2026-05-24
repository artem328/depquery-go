package plan

import "github.com/artem328/depquery-go/internal/gen/semantic"

func (p *Planner) initNested() {
	p.nestedsByEntity = make(map[semantic.EntityID][]NestedID)
	p.plan.Nesteds = make([]Nested, 0, len(p.model.Nesteds))

	for id, n := range p.model.Nesteds {
		nid := NestedID(id)
		p.plan.Nesteds = append(p.plan.Nesteds, Nested{
			ID:     nid,
			Nested: n,
		})
		p.nestedsByEntity[n.From] = append(p.nestedsByEntity[n.From], nid)
	}
}
