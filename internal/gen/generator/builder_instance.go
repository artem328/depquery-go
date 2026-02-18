package generator

import (
	. "github.com/dave/jennifer/jen"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (i *nameIndex) InstanceInterface() string {
	return "Instance"
}

func (i *nameIndex) InstanceResolveMethod() string {
	return "Resolve"
}

func (i *nameIndex) InstanceStruct() string {
	return "instance"
}

type instanceBuilder struct {
	naming   *nameIndex
	entities []*schema.Entity
	refedRev []refedReversed
	iface    Code
	impl     Code
}

func newInstanceBuilder(naming *nameIndex, entities []*schema.Entity, refedRev []refedReversed) *instanceBuilder {
	return &instanceBuilder{
		naming:   naming,
		entities: entities,
		refedRev: refedRev,
	}
}

func (b *instanceBuilder) Builders() []builder {
	return []builder{builderFunc(b.buildIface), builderFunc(b.buildImplementation)}
}

func (b *instanceBuilder) Interface() Code {
	return b.iface
}

func (b *instanceBuilder) Implementation() Code {
	return b.impl
}

func (b *instanceBuilder) buildIface() {
	//	type <iface> interface {
	//		<resolve>(context.Context) (<stateIface>, error)
	//	}
	b.iface = Type().Id(b.naming.InstanceInterface()).Interface(
		Id(b.naming.InstanceResolveMethod()).Params(contextCtx).Params(Id(b.naming.StateInterface()), Error()),
	)
}

//nolint:funlen // long but okay
func (b *instanceBuilder) buildImplementation() {
	//	type <struct> struct {
	//		ctx *<fetchContextStruct>
	//		resolvers [][]<resolverType>
	//		resolver <prefetchResolverIface>
	//		prefetcher <entityPrefetcherIface>
	//	}
	//
	//	func (i <struct>) <resolve>(ctx context.Context) (<stateIface>, error) {
	//		for _, level := range i.resolvers {
	//			for _, r := range level {
	//				r(i.ctx, i.resolver)
	//			}
	//
	//			if err := i.prefetch(ctx); err != nil {
	//				return nil, err
	//			}
	//		}
	//
	//		return i.ctx.state, nil
	//	}
	//
	//	func (i <struct>) prefetch(ctx context.Context) error {
	//		defer i.ctx.<flush>()
	//
	//		<prefetch> ...
	//
	//		return nil
	//	}
	b.impl = Do(func(s *Statement) {
		s.Type().Id(b.naming.InstanceStruct()).Struct(
			Id("ctx").Op("*").Id(b.naming.FetchContextStruct()),
			Id("resolvers").Index().Index().Id(b.naming.ResolverFunc()),
			Id("resolver").Id(b.naming.PrefetchResolverInterface()),
			Id("prefetcher").Id(b.naming.EntityPrefetcherInterface()),
		)
		s.Line()

		rcv := Id("i")
		methodParams := Params(rcv.Clone().Id(b.naming.InstanceStruct()))
		ctx := Id("ctx")

		s.Func().Add(methodParams).Id(b.naming.InstanceResolveMethod()).
			Params(ctx.Clone().Add(contextCtx)).
			Params(Id(b.naming.StateInterface()), Error()).
			BlockFunc(func(body *Group) {
				_u := Id("_")
				level := Id("level")
				r := Id("r")
				resolvers := rcv.Clone().Dot("resolvers")

				body.For(List(_u, level).Op(":=").Range().Add(resolvers)).BlockFunc(func(g *Group) {
					g.For(List(_u, r).Op(":=").Range().Add(level)).Block(
						r.Clone().Call(rcv.Clone().Dot("ctx"), rcv.Clone().Dot("resolver")),
					)
					g.Line()

					g.If(Err().Op(":=").Add(rcv).Dot("prefetch").Call(ctx), Err().Op("!=").Nil()).Block(
						Return(List(Nil(), Err())),
					)
				})
				body.Line()

				body.Return(List(rcv.Clone().Dot("ctx").Dot(b.naming.StateStruct()), Nil()))
			})
		s.Line().Line()

		s.Func().Add(methodParams).Id("prefetch").Params(ctx.Clone().Add(contextCtx)).Error().BlockFunc(func(body *Group) {
			body.Defer().Add(rcv).Dot("ctx").Dot(b.naming.FetchContextFlushMethod()).Call()
			body.Line()

			for _, e := range b.entities {
				body.Line().Add(b.prefetch(
					e,
					"i",
					"ctx",
					b.naming.FetchContextPendingField(e),
					b.naming.EntityPrefetcherMethodName(e),
				)).Line()
			}

			for _, r := range b.refedRev {
				body.Line().Add(b.prefetch(
					r.Entity,
					"i",
					"ctx",
					b.naming.FetchContextPendingReversedField(r.Entity, r.By),
					b.naming.EntityPrefetcherReversedMethodName(r.Entity, r.By),
				)).Line()
			}

			body.Return(Nil())
		})
	})
}

func (b *instanceBuilder) prefetch(e *schema.Entity, rcvName, ctxParam, pendingField, prefetchMethod string) Code {
	//	if p := i.ctx.<pendingField>; len(p) > 0 {
	//		ee, err := i.prefetcher.<prefetchMethod>(<ctxParam>, p)
	//		if err != nil {
	//			return err
	//		}
	//
	//		for e := range ee {
	//			i.ctx.<addEntity>(e)
	//		}
	//	}
	rcv := Id(rcvName)
	p := Id("p")

	return If(
		p.Clone().Op(":=").Add(rcv).Dot("ctx").Dot(pendingField),
		Len(p).Op(">").Lit(0),
	).BlockFunc(func(g *Group) {
		ee := Id("ee")

		g.List(ee, Err()).Op(":=").
			Add(rcv).Dot("prefetcher").Dot(prefetchMethod).Call(Id(ctxParam), p)
		g.If(Err().Op("!=").Nil()).Block(Return(Err()))
		g.Line()

		entity := Id("e")
		g.For(entity.Clone().Op(":=").Range().Add(ee)).Block(
			rcv.Clone().Dot("ctx").Dot(b.naming.StateAddMethod(e)).Call(entity),
		)
	})
}
