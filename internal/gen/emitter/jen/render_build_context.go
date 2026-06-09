package jen

import (
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderBuildContext() {
	r.renderBuildContextStruct()
	r.renderBuildContextConstructor()

	rcv := Id("ctx")

	r.renderBuildContextNextIDMethod(rcv)
	r.renderBuildContextAddCandidateMethod(rcv)
	r.renderBuildContextPlanMethod(rcv)
}

func (r *Renderer) renderBuildContextStruct() {
	r.f.Add(block(
		Type().Id(r.naming.BuildContext.Type).Struct(
			Id(r.naming.BuildContext.FieldCandidates).Index().Id(r.naming.Candidate.Interface),
			Id(r.naming.BuildContext.FieldCandidatesByID).Add(r.generateBuildContextFieldCandidatesByIDType()),
			Id(r.naming.BuildContext.FieldLastId).Add(libID),
		),
	))
}

func (r *Renderer) renderBuildContextConstructor() {
	r.f.Add(block(
		Func().Id(r.naming.BuildContext.Constructor).Params().Params(Op("*").Id(r.naming.BuildContext.Type)).Block(
			Return(Op("&").Id(r.naming.BuildContext.Type).Values(Id(r.naming.BuildContext.FieldCandidatesByID).Op(":").Make(r.generateBuildContextFieldCandidatesByIDType()))),
		),
	))
}

func (r *Renderer) generateBuildContextConstructorCall() Code {
	return Id(r.naming.BuildContext.Constructor).Call()
}

func (r *Renderer) generateBuildContextFieldCandidatesByIDType() Code {
	return Map(libID).Id(r.naming.Candidate.Interface)
}

func (r *Renderer) generateBuildContextMethodBase(rcv Code) Code {
	return Func().Params(Add(rcv).Op("*").Id(r.naming.BuildContext.Type))
}

func (r *Renderer) renderBuildContextNextIDMethod(rcv Code) {
	r.f.Add(block(
		Add(r.generateBuildContextMethodBase(rcv)).Id(r.naming.BuildContext.NextIDMethod).Params().Params(libID).Block(
			Add(rcv).Dot(r.naming.BuildContext.FieldLastId).Op("++"),
			Empty(),
			Return(Add(rcv).Dot(r.naming.BuildContext.FieldLastId)),
		),
	))
}

func (r *Renderer) generateBuildContextNextIDMethodCall(rcv Code) Code {
	return Add(rcv).Dot(r.naming.BuildContext.NextIDMethod).Call()
}

func (r *Renderer) renderBuildContextAddCandidateMethod(rcv Code) {
	id := Id("id")
	candidate := Id("c")

	candidates := Add(rcv).Dot(r.naming.BuildContext.FieldCandidates)

	r.f.Add(block(
		Add(r.generateBuildContextMethodBase(rcv)).Id(r.naming.BuildContext.AddCandidateMethod).Params(
			Add(id).Add(libID),
			Add(candidate).Id(r.naming.Candidate.Interface),
		).Block(
			appendSlice(candidates, candidate),
			Add(rcv).Dot(r.naming.BuildContext.FieldCandidatesByID).Index(id).Op("=").Add(candidate),
		),
	))
}

func (r *Renderer) generateBuildContextAddCandidateMethodCall(rcv, id, candidate Code) Code {
	return Add(rcv).Dot(r.naming.BuildContext.AddCandidateMethod).Call(id, candidate)
}

func (r *Renderer) renderBuildContextPlanMethod(rcv Code) {
	planner := Id("p")

	r.f.Add(block(
		Add(r.generateBuildContextMethodBase(rcv)).Id(r.naming.BuildContext.PlanMethod).Params(
			Add(planner).Add(libPlanner),
		).Params(
			Index().Index().Id(r.naming.Resolver.Type),
		).BlockFunc(func(g *Group) {
			inputCandidates := Id("cc")

			g.Add(inputCandidates).Op(":=").Make(Index().Add(libCandidate), Len(Add(rcv).Dot(r.naming.BuildContext.FieldCandidates)))

			i := Id("i")
			candidate := Id("c")
			g.For(List(i, candidate).Op(":=").Range().Add(rcv).Dot(r.naming.BuildContext.FieldCandidates)).Block(
				Add(inputCandidates).Index(i).Op("=").Add(r.generateCandidateCandidateMethodCall(candidate)),
			)

			g.Empty()

			planned := Id("planned")
			g.Add(planned).Op(":=").Add(planner).Dot("Plan").Call(inputCandidates)

			resolvers := Id("resolvers")
			g.Add(resolvers).Op(":=").Make(Index().Index().Id(r.naming.Resolver.Type), Len(planned))

			j := Id("j")
			level := Id("level")
			g.For(List(i, level).Op(":=").Range().Add(planned)).Block(
				Add(resolvers).Index(i).Op("=").Make(Index().Id(r.naming.Resolver.Type), Len(level)),
				For(List(j, candidate).Op(":=").Range().Add(level)).Block(
					Add(resolvers).Index(i).Index(j).Op("=").Add(r.generateCandidateResolverMethodCall(Add(rcv).Dot(r.naming.BuildContext.FieldCandidatesByID).Index(Add(candidate).Dot("ID")))),
				),
			)

			g.Empty()

			g.Return(resolvers)
		}),
	))
}

func (r *Renderer) generateBuildContextPlanMethodCall(rcv, planner Code) Code {
	return Add(rcv).Dot(r.naming.BuildContext.PlanMethod).Call(planner)
}
