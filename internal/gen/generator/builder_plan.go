package generator

import . "github.com/dave/jennifer/jen"

func (i *nameIndex) PlanInterface() string {
	return "Plan"
}

func (i *nameIndex) PlanStruct() string {
	return "plan"
}

func (i *nameIndex) PlanNewInstanceMethod() string {
	return "NewInstance"
}

type planBuilder struct {
	naming *nameIndex
	iface  Code
	impl   Code
}

func newPlanBuilder(naming *nameIndex) *planBuilder {
	return &planBuilder{naming: naming}
}

func (b *planBuilder) Builders() []builder {
	return []builder{builderFunc(b.buildIface), builderFunc(b.buildImplementation)}
}

func (b *planBuilder) Interface() Code {
	return b.iface
}

func (b *planBuilder) Implementation() Code {
	return b.impl
}

func (b *planBuilder) buildIface() {
	b.iface = Type().Id(b.naming.PlanInterface()).Types(Id("R").Any()).Interface(
		b.newInstanceMethodSignature("R", "", "", ""),
	)
}

func (b *planBuilder) buildImplementation() {
	b.impl = Do(func(s *Statement) {
		s.Type().Id(b.naming.PlanStruct()).Types(Id("R").Any()).Struct(
			Id("ctxConstructor").Func().Params(iterSeq.Clone().Types(Id("R"))).Op("*").Id(b.naming.FetchContextStruct()),
			Id("resolvers").Index().Index().Id(b.naming.ResolverFunc()),
		)
		s.Line()

		s.Func().Params(Id("p").Id(b.naming.PlanStruct()).Types(Id("R"))).
			Add(b.newInstanceMethodSignature("R", "i", "pr", "ep")).
			Block(Return(
				Id(b.naming.InstanceStruct()).CustomFunc(multilineValuesOpts, func(s *Group) {
					s.Id("ctx").Op(":").Id("p").Dot("ctxConstructor").Call(Id("i"))
					s.Id("resolvers").Op(":").Id("p").Dot("resolvers")
					s.Id("resolver").Op(":").Id("pr")
					s.Id("prefetcher").Op(":").Id("ep")
				}),
			))
	})
}

func (b *planBuilder) newInstanceMethodSignature(typeParam, iterParam, resParam, prefParam string) Code {
	return Id(b.naming.PlanNewInstanceMethod()).Params(
		Id(iterParam).Add(iterSeq).Types(Id(typeParam)),
		Id(resParam).Id(b.naming.PrefetchResolverInterface()),
		Id(prefParam).Id(b.naming.EntityPrefetcherInterface()),
	).Id(b.naming.InstanceInterface())
}
