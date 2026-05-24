package jen

import (
	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderCandidateInterface() {
	r.f.Add(block(
		Type().Id(r.naming.Candidate.Interface).Interface(
			r.generateCandidateCandidateMethodSignature(),
			r.generateCandidateResolverMethodSignature(),
		),
	))
}

func (r *Renderer) generateCandidateCandidateMethodSignature() Code {
	return Id(r.naming.Candidate.CandidateMethod).Params().Params(libCandidate)
}

func (r *Renderer) generateCandidateRelationCandidateMethodBody(rcv, builderRcv Code) []Code {
	return []Code{
		Return(Add(libCandidate).Values(
			Id("ID").Op(":").Add(builderRcv).Dot(r.naming.Builder.FieldID),
			Id("SubID").Op(":").Add(rcv).Dot(r.naming.Candidate.FieldSubID),
			Id("ParentID").Op(":").Add(builderRcv).Dot(r.naming.Builder.FieldParentID),
			Id("HasChildren").Op(":").Add(builderRcv).Dot(r.naming.Builder.FieldRelations).Op("!=").Lit(0),
		)),
	}
}

func (r *Renderer) generateCandidateNestedCandidateMethodBody(rcv, builderRcv Code) []Code {
	return []Code{
		Return(Add(libCandidate).Values(
			Id("ID").Op(":").Add(builderRcv).Dot(r.naming.Builder.FieldID),
			Id("SubID").Op(":").Add(rcv).Dot(r.naming.Candidate.FieldSubID),
			Id("ParentID").Op(":").Add(builderRcv).Dot(r.naming.Builder.FieldParentID),
			Id("HasChildren").Op(":").Add(builderRcv).Dot(r.naming.Builder.FieldNested).Op("!=").Lit(0),
			Id("Nested").Op(":").Add(True()),
		)),
	}
}

func (r *Renderer) generateCandidateCandidateMethodCall(rcv Code) Code {
	return Add(rcv).Dot(r.naming.Candidate.CandidateMethod).Call()
}

func (r *Renderer) generateCandidateResolverMethodSignature() Code {
	return Id(r.naming.Candidate.ResolverMethod).Params().Params(Id(r.naming.Resolver.Type))
}

func (r *Renderer) generateCandidateRelationsResolverMethodBody(erid plan.EntityResolverID, rcv Code) []Code {
	return []Code{
		Return(r.generateResolverEntityConstructorCall(
			erid,
			Add(rcv).Dot(r.naming.Builder.FieldID),
			Add(rcv).Dot(r.naming.Builder.FieldParentID),
			Add(rcv).Dot(r.naming.Builder.FieldRelations),
		)),
	}
}

func (r *Renderer) generateCandidateNestedResolverMethodBody(nrid plan.NestedResolverID, rcv Code) []Code {
	return []Code{
		Return(r.generateResolverNestedConstructorCall(
			nrid,
			Add(rcv).Dot(r.naming.Builder.FieldID),
			Add(rcv).Dot(r.naming.Builder.FieldParentID),
			Add(rcv).Dot(r.naming.Builder.FieldNested),
		)),
	}
}

func (r *Renderer) generateCandidateResolverMethodCall(rcv Code) Code {
	return Add(rcv).Dot(r.naming.Candidate.ResolverMethod).Call()
}
