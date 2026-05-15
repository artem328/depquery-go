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

func (r *Renderer) generateCandidateCandidateMethodBody(rcv Code) []Code {
	return []Code{
		Return(Add(libCandidate).Values(
			Id("ID").Op(":").Add(rcv).Dot(r.naming.Builder.FieldID),
			Id("ParentID").Op(":").Add(rcv).Dot(r.naming.Builder.FieldParentID),
			Id("Relations").Op(":").Add(rcv).Dot(r.naming.Builder.FieldRelations),
		)),
	}
}

func (r *Renderer) generateCandidateCandidateMethodCall(rcv Code) Code {
	return Add(rcv).Dot(r.naming.Candidate.CandidateMethod).Call()
}

func (r *Renderer) generateCandidateResolverMethodSignature() Code {
	return Id(r.naming.Candidate.ResolverMethod).Params().Params(Id(r.naming.Resolver.Type))
}

func (r *Renderer) generateCandidateResolverMethodBody(erid plan.EntityResolverID, rcv Code) []Code {
	return []Code{
		Return(r.generateResolverConstructorCall(
			erid,
			Add(rcv).Dot(r.naming.Builder.FieldID),
			Add(rcv).Dot(r.naming.Builder.FieldParentID),
			Add(rcv).Dot(r.naming.Builder.FieldRelations),
		)),
	}
}

func (r *Renderer) generateCandidateResolverMethodCall(rcv Code) Code {
	return Add(rcv).Dot(r.naming.Candidate.ResolverMethod).Call()
}
