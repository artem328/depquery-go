package jen

import (
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderPlanInterface() {
	typeParamRoot := Id("R")

	r.f.Add(block(
		Type().Id(r.naming.Plan.Interface).Types(Add(typeParamRoot).Any()).Interface(
			r.generatePlanNewInstanceMethodSignature(typeParamRoot, Null(), Null(), Null()),
		),
	))
}

func (r *Renderer) generatePlanInterfaceType(typeParamRoot Code) Code {
	return Id(r.naming.Plan.Interface).Types(typeParamRoot)
}

func (r *Renderer) generatePlanNewInstanceMethodSignature(typeParamRoot, iterParam, prefetchResolverParam, entityPrefetcher Code) Code {
	return Id(r.naming.Plan.NewInstanceMethod).Params(Add(iterParam, iterSeq).Types(typeParamRoot), Add(prefetchResolverParam).Id(r.naming.PrefetchResolver.Interface), Add(entityPrefetcher).Id(r.naming.EntityPrefetcher.Interface)).Params(Id(r.naming.Instance.Interface))
}

func (r *Renderer) renderPlanImplementation() {
	typeParamRoot := Id("R")

	r.renderPlanStruct(typeParamRoot)

	rcv := Id("p")

	r.renderPlanStructNewInstanceMethod(rcv, typeParamRoot)
}

func (r *Renderer) renderPlanStruct(typeParamRoot Code) {
	r.f.Add(block(
		Type().Id(r.naming.Plan.Struct).Types(Add(typeParamRoot).Any()).Struct(
			Id(r.naming.Plan.FieldFetchContextConstructor).Add(r.generateFetchContextConstructorForEntityType(typeParamRoot)),
			Id(r.naming.Plan.FieldResolvers).Index().Index().Id(r.naming.Resolver.Type),
		),
	))
}

func (r *Renderer) generatePlanStructMethodBase(rcv, typeParamRoot Code) Code {
	return Func().Params(Add(rcv).Id(r.naming.Plan.Struct).Types(typeParamRoot))
}

func (r *Renderer) renderPlanStructNewInstanceMethod(rcv, typeParamRoot Code) {
	iter := Id("i")
	prefetchResolver := Id("pr")
	entityPrefetcher := Id("ep")

	r.f.Add(block(
		Add(r.generatePlanStructMethodBase(rcv, typeParamRoot), r.generatePlanNewInstanceMethodSignature(typeParamRoot, iter, prefetchResolver, entityPrefetcher)).Block(
			Return(r.generateInstanceStructInit(
				r.generateFetchContextConstructorForEntityCall(Add(rcv).Dot(r.naming.Plan.FieldFetchContextConstructor), iter),
				Add(rcv).Dot(r.naming.Plan.FieldResolvers),
				prefetchResolver,
				entityPrefetcher,
			)),
		),
	))
}

func (r *Renderer) generatePlanStructInit(typeParamRoot, ctxConstructor, resolvers Code) Code {
	return Id(r.naming.Plan.Struct).Types(typeParamRoot).Add(valuesMultiline(
		Id(r.naming.Plan.FieldFetchContextConstructor).Op(":").Add(ctxConstructor),
		Id(r.naming.Plan.FieldResolvers).Op(":").Add(resolvers),
	))
}
