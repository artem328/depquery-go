package jen

import (
	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderEntityPrefetcherInterface() {
	r.f.Add(block(
		Type().Id(r.naming.EntityPrefetcher.Interface).Interface(r.generateEntityPrefetcherMethodSignatures()...),
	))
}

func (r *Renderer) generateEntityPrefetcherMethodSignatures() []Code {
	var methods []Code

	for _, pm := range r.plan.PrefetchMethods {
		methods = append(methods, r.generateEntityPrefetcherMethodSignature(pm))
	}

	return methods
}

func (r *Renderer) generateEntityPrefetcherMethodSignature(pm plan.PrefetchMethod) Code {
	e := r.plan.Model.Entities[pm.Entity]
	im := r.plan.Model.Members[e.IDMember]

	idType := r.types[im.Type]
	if pm.Reversed {
		rev := r.plan.Model.Entities[pm.ReversedByEntity]
		revim := r.plan.Model.Members[rev.IDMember]
		idType = r.types[revim.Type]
	}

	return Id(r.naming.EntityPrefetcher.Method[pm.ID]).Params(contextContext, Index().Add(idType)).Params(Add(iterSeq).Types(r.types[e.Type]), Error())
}

func (r *Renderer) generateEntityPrefetcherMethodCall(pmid plan.PrefetchMethodID, rcv, contextArg, idsArg Code) Code {
	return Add(rcv).Dot(r.naming.EntityPrefetcher.Method[pmid]).Call(contextArg, idsArg)
}
