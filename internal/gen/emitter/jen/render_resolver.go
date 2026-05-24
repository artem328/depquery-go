package jen

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderResolvers() {
	r.renderResolverType()

	for _, er := range r.plan.EntityResolvers {
		r.renderResolverEntity(er)
	}

	for _, nr := range r.plan.NestedResolvers {
		r.renderResolverNested(nr)
	}
}

func (r *Renderer) renderResolverType() {
	r.f.Add(block(
		Type().Id(r.naming.Resolver.Type).Func().Add(r.generateResolverFuncSignature(Null(), Null())),
	))
}

func (r *Renderer) generateResolverFuncSignature(fetchContextParam, prefetchResolverParam Code) Code {
	return Params(Add(fetchContextParam).Op("*").Id(r.naming.FetchContext.Struct), Add(prefetchResolverParam).Id(r.naming.PrefetchResolver.Interface))
}

func (r *Renderer) renderResolverEntity(er plan.EntityResolver) {
	r.renderResolverEntityOptionConstants(er)
	r.renderResolverEntityConstructor(er)
}

func (r *Renderer) generateResolverEntityCall(resolver, fetchContext, prefetchResolver Code) Code {
	return Add(resolver).Call(fetchContext, prefetchResolver)
}

func (r *Renderer) renderResolverEntityOptionConstants(er plan.EntityResolver) {
	consts := make([]Code, 0, len(er.Resolutions))

	for i, rid := range r.flattenResolverRelations(er.Resolutions) {
		c := Id(r.naming.Relation.ConstantName[rid])

		if i == 0 {
			c.Uint64().Op("=").Lit(1).Op("<<").Iota()
		}

		consts = append(consts, c)
	}

	r.f.Add(block(
		Const().Defs(consts...),
	))
}

func (r *Renderer) flattenResolverRelations(resolutions []plan.EntityResolution) []plan.RelationID {
	relations := make([]plan.RelationID, 0, len(resolutions))

	for _, res := range resolutions {
		switch rr := res.(type) {
		case plan.EntityResolutionRelation:
			relations = append(relations, rr.Relation)
		case plan.EntityResolutionVariant:
			relations = append(relations, r.flattenResolverRelations(rr.Resolutions)...)
		default:
			panic(fmt.Sprintf("unknown resolution type: %T", rr))
		}
	}

	return relations
}

func (r *Renderer) renderResolverEntityConstructor(er plan.EntityResolver) {
	pfg := r.plan.ParentFetchGetters[er.ParentFetchGetter]

	entityLevelID := Id("eid")
	parentLevelID := Id("pid")
	include := Id("include")
	fetchContext := Id("ctx")
	prefetchResolver := Id("pr")
	entity := Id("e")

	var iterVars []Code
	if pfg.Synthetic {
		iterVars = []Code{Id("_"), entity}
	} else {
		iterVars = []Code{entity}
	}

	r.f.Add(block(
		Func().Id(r.naming.Resolver.EntityConstructor[er.ID]).Params(Add(entityLevelID), Add(parentLevelID, libID), Add(include, Uint64())).Params(Id(r.naming.Resolver.Type)).Block(
			Return(
				Func().Add(r.generateResolverFuncSignature(fetchContext, prefetchResolver)).Block(
					For(List(iterVars...).Op(":=").Range().Add(r.generateFetchContextParentGetterMethodCall(er.ParentFetchGetter, fetchContext, parentLevelID))).Block(
						r.generateResolverEntityResolveBlocks(er, fetchContext, prefetchResolver, entity, include, entityLevelID)...,
					),
				),
			),
		),
	))
}

func (r *Renderer) generateResolverEntityResolveBlocks(er plan.EntityResolver, fetchContext, prefetchResolver, entity, include, entityLevelID Code) []Code {
	var blocks []Code

	for _, res := range er.Resolutions {
		blocks = append(blocks, r.generateResolverEntityResolveBlock(res, fetchContext, prefetchResolver, entity, include, entityLevelID)...)
	}

	return blocks
}

