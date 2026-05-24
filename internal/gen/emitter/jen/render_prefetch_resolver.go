package jen

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderPrefetchResolverInterface() {
	r.f.Add(block(
		Type().Id(r.naming.PrefetchResolver.Interface).Interface(r.generatePrefetchResolverMethodSignatures()...),
	))
}

func (r *Renderer) generatePrefetchResolverMethodSignatures() []Code {
	signatures := make([]Code, 0, len(r.plan.Relations))

	for _, rm := range r.plan.ResolveMethods {
		signatures = append(signatures, r.generatePrefetchResolverMethodSignature(rm))
	}

	return signatures
}

func (r *Renderer) generatePrefetchResolverMethodSignature(rm plan.ResolveMethod) Code {
	var (
		argType    Code
		returnType Code
	)

	switch rmm := rm.(type) {
	case plan.RelationResolveMethod:
		rel := r.plan.Relations[rmm.Relation]
		f := r.plan.Model.Entities[rel.From]
		t := r.plan.Model.Entities[rel.To]
		im := r.plan.Model.Members[t.IDMember]

		argType = r.types[f.Type]
		if rel.Variant.Set {
			argType = Op("*").Add(r.types[r.plan.Model.Variants[rel.Variant.V].Type]) // TODO: probably variant type should be always a pointer in semantic model
		}
		returnType = Add(iterSeq).Types(r.types[im.Type])
	case plan.NestedResolveMethod:
		n := r.plan.Nesteds[rmm.Nested]
		f := r.plan.Model.Entities[n.From]
		t := r.plan.Model.Entities[n.To]

		if t.Synthetic {
			returnType = Add(iterSeq2).Types(Int(), r.types[t.Type])
		} else {
			returnType = Add(iterSeq).Types(r.types[t.Type])
		}

		argType = r.types[f.Type]
	default:
		panic(fmt.Errorf("unknown resolve method type: %T", rm))
	}

	return Id(r.naming.PrefetchResolver.Method[rm.GetID()]).Params(argType).Params(returnType)
}

func (r *Renderer) generatePrefetchResolverMethodCall(rmid plan.ResolveMethodID, rcv, entityArg Code) Code {
	return Add(rcv).Dot(r.naming.PrefetchResolver.Method[rmid]).Call(entityArg)
}
