package generator

import (
	"sync"

	. "github.com/dave/jennifer/jen"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (i *nameIndex) ResolverFunc() string {
	return "resolver"
}

func (i *nameIndex) ResolverRelConst(e *schema.Entity, r schema.Relation) string {
	return i.getOrCreate(nKey("ResolverRelConst", e.Name, "", r.Name), func() string {
		return "rel" + sanitizeID(e.Name, sanitizeRawCapitalized) + sanitizeID(r.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) ResolverVariantRelConst(e *schema.Entity, v schema.EntityVariant, r schema.Relation) string {
	return i.getOrCreate(nKey("ResolverVariantRelConst", e.Name, v.Name, r.Name), func() string {
		return "rel" + sanitizeID(e.Name, sanitizeRawCapitalized) +
			"V" + sanitizeID(v.Name, sanitizeRawCapitalized) +
			sanitizeID(r.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) ResolverName(e *schema.Entity) string {
	return i.getOrCreate(nKey("ResolverName", e.Name, "", ""), func() string {
		return "resolve" + sanitizeID(e.Name, sanitizeRawCapitalized) + "Relations"
	})
}

type resolverBuilder struct {
	naming   *nameIndex
	entities []*schema.Entity
	implsMu  sync.Mutex
	impls    map[*schema.Entity][]Code
}

func newResolverBuilder(naming *nameIndex, entities []*schema.Entity) *resolverBuilder {
	return &resolverBuilder{naming: naming, entities: entities, impls: make(map[*schema.Entity][]Code)}
}

func (b *resolverBuilder) Builders() []builder {
	builders := make([]builder, 0, len(b.entities))

	for _, e := range b.entities {
		builders = append(builders, builderFunc(func() {
			b.buildImplementation(e)
		}))
	}

	return builders
}

func (b *resolverBuilder) Implementation() Code {
	return Do(func(s *Statement) {
		s.Type().Id(b.naming.ResolverFunc()).Func().
			Params(Op("*").Id(b.naming.FetchContextStruct()), Id(b.naming.PrefetchResolverInterface()))
		s.Line()

		for _, e := range b.entities {
			s.Add(b.impls[e]...)
		}
	})
}

func (b *resolverBuilder) buildImplementation(e *schema.Entity) {
	if c := b.entityConsts(e); c != nil {
		b.addImpl(e, Add(c).Line().Line())
	}

	if r := b.entityResolver(e); r != nil {
		b.addImpl(e, Add(r).Line().Line())
	}
}

func (b *resolverBuilder) addImpl(e *schema.Entity, code Code) {
	b.implsMu.Lock()
	defer b.implsMu.Unlock()
	b.impls[e] = append(b.impls[e], code)
}

func (b *resolverBuilder) entityConsts(e *schema.Entity) Code {
	//	const (
	//		<const> ...
	//	)
	added := false
	c := Const().DefsFunc(func(c *Group) {
		for _, r := range e.Relations {
			c.Add(b.relConst(b.naming.ResolverRelConst(e, r), !added))

			added = true
		}

		for _, v := range e.Variants {
			for _, r := range v.Relations {
				c.Add(b.relConst(b.naming.ResolverVariantRelConst(e, v, r), !added))

				added = true
			}
		}
	})

	if added {
		return c
	}

	return nil
}

//nolint:gocognit,funlen // needs refactoring maybe
func (b *resolverBuilder) entityResolver(e *schema.Entity) Code {
	//	func <resolver>(eid, pid depquery.ID, include uint64) <resolverFunc> {
	//		return func(ctx *<fetchCtx>, res <prefetchResolver>) {
	//			for e := range ctx.<getEntities>(pid) {
	//				<resolve> ...
	//
	//				if v, ok := e.(<variantType>); ok {
	//					<resolve> ...
	//				} ...
	//			}
	//		}
	//	}
	added := false
	eid := Id("eid")
	pid := Id("pid")
	include := Id("include")

	r := Func().Id(b.naming.ResolverName(e)).Params(
		Add(eid),
		Add(pid, libID),
		Add(include, Uint64()),
	).Id(b.naming.ResolverFunc()).BlockFunc(func(body *Group) {
		ctx := Id("ctx")
		res := Id("res")
		body.Return().Func().
			Params(ctx.Clone().Op("*").Id(b.naming.FetchContextStruct()), res.Clone().Id(b.naming.PrefetchResolverInterface())).
			BlockFunc(func(body *Group) {
				entity := Id("e")
				body.For(entity.Clone().Op(":=").Range().Add(ctx).Dot(b.naming.FetchContextGetMethod(e)).Call(pid).
					BlockFunc(func(g *Group) {
						for _, r := range e.Relations {
							if r.IsReversed() {
								g.Add(b.resolveReversed(
									"include",
									b.naming.ResolverRelConst(e, r),
									"ctx",
									b.naming.FetchContextAddPrefetchReverseMethod(r.Entity, e),
									"eid",
									entity.Clone().Add(memberToJen(e.ID)),
								))

								continue
							}

							g.Add(b.resolve(
								"include",
								b.naming.ResolverRelConst(e, r),
								"res",
								b.naming.PrefetchResolverMethodName(e, r),
								"e",
								"ctx",
								b.naming.FetchContextAddPrefetchMethod(r.Entity),
								"eid",
							))

							added = true
						}

						for _, v := range e.Variants {
							if len(v.Relations) == 0 {
								continue
							}

							g.If(List(Id("v"), Id("ok")).Op(":=").Add(entity).Assert(typeToJen(v.Type)), Id("ok")).BlockFunc(func(g *Group) {
								for _, r := range v.Relations {
									if r.IsReversed() {
										g.Add(b.resolveReversed(
											"include",
											b.naming.ResolverVariantRelConst(e, v, r),
											"ctx",
											b.naming.FetchContextAddPrefetchReverseMethod(r.Entity, e),
											"eid",
											entity.Clone().Add(memberToJen(e.ID)),
										))

										continue
									}

									g.Add(b.resolve(
										"include",
										b.naming.ResolverVariantRelConst(e, v, r),
										"res",
										b.naming.PrefetchResolverVariantMethodName(e, v, r),
										"v",
										"ctx",
										b.naming.FetchContextAddPrefetchMethod(r.Entity),
										"eid",
									))
								}
							})

							added = true
						}
					}))
			})
	})
	if added {
		return r
	}

	return nil
}

func (b *resolverBuilder) relConst(name string, first bool) Code {
	// <name>
	//
	// or if first
	//
	// <name> uint64 = 1 << iota
	return Do(func(s *Statement) {
		s.Id(name)

		if first {
			s.Uint64().Op("=").Lit(1).Op("<<").Iota()
		}
	})
}

func (b *resolverBuilder) resolve(
	includeParam, constName string,
	resolverParam, resolveMethod, entityParam string,
	ctxParam, addPrefetchMethod, eidParam string,
) Code {
	//	if <includeParam>&<constName> != 0 {
	//		for r := range <resolverParam>.<resolveMethod>(<entityParam>) {
	//			<ctxParam>.<addPrefetchMethod>(r, <eidParam>)
	//		}
	//	}
	return If(Id(includeParam).Op("&").Id(constName).Op("!=").Lit(0)).BlockFunc(func(g *Group) {
		g.For(Id("r").Op(":=").Range().Id(resolverParam).Dot(resolveMethod).Call(Id(entityParam))).Block(
			Id(ctxParam).Dot(addPrefetchMethod).Call(Id("r"), Id(eidParam)),
		)
	})
}

func (b *resolverBuilder) resolveReversed(
	includeParam, constName string,
	ctxParam, addPrefetchReverseMethod, eidParam string,
	idMember Code,
) Code {
	//	if <includeParam>&<constName> != 0 {
	//		<ctxParam>.<addPrefetchReverseMethod>(<idMember>, <eidParam>)
	//	}
	return If(Id(includeParam).Op("&").Id(constName).Op("!=").Lit(0)).Block(
		Id(ctxParam).Dot(addPrefetchReverseMethod).Call(idMember, Id(eidParam)),
	)
}
