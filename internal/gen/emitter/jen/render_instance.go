package jen

import (
	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderInstanceInterface() {
	r.f.Add(block(
		Type().Id(r.naming.Instance.Interface).Interface(
			r.generateInstanceResolveMethodSignature(Null()),
		),
	))
}

func (r *Renderer) generateInstanceResolveMethodSignature(ctxParam Code) Code {
	return Id(r.naming.Instance.ResolveMethod).Params(Add(ctxParam, contextContext)).Params(Id(r.naming.State.Interface), Error())
}

func (r *Renderer) renderInstanceImplementation() {
	r.renderInstanceDefaultConfig()
	r.renderInstanceStruct()

	rcv := Id("i")

	r.renderInstanceResolveMethodImplementation(rcv)
	r.renderInstancePrefetchMethodImplementation(rcv)
}

func (r *Renderer) renderInstanceDefaultConfig() {
	r.f.Add(block(
		Var().Id(r.naming.Instance.DefaultConfig).Op("=").Add(libInstanceConfig).Add(valuesMultiline(
			Id("Executor").Op(":").Add(libConcurrentExecutor).Values(),
		)),
	))
}

func (r *Renderer) renderInstanceStruct() {
	r.f.Add(block(
		Type().Id(r.naming.Instance.Struct).Struct(
			Id(r.naming.Instance.FieldFetchContext).Op("*").Id(r.naming.FetchContext.Struct),
			Id(r.naming.Instance.FieldResolvers).Index().Index().Id(r.naming.Resolver.Type),
			Id(r.naming.Instance.FieldPrefetchResolver).Id(r.naming.PrefetchResolver.Interface),
			Id(r.naming.Instance.FieldEntityPrefetcher).Id(r.naming.EntityPrefetcher.Interface),
			Id(r.naming.Instance.FieldConfig).Add(libInstanceConfig),
		),
	))
}

func (r *Renderer) generateInstanceStructInit(fetchContext, resolvers, prefetchResolver, entityPrefetcher, config Code) Code {
	return Id(r.naming.Instance.Struct).Add(valuesMultiline(
		Id(r.naming.Instance.FieldFetchContext).Op(":").Add(fetchContext),
		Id(r.naming.Instance.FieldResolvers).Op(":").Add(resolvers),
		Id(r.naming.Instance.FieldPrefetchResolver).Op(":").Add(prefetchResolver),
		Id(r.naming.Instance.FieldEntityPrefetcher).Op(":").Add(entityPrefetcher),
		Id(r.naming.Instance.FieldConfig).Op(":").Add(config),
	))
}

func (r *Renderer) generateInstanceMethodBase(rcv Code) Code {
	return Func().Params(Add(rcv).Id(r.naming.Instance.Struct))
}

func (r *Renderer) renderInstanceResolveMethodImplementation(rcv Code) {
	ctx := Id("ctx")
	level := Id("level")
	resolver := Id("r")
	swallow := Id("_")

	r.f.Add(block(
		Add(r.generateInstanceMethodBase(rcv), r.generateInstanceResolveMethodSignature(ctx)).Block(
			For(List(swallow, level).Op(":=").Range().Add(rcv).Dot(r.naming.Instance.FieldResolvers)).Block(
				For(List(swallow, resolver).Op(":=").Range().Add(level)).Block(
					r.generateResolverEntityCall(resolver, Add(rcv).Dot(r.naming.Instance.FieldFetchContext), Add(rcv).Dot(r.naming.Instance.FieldPrefetchResolver)),
				),
				Empty(),
				If(Err().Op(":=").Add(r.generateInstancePrefetchMethodCall(rcv, ctx)), Err().Op("!=").Nil()).Block(
					Return(Nil(), Err()),
				),
			),
			Empty(),
			Return(Add(rcv).Dot(r.naming.Instance.FieldFetchContext).Dot(r.naming.State.Struct), Nil()),
		),
	))
}

func (r *Renderer) renderInstancePrefetchMethodImplementation(rcv Code) {
	ctx := Id("ctx")

	r.f.Add(block(
		Add(r.generateInstanceMethodBase(rcv)).Id(r.naming.Instance.PrefetchMethod).Params(Add(ctx, contextContext)).Params(Error()).Block(
			r.generateInstancePrefetchMethodBody(rcv, ctx)...,
		),
	))
}

func (r *Renderer) generateInstancePrefetchMethodBody(rcv, ctx Code) []Code {
	prefetchTasks := Id("prefetch")
	processTasks := Id("process")
	processTask := Id("task")

	var statements []Code

	statements = append(statements,
		Defer().Add(r.generateFetchContextFlushMethodCall(Add(rcv).Dot(r.naming.Instance.FieldFetchContext))),
		Empty(),
		Add(prefetchTasks).Op(":=").Make(Index().Add(libTask), Lit(0), Lit(len(r.plan.EntityFetches))),
		Add(processTasks).Op(":=").Make(Index().Func().Params(), Lit(0), Lit(len(r.plan.EntityFetches))),
	)

	for _, ef := range r.plan.EntityFetches {
		statements = append(statements, Empty(), r.generateInstancePrefetchBlock(ef, rcv, prefetchTasks, processTasks))
	}

	statements = append(statements,
		Empty(),
		If(Err().Op(":=").Add(rcv).Dot(r.naming.Instance.FieldConfig).Dot("Executor").Dot("Execute").Call(ctx, prefetchTasks), Err().Op("!=").Nil()).Block(
			Return(Err()),
		),
		Empty(),
		For(List(Id("_"), processTask).Op(":=").Range().Add(processTasks)).Block(
			Add(processTask).Call(),
		),
		Empty(),
		Return(Nil()),
	)

	return statements
}

func (r *Renderer) generateInstancePrefetchBlock(ef plan.EntityFetch, rcv, prefetchTasks, processTasks Code) Code {
	e := r.plan.Model.Entities[ef.Entity]

	pending := Id("p")
	entities := Id("ee")
	entity := Id("e")
	ctx := Id("ctx")

	return If(Add(pending).Op(":=").Add(rcv).Dot(r.naming.Instance.FieldFetchContext).Dot(r.naming.FetchContext.FieldPending[ef.Child]), Len(pending).Op(">").Lit(0)).Block(
		Var().Add(entities).Add(iterSeq).Types(r.types[e.Type]),
		appendSlice(prefetchTasks, Func().Params(Add(ctx, contextContext)).Params(Err().Error()).Block(
			List(entities, Err()).Op("=").Add(r.generateEntityPrefetcherMethodCall(ef.PrefetchMethod, Add(rcv).Dot(r.naming.Instance.FieldEntityPrefetcher), ctx, pending)),
			Empty(),
			Return(Err()),
		)),
		appendSlice(processTasks, Func().Params().Block(
			For(Add(entity).Op(":=").Range().Add(entities)).Block(
				r.generateStateAdderMethodCall(ef.StateContainer, Add(rcv).Dot(r.naming.Instance.FieldFetchContext), entity),
			),
		)),
	)
}

func (r *Renderer) generateInstancePrefetchMethodCall(rcv, ctxArg Code) Code {
	return Add(rcv).Dot(r.naming.Instance.PrefetchMethod).Call(ctxArg)
}