func (r *Renderer) generateResolverEntityResolveBlock(er plan.EntityResolution, fetchContext, prefetchResolver, entity, include, entityLevelID Code) []Code {
	switch err := er.(type) {
	case plan.EntityResolutionRelation:
		return r.generateResolverEntityResolutionRelationBlock(err, fetchContext, prefetchResolver, entity, include, entityLevelID)
	case plan.EntityResolutionVariant:
		entityVariant := Id("v")

		var blocks []Code
		for _, ier := range err.Resolutions {
			blocks = append(blocks, r.generateResolverEntityResolveBlock(ier, fetchContext, prefetchResolver, entityVariant, include, entityLevelID)...)
		}

		var statements []Code

		v := r.plan.Model.Variants[err.Variant]
		ptrType := Op("*").Add(r.types[v.Type])

		ok := Id("ok")

		statements = append(statements, List(entityVariant, ok).Op(":=").Add(entity).Assert(ptrType))
		if v.ValueAssignable {
			valueVariant := Id("vv")
			valueVariantOk := Id("vvOk")

			statements = append(statements,
				If(Op("!").Add(ok)).Block(
					If(List(valueVariant, valueVariantOk).Op(":=").Add(entity).Assert(r.types[v.Type]), valueVariantOk).Block(
						Add(entityVariant).Op("=").Op("&").Add(valueVariant),
						Add(ok).Op("=").Add(valueVariantOk),
					),
				),
			)
		}

		statements = append(statements, If(ok).Block(blocks...))

		return []Code{Block(statements...)}
	default:
		panic(fmt.Errorf("unknown resolution type: %T", err))
	}
}

func (r *Renderer) generateResolverEntityResolutionRelationBlock(err plan.EntityResolutionRelation, fetchContext, prefetchResolver, entity, include, entityLevelID Code) []Code {
	var code Code

	rel := r.plan.Relations[err.Relation]
	if rel.ReversedBy.Set {
		code = r.generateFetchContextEnqueueMethodCall(err.EntityFetch, fetchContext, r.members.Member(entity, r.plan.Model.Entities[rel.From].IDMember), entityLevelID)
	} else {
		relationId := Id("r")

		code = For(Add(relationId).Op(":=").Range().Add(r.generatePrefetchResolverMethodCall(err.ResolveMethod, prefetchResolver, entity))).Block(
			r.generateFetchContextEnqueueMethodCall(err.EntityFetch, fetchContext, relationId, entityLevelID),
		)
	}

	return []Code{
		If(r.generateResolverEntityIncludeCheck(err.Relation, include)).Block(code),
	}
}

func (r *Renderer) generateResolverEntityIncludeCheck(rid plan.RelationID, include Code) Code {
	return Add(include).Op("&").Id(r.naming.Relation.ConstantName[rid]).Op("!=").Lit(0)
}

func (r *Renderer) generateResolverEntityConstructorCall(erid plan.EntityResolverID, entityLevelIDArg, parentLevelIDArg, includeArg Code) Code {
	return Id(r.naming.Resolver.EntityConstructor[erid]).Call(entityLevelIDArg, parentLevelIDArg, includeArg)
}

func (r *Renderer) renderResolverNested(nr plan.NestedResolver) {
	r.renderResolverNestedOptionConstants(nr)
	r.renderResolverNestedConstructor(nr)
}

func (r *Renderer) renderResolverNestedOptionConstants(nr plan.NestedResolver) {
	consts := make([]Code, 0, len(nr.Resolutions))

	for i, res := range nr.Resolutions {
		c := Id(r.naming.Nested.ConstantName[res.Nested])

		if i == 0 {
			c.Uint64().Op("=").Lit(1).Op("<<").Iota()
		}

		consts = append(consts, c)
	}

	r.f.Add(block(
		Const().Defs(consts...),
	))
}

