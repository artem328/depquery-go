package jen

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderResolvers() {
	r.renderResolverType()

	for _, er := range r.plan.EntityResolvers {
		r.renderResolver(er)
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

func (r *Renderer) renderResolver(er plan.EntityResolver) {
	r.renderResolverOptionConstants(er)
	r.renderResolverConstructor(er)
}

func (r *Renderer) generateResolverCall(resolver, fetchContext, prefetchResolver Code) Code {
	return Add(resolver).Call(fetchContext, prefetchResolver)
}

func (r *Renderer) renderResolverOptionConstants(er plan.EntityResolver) {
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

func (r *Renderer) renderResolverConstructor(er plan.EntityResolver) {
	entityLevelID := Id("eid")
	parentLevelID := Id("pid")
	include := Id("include")
	fetchContext := Id("ctx")
	prefetchResolver := Id("pr")
	entity := Id("e")

	r.f.Add(block(
		Func().Id(r.naming.Resolver.Constructor[er.ID]).Params(Add(entityLevelID), Add(parentLevelID, libID), Add(include, Uint64())).Params(Id(r.naming.Resolver.Type)).Block(
			Return(
				Func().Add(r.generateResolverFuncSignature(fetchContext, prefetchResolver)).Block(
					For(Add(entity).Op(":=").Range().Add(r.generateFetchContextParentGetterMethodCall(er.FetchParent, fetchContext, parentLevelID))).Block(
						r.generateResolverResolveBlocks(er, fetchContext, prefetchResolver, entity, include, entityLevelID)...,
					),
				),
			),
		),
	))
}

func (r *Renderer) generateResolverResolveBlocks(er plan.EntityResolver, fetchContext, prefetchResolver, entity, include, entityLevelID Code) []Code {
	var blocks []Code

	for _, rel := range er.Resolutions {
		blocks = append(blocks, r.generateResolverResolveBlock(rel, fetchContext, prefetchResolver, entity, include, entityLevelID)...)
	}

	return blocks
}

func (r *Renderer) generateResolverResolveBlock(er plan.EntityResolution, fetchContext, prefetchResolver, entity, include, entityLevelID Code) []Code {
	switch err := er.(type) {
	case plan.EntityResolutionRelation:
		return r.generateResolverEntityResolutionRelationBlock(err, fetchContext, prefetchResolver, entity, include, entityLevelID)
	case plan.EntityResolutionVariant:
		entityVariant := Id("v")

		var blocks []Code
		for _, ier := range err.Resolutions {
			blocks = append(blocks, r.generateResolverResolveBlock(ier, fetchContext, prefetchResolver, entityVariant, include, entityLevelID)...)
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
		If(r.generateResolverIncludeCheck(err.Relation, include)).Block(code),
	}
}

func (r *Renderer) generateResolverIncludeCheck(rid plan.RelationID, include Code) Code {
	return Add(include).Op("&").Id(r.naming.Relation.ConstantName[rid]).Op("!=").Lit(0)
}

func (r *Renderer) generateResolverConstructorCall(erid plan.EntityResolverID, entityLevelIDArg, parentLevelIDArg, includeArg Code) Code {
	return Id(r.naming.Resolver.Constructor[erid]).Call(entityLevelIDArg, parentLevelIDArg, includeArg)
}
