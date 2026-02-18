package generator

import (
	"context"
	"slices"
	"strings"
	"sync"

	"github.com/dave/jennifer/jen"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func typeToJen(t schema.Type) *jen.Statement {
	j := jen.Op(t.Wrapper).Qual(t.Package, t.Name)
	if len(t.Params) > 0 {
		j.TypesFunc(func(p *jen.Group) {
			for _, tp := range t.Params {
				p.Add(typeToJen(tp))
			}
		})
	}

	return j
}

func memberToJen(i schema.IDMember) *jen.Statement {
	m := jen.Dot(i.Name)
	if i.RcvType == schema.MemberTypeMethod {
		m.Call()
	}

	return m
}

var multilineValuesOpts = jen.Options{
	Open:      "{",
	Close:     "}",
	Separator: ",",
	Multi:     true,
}

type refedReversed struct {
	Entity *schema.Entity
	By     *schema.Entity
}

type reverseRelation struct {
	Ref    *schema.Entity
	Member schema.IDMember
}

func refedEntities(s []*schema.Entity) ([]*schema.Entity, []refedReversed, map[*schema.Entity][]reverseRelation) {
	refed := make(map[*schema.Entity]struct{}, len(s))
	refedRev := make(map[refedReversed]struct{}, len(s))
	reverseRels := make(map[*schema.Entity][]reverseRelation)

	for _, e := range s {
		for _, r := range e.Relations {
			if r.IsReversed() {
				refedRev[refedReversed{Entity: r.Entity, By: e}] = struct{}{}

				reverseRels[r.Entity] = append(reverseRels[r.Entity], reverseRelation{
					Ref:    e,
					Member: *r.ReversedBy,
				})

				continue
			}

			refed[r.Entity] = struct{}{}
		}

		for _, v := range e.Variants {
			for _, r := range v.Relations {
				if r.IsReversed() {
					refedRev[refedReversed{Entity: r.Entity, By: e}] = struct{}{}

					reverseRels[r.Entity] = append(reverseRels[r.Entity], reverseRelation{
						Ref:    e,
						Member: *r.ReversedBy,
					})

					continue
				}

				refed[r.Entity] = struct{}{}
			}
		}
	}

	entities := make([]*schema.Entity, 0, len(refed))
	for e := range refed {
		entities = append(entities, e)
	}

	reversedRefed := make([]refedReversed, 0, len(refedRev))
	for r := range refedRev {
		reversedRefed = append(reversedRefed, r)
	}

	slices.SortStableFunc(entities, func(a, b *schema.Entity) int {
		return strings.Compare(a.Name, b.Name)
	})
	slices.SortStableFunc(reversedRefed, func(a, b refedReversed) int {
		return strings.Compare(a.Entity.Name, b.Entity.Name)
	})

	return entities, reversedRefed, reverseRels
}

func generate(ctx context.Context, workers int, s []*schema.Entity, pkg string) ([]artifact, error) {
	refed, refedRev, reverseRels := refedEntities(s)
	naming := newNameIndex()

	var (
		state            = newStateBuilder(naming, s, refedRev, reverseRels)
		prefetchResolver = newPrefetchResolverBuilder(naming, s)
		entityPrefetcher = newEntityPrefetcherBuilder(naming, refed, refedRev)
		fetchCtx         = newFetchContextBuilder(naming, s, refed, refedRev, reverseRels)
		instance         = newInstanceBuilder(naming, refed, refedRev)
		plan             = newPlanBuilder(naming)
		resolver         = newResolverBuilder(naming, s)
		compiler         = newCompilerBuilder(naming)
		buildContext     = newBuildContextBuilder(naming)
		bder             = newBuilderBuilder(naming, s)
	)

	if err := prebuild(ctx, workers, []builderResolver{
		state,
		prefetchResolver,
		entityPrefetcher,
		fetchCtx,
		instance,
		plan,
		resolver,
		compiler,
		buildContext,
		bder,
	}); err != nil {
		return nil, err
	}

	// depquery file
	dm := jen.NewFile(pkg)

	dm.Add(prefetchResolver.Interface()).Line()
	dm.Add(entityPrefetcher.Interface()).Line()
	dm.Add(state.Interface()).Line()
	dm.Add(instance.Interface()).Line()
	dm.Add(plan.Interface()).Line()

	dm.Add(state.Implementation()).Line()
	dm.Add(instance.Implementation()).Line()
	dm.Add(plan.Implementation()).Line()
	dm.Add(fetchCtx.Implementation()).Line()
	dm.Add(resolver.Implementation()).Line()

	// builder file
	b := jen.NewFile(pkg)

	b.Add(compiler.Interface()).Line()
	b.Add(bder.Interface()).Line()

	b.Add(buildContext.Implementation()).Line()
	b.Add(bder.Implementation())

	return []artifact{
		{f: dm, name: "depquery"},
		{f: b, name: "builder"},
	}, nil
}

func prebuild(ctx context.Context, workers int, br []builderResolver) error {
	jobs := make(chan builder, 1)

	var wg sync.WaitGroup

	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case b, ok := <-jobs:
					if !ok {
						return
					}

					b.Build()
				}
			}
		}()
	}

	go func() {
		defer close(jobs)

		for _, r := range br {
			for _, b := range r.Builders() {
				select {
				case <-ctx.Done():
					return
				case jobs <- b:
				}
			}
		}
	}()

	wg.Wait()

	// if context cancelled return early
	return ctx.Err()
}
