package generator

import . "github.com/dave/jennifer/jen"

func (i *nameIndex) BuildContextCandidateInterface() string {
	return "candidate"
}

func (i *nameIndex) BuildContextCandidateCandidateMethod() string {
	return "Candidate"
}

func (i *nameIndex) BuildContextCandidateResolverMethod() string {
	return "Resolver"
}

func (i *nameIndex) BuildContextStruct() string {
	return "buildContext"
}

func (i *nameIndex) BuildContextConstructor() string {
	return "newBuildContext"
}

func (i *nameIndex) BuildContextNextIDMethod() string {
	return "nextID"
}

func (i *nameIndex) BuildContextAddCandidateMethod() string {
	return "addCandidate"
}

func (i *nameIndex) BuildContextPlanMethod() string {
	return "plan"
}

type buildContextBuilder struct {
	naming *nameIndex
	impl   Code
}

func newBuildContextBuilder(naming *nameIndex) *buildContextBuilder {
	return &buildContextBuilder{naming: naming}
}

func (b *buildContextBuilder) Builders() []builder {
	return []builder{builderFunc(b.buildImplementation)}
}

func (b *buildContextBuilder) Implementation() Code {
	return b.impl
}

//nolint:funlen // worth refactoring
func (b *buildContextBuilder) buildImplementation() {
	rcv := Id("ctx").Op("*").Id(b.naming.BuildContextStruct())

	b.impl = Add(
		Type().Id(b.naming.BuildContextCandidateInterface()).Interface(
			Id(b.naming.BuildContextCandidateCandidateMethod()).Params().Add(libCandidate),
			Id(b.naming.BuildContextCandidateResolverMethod()).Params().Id(b.naming.ResolverFunc()),
		),
		Line().Line(),

		Type().Id(b.naming.BuildContextStruct()).Struct(
			Id("candidates").Index().Id(b.naming.BuildContextCandidateInterface()),
			Id("candidatesByID").Map(libID).Id(b.naming.BuildContextCandidateInterface()),
			Id("lastID").Add(libID),
		),
		Line().Line(),

		Func().Id(b.naming.BuildContextConstructor()).Params().Op("*").Id(b.naming.BuildContextStruct()).
			Block(Return(Op("&").Id(b.naming.BuildContextStruct()).Values(
				Id("candidatesByID").Op(":").Make(Map(libID).Id(b.naming.BuildContextCandidateInterface())),
			))),
		Line().Line(),

		Func().Params(rcv).Id(b.naming.BuildContextNextIDMethod()).Params().Add(libID).Block(
			Id("ctx").Dot("lastID").Op("++").Line(),
			Return(Id("ctx").Dot("lastID")),
		),
		Line().Line(),

		Func().Params(rcv).Id(b.naming.BuildContextAddCandidateMethod()).Params(
			Id("id").Add(libID),
			Id("c").Id(b.naming.BuildContextCandidateInterface()),
		).
			Block(
				Id("ctx").Dot("candidates").Op("=").Append(Id("ctx").Dot("candidates"), Id("c")),
				Id("ctx").Dot("candidatesByID").Index(Id("id")).Op("=").Id("c"),
			),
		Line().Line(),

		//	func (<rcv>) <plan>(p depquery.Planner) [][]<resolver> {
		//		cc := make([]depquery.Candidate, len(ctx.candidates))
		//		for i, c := range ctx.candidates {
		//			cc[i] = c.<candidateMethod>()
		//		}
		//
		//		candidates := p.Plan(cc)
		//		resolvers := make([][]<resolver>, len(candidates))
		//		for i, level := range candidates {
		//			resolvers[i] = make([]<resolver>, len(level)),
		//			for j, c := range level {
		//				resolvers[i][j] = ctx.candidatesByID[c.ID].<resolverMethod>()
		//			}
		//		}
		//
		//		return resolvers
		//	}
		Func().Params(rcv).Id(b.naming.BuildContextPlanMethod()).
			Params(Id("p").Qual(libPkg, "Planner")).
			Index().Index().Id(b.naming.ResolverFunc()).
			BlockFunc(func(body *Group) {
				cc := Id("cc")
				ctxCandidates := Id("ctx").Dot("candidates")
				body.Add(cc).Op(":=").Make(Index().Add(libCandidate), Len(ctxCandidates))

				i := Id("i")
				c := Id("c")
				body.For(List(i, c).Op(":=").Range().Add(ctxCandidates)).Block(
					cc.Clone().Index(i).Op("=").Add(c).Dot(b.naming.BuildContextCandidateCandidateMethod()).Call(),
				)
				body.Line()

				candidates := Id("candidates")
				body.Add(candidates).Op(":=").Id("p").Dot("Plan").Call(cc)

				resolvers := Id("resolvers")
				body.Add(resolvers).Op(":=").Make(Index().Index().Id(b.naming.ResolverFunc()), Len(candidates))

				level := Id("level")
				body.For(List(i, level).Op(":=").Range().Add(candidates)).BlockFunc(func(g *Group) {
					g.Add(resolvers).Index(i).Op("=").Make(Index().Id(b.naming.ResolverFunc()), Len(level))

					j := Id("j")
					g.For(List(j, c).Op(":=").Range().Add(level)).Block(
						resolvers.Clone().Index(i).Index(j).Op("=").
							Id("ctx").Dot("candidatesByID").Index(c.Clone().Dot("ID")).
							Dot(b.naming.BuildContextCandidateResolverMethod()).Call(),
					)
				})
				body.Line()

				body.Return(resolvers)
			}),
		Line(),
	)
}