func (r *Renderer) renderResolverNestedConstructor(nr plan.NestedResolver) {
	pfg := r.plan.ParentFetchGetters[nr.ParentFetchGetter]

	entityLevelID := Id("eid")
	parentLevelID := Id("pid")
	include := Id("include")
	fetchContext := Id("ctx")
	prefetchResolver := Id("pr")
	rawParentID := Id("_id")
	syntheticParentID := Id("id")
	entity := Id("e")

	var (
		syntheticIDNamespace Code
		iterVars             []Code
		iterCode             []Code
	)
	if pfg.Synthetic {
		iterVars = []Code{syntheticParentID, entity}
	} else {
		e := r.plan.Model.Entities[nr.Entity]
		iterVars = []Code{entity}
		iterCode = append(iterCode,
			Add(rawParentID).Op(":=").Add(r.members.Member(entity, e.IDMember)),
			Add(syntheticParentID).Op(":=").Add(libSyntheticID).Call(Id(r.naming.FetchContext.SyntheticNamespaceConst), generateTypeToBytes(e.IDUnderlyingType, rawParentID)),
			Empty(),
		)
		syntheticIDNamespace = Const().Id(r.naming.FetchContext.SyntheticNamespaceConst).Op("=").Lit(nr.SyntheticIDNamespace)
	}

	iterCode = append(iterCode, r.generateResolverNestedResolveBlocks(nr, fetchContext, prefetchResolver, entity, include, syntheticParentID, entityLevelID)...)
	iterCode = append(iterCode,
		Empty(),
		Id("_").Op("=").Add(syntheticParentID).Comment("In case it wasn't ever used"),
	)

	var funcBody []Code

	if syntheticIDNamespace != nil {
		funcBody = append(funcBody, syntheticIDNamespace, Empty())
	}

	funcBody = append(funcBody, For(List(iterVars...).Op(":=").Range().Add(r.generateFetchContextParentGetterMethodCall(nr.ParentFetchGetter, fetchContext, parentLevelID))).Block(iterCode...))

	r.f.Add(block(
		Func().Id(r.naming.Resolver.NestedConstructor[nr.ID]).Params(entityLevelID, Add(parentLevelID, libID), Add(include, Uint64())).Params(Id(r.naming.Resolver.Type)).Block(
			Return(
				Func().Add(r.generateResolverFuncSignature(fetchContext, prefetchResolver)).Block(funcBody...),
			),
		),
	))
}

func (r *Renderer) generateResolverNestedResolveBlocks(nr plan.NestedResolver, fetchContext, prefetchResolver, entity, include, entityParentID, entityLevelID Code) []Code {
	var blocks []Code

	for _, res := range nr.Resolutions {
		blocks = append(blocks, r.generateResolverNestedResolveBlock(res, fetchContext, prefetchResolver, entity, include, entityParentID, entityLevelID)...)
	}

	return blocks
}

func (r *Renderer) generateResolverNestedResolveBlock(res plan.NestedResolution, fetchContext, prefetchResolver, entity, include, parentEntityID, entityLevelID Code) []Code {
	index := Id("i")
	nested := Id("n")

	var (
		iterVars   Code
		methodCall Code
	)
	if res.Synthetic {
		iterVars = List(index, nested)
		methodCall = r.generateFetchContextAddNestedSyntheticMethodCall(res.NestedEntityFetch, fetchContext, nested, entityLevelID, Add(libSyntheticID).Call(parentEntityID, Add(libIBytes).Call(index)))
	} else {
		iterVars = nested
		methodCall = r.generateFetchContextAddNestedMethodCall(res.NestedEntityFetch, fetchContext, nested, entityLevelID)
	}

	code := For(Add(iterVars).Op(":=").Range().Add(r.generatePrefetchResolverMethodCall(res.ResolveMethod, prefetchResolver, entity))).Block(methodCall)

	return []Code{
		If(r.generateResolverNestedIncludeCheck(res.Nested, include)).Block(code),
	}
}

func (r *Renderer) generateResolverNestedIncludeCheck(nid plan.NestedID, include Code) Code {
	return Add(include).Op("&").Id(r.naming.Nested.ConstantName[nid]).Op("!=").Lit(0)
}

func (r *Renderer) generateResolverNestedConstructorCall(erid plan.NestedResolverID, entityLevelIDArg, parentLevelIDArg, includeArg Code) Code {
	return Id(r.naming.Resolver.NestedConstructor[erid]).Call(entityLevelIDArg, parentLevelIDArg, includeArg)
}
