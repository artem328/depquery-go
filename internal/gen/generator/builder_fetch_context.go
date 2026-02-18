package generator

import (
	. "github.com/dave/jennifer/jen"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (i *nameIndex) FetchContextStruct() string {
	return "fetchContext"
}

func (i *nameIndex) FetchContextConstructor() string {
	return "newFetchContext"
}

func (i *nameIndex) FetchContextFlushMethod() string {
	return "flush"
}

func (i *nameIndex) FetchContextEntityConstructor(e *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextEntityConstructor", e.Name, "", ""), func() string {
		return "new" + sanitizeID(e.Name, sanitizeRawCapitalized) + "FetchContext"
	})
}

func (i *nameIndex) FetchContextByIDField(e *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextByIDField", e.Name, "", ""), func() string {
		return sanitizeID(e.Name, sanitizeUnexported) + "ByID"
	})
}

func (i *nameIndex) FetchContextByIDReversedField(e, by *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextByIDReversedField", e.Name, "", by.Name), func() string {
		return sanitizeID(e.Name, sanitizeUnexported) + "By" + sanitizeID(by.Name, sanitizeRawCapitalized) + "ByID"
	})
}

func (i *nameIndex) FetchContextPendingField(e *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextPendingField", e.Name, "", ""), func() string {
		return "pending" + sanitizeID(e.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) FetchContextPendingReversedField(e, by *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextPendingReversedField", e.Name, "", by.Name), func() string {
		return "pending" + sanitizeID(e.Name, sanitizeRawCapitalized) + "By" + sanitizeID(by.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) FetchContextSeenField(e *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextSeenField", e.Name, "", ""), func() string {
		return "seen" + sanitizeID(e.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) FetchContextSeenReversedField(e, by *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextSeenReversedField", e.Name, "", by.Name), func() string {
		return "seen" + sanitizeID(e.Name, sanitizeRawCapitalized) + "By" + sanitizeID(by.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) FetchContextAddPrefetchMethod(e *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextAddPrefetchMethod", e.Name, "", ""), func() string {
		return "addPrefetch" + sanitizeID(e.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) FetchContextAddPrefetchReverseMethod(e, by *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextAddPrefetchReverseMethod", e.Name, "", by.Name), func() string {
		return "addPrefetch" + sanitizeID(e.Name, sanitizeRawCapitalized) + "By" + sanitizeID(by.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) FetchContextGetMethod(e *schema.Entity) string {
	return i.getOrCreate(nKey("FetchContextGetMethod", e.Name, "", ""), func() string {
		return "get" + sanitizeID(e.Name, sanitizeRawCapitalized)
	})
}

type fetchContextBuilder struct {
	naming            *nameIndex
	entities          []*schema.Entity
	refed             []*schema.Entity
	refedReversed     []refedReversed
	reversedRelations map[*schema.Entity][]reverseRelation
	impl              Code
}

func newFetchContextBuilder(
	naming *nameIndex,
	entities, refed []*schema.Entity,
	refedReversed []refedReversed,
	reversedRelations map[*schema.Entity][]reverseRelation,
) *fetchContextBuilder {
	return &fetchContextBuilder{
		naming:            naming,
		entities:          entities,
		refed:             refed,
		refedReversed:     refedReversed,
		reversedRelations: reversedRelations,
	}
}

func (b *fetchContextBuilder) Builders() []builder {
	return []builder{builderFunc(b.buildImplementation)}
}

func (b *fetchContextBuilder) Implementation() Code {
	return b.impl
}

//nolint:gocognit,cyclop,funlen // needs refactoring maybe
func (b *fetchContextBuilder) buildImplementation() {
	//	type <struct> struct {
	//		<state>
	//
	//		<byIdField> <byIdType>...
	//
	//		<reversedByIdField> <byIdType>...
	//
	//		<seenField> <byIdSet>...
	//
	//		<reverseSeenField> <byIdSet>...
	//
	//		<pendingField> []<idType>...
	//
	//		<reversePendingField> []<idType>...
	//	}
	//
	//	<constructor>
	//
	//	<entityConstructor>...
	//
	//	<addPrefetchMethod>...
	//
	//  <getMethod>...
	//
	//	<addOverrideMethod>...
	//
	//  <flush>
	b.impl = Do(func(s *Statement) {
		s.Type().Id(b.naming.FetchContextStruct()).StructFunc(func(s *Group) {
			s.Id(b.naming.StateStruct())

			// byId fields
			for i, e := range b.entities {
				f := s.Null()
				if i == 0 {
					f.Line()
				}

				f.Id(b.naming.FetchContextByIDField(e)).Add(b.byIDFieldType(e))
			}

			// byId reversed fields
			for i, r := range b.refedReversed {
				f := s.Null()
				if i == 0 {
					f.Line()
				}

				f.Id(b.naming.FetchContextByIDReversedField(r.Entity, r.By)).Add(b.byIDFieldType(r.By))
			}

			// seen fields
			for i, e := range b.refed {
				f := s.Null()
				if i == 0 {
					f.Line()
				}

				f.Id(b.naming.FetchContextSeenField(e)).Add(b.byIDSet(e))
			}

			// seen reversed fields
			for i, r := range b.refedReversed {
				f := s.Null()
				if i == 0 {
					f.Line()
				}

				f.Id(b.naming.FetchContextSeenReversedField(r.Entity, r.By)).Add(b.byIDSet(r.By))
			}

			// pending fields
			for i, e := range b.refed {
				f := s.Null()
				if i == 0 {
					f.Line()
				}

				f.Id(b.naming.FetchContextPendingField(e)).Index().Add(typeToJen(e.ID.Type))
			}

			// pending reversed fields
			for i, r := range b.refedReversed {
				f := s.Null()
				if i == 0 {
					f.Line()
				}

				f.Id(b.naming.FetchContextPendingReversedField(r.Entity, r.By)).Index().Add(typeToJen(r.By.ID.Type))
			}
		})
		s.Line()

		s.Add(b.constructor())
		s.Line()

		for _, e := range b.entities {
			s.Line().Add(b.entityConstructor(e)).Line()
		}

		for _, e := range b.refed {
			s.Line().Add(b.addPrefetchMethod(e)).Line()
		}

		for _, r := range b.refedReversed {
			s.Line().Add(b.addPrefetchReversedMethod(r.Entity, r.By)).Line()
		}

		for _, e := range b.entities {
			s.Line().Add(b.getMethod(e)).Line()
		}

		for _, e := range b.entities {
			rr, ok := b.reversedRelations[e]
			if !ok {
				continue
			}

			s.Line().Add(b.addMethodOverride(e, rr)).Line()
		}

		s.Line().Add(b.flush())
	})
}

func (b *fetchContextBuilder) byIDSet(e *schema.Entity) Code {
	// map[<entityIdType>]struct{}
	return Map(typeToJen(e.ID.Type)).Struct()
}

func (b *fetchContextBuilder) byIDFieldType(e *schema.Entity) Code {
	// map[<libID>]<byIDSet>
	return Map(libID).Add(b.byIDSet(e))
}

func (b *fetchContextBuilder) constructor() Code {
	//	func <constructor>() *<struct> {
	//		return &<struct>{
	//			state: <stateConstructor>(),
	//
	//			<seenField>: make(<byIdSet>), ...
	//
	//			<reverseSeenField> <byIdSet>, ...
	//
	//			<byIdField>: make(<byIdType>), ...
	//
	//			<reversePendingField> []<idType>...
	//		}
	//	}
	return Func().Id(b.naming.FetchContextConstructor()).Params().Op("*").Id(b.naming.FetchContextStruct()).Block(
		Return(
			Op("&").Id(b.naming.FetchContextStruct()).CustomFunc(multilineValuesOpts, func(s *Group) {
				s.Id(b.naming.StateStruct()).Op(":").Id(b.naming.StateConstructor()).Call()

				for i, e := range b.entities {
					l := Null()
					if i == 0 {
						l = Line()
					}

					s.Add(l).Id(b.naming.FetchContextByIDField(e)).Op(":").Make(b.byIDFieldType(e))
				}

				for i, r := range b.refedReversed {
					l := s.Null()
					if i == 0 {
						l.Line()
					}

					l.Id(b.naming.FetchContextByIDReversedField(r.Entity, r.By)).Op(":").Make(b.byIDFieldType(r.By))
				}

				for i, e := range b.refed {
					l := Null()
					if i == 0 {
						l = Line()
					}

					s.Add(l).Id(b.naming.FetchContextSeenField(e)).Op(":").Make(b.byIDSet(e))
				}

				for i, r := range b.refedReversed {
					l := s.Null()
					if i == 0 {
						l.Line()
					}

					l.Id(b.naming.FetchContextSeenReversedField(r.Entity, r.By)).Op(":").Make(b.byIDSet(r.By))
				}
			}),
		),
	)
}

func (b *fetchContextBuilder) entityConstructor(e *schema.Entity) Code {
	//	func <entityConstructorName>(i iter.Seq[<entityType>]) *<struct> {
	//		ctx := <fetchCtxConstructor>()
	//
	//		zero := make(<byIdSet>)
	//		for x := range i {
	//			zero[x.<idMember>] = struct{}{}
	//			ctx.<stateField>[x.<idMember>] = <Fetched>(x)
	//		}
	//		ctx.<byIdField>[0] = zero
	//
	//		return ctx
	//	}
	return Func().Id(b.naming.FetchContextEntityConstructor(e)).
		Params(Id("i").Add(iterSeq).Types(typeToJen(e.Type))).
		Op("*").Id(b.naming.FetchContextStruct()).
		BlockFunc(func(body *Group) {
			ctx := Id("ctx")

			body.Add(ctx).Op(":=").Id(b.naming.FetchContextConstructor()).Call()
			body.Line()

			zero := Id("zero")
			body.Add(zero).Op(":=").Make(b.byIDSet(e))

			x := Id("x")
			body.For(x.Clone().Op(":=").Range().Id("i")).BlockFunc(func(g *Group) {
				idRcv := x.Clone().Add(memberToJen(e.ID))
				g.Add(zero).Index(idRcv).Op("=").Struct().Values()
				g.Add(ctx).Dot(b.naming.StateField(e)).Index(idRcv).Op("=").Add(libFetched).Call(x)
			})
			body.Add(ctx).Dot(b.naming.FetchContextByIDField(e)).Index(Lit(0)).Op("=").Add(zero)
			body.Line()
			body.Return(ctx)
		})
}

func (b *fetchContextBuilder) addPrefetchMethod(e *schema.Entity) Code {
	//	func (ctx *<struct>) <method>(id <entityIdType>, rid depquery.ID) {
	//		if _, ok := ctx.<byIdField>[rid]; !ok {
	//			ctx.<byIdField>[rid] = make(<byIdSet>)
	//		}
	//
	//		ctx.<byIdField>[rid][id] = struct{}{}
	//
	//		if _, ok := ctx.<seenField>[id]; ok || ctx.<stateField>[id].Fetched() {
	//			return
	//		}
	//
	//		ctx.<seenField>[id] = struct{}{}
	//		ctx.<pendingField> = append(ctx.<pendingField>, id)
	//	}
	ctx := Id("ctx")
	id := Id("id")
	rid := Id("rid")
	_u := Id("_")

	return Func().Params(ctx.Clone().Op("*").Id(b.naming.FetchContextStruct())).
		Id(b.naming.FetchContextAddPrefetchMethod(e)).
		Params(
			Add(id, typeToJen(e.ID.Type)),
			Add(rid, libID),
		).
		BlockFunc(func(body *Group) {
			ok := Id("ok")

			body.If(List(_u, ok).Op(":=").Add(ctx).Dot(b.naming.FetchContextByIDField(e)).Index(rid), Op("!").Add(ok)).Block(
				ctx.Clone().Dot(b.naming.FetchContextByIDField(e)).Index(rid).Op("=").Make(b.byIDSet(e)),
			)
			body.Line()

			body.Add(ctx).Dot(b.naming.FetchContextByIDField(e)).Index(rid).Index(id).Op("=").Struct().Values()
			body.Line()

			body.If(
				List(_u, ok).Op(":=").Add(ctx).Dot(b.naming.FetchContextSeenField(e)).Index(id),
				ok.Clone().Op("||").Add(ctx).Dot(b.naming.StateField(e)).Index(id).Dot("Fetched").Call(),
			).Block(Return())
			body.Line()

			body.Add(ctx).Dot(b.naming.FetchContextSeenField(e)).Index(id).Op("=").Struct().Values()
			pending := ctx.Clone().Dot(b.naming.FetchContextPendingField(e))

			body.Add(pending).Op("=").Append(pending, id)
		}).Line()
}

func (b *fetchContextBuilder) addPrefetchReversedMethod(e, by *schema.Entity) Code {
	//	func (ctx *<struct>) <method>(id <revIdType>, rid depquery.ID) {
	//		if _, ok := ctx.<revByIdField>[rid]; !ok {
	//			ctx.<revByIdField>[rid] = make(<byIdSet>)
	//		}
	//
	//		ctx.<revByIdField>[rid][id] = struct{}{}
	//
	//		if _, ok := ctx.<revSeenField>[id]; ok || ctx.<stateField>[ctx.<revStateField>[id]].Fetched() {
	//			return
	//		}
	//
	//		ctx.<revSeenField>[id] = struct{}{}
	//		ctx.<revPendingField> = append(ctx.<revPendingField>, id)
	//	}
	ctx := Id("ctx")
	id := Id("id")
	rid := Id("rid")
	_u := Id("_")

	return Func().Params(ctx.Clone().Op("*").Id(b.naming.FetchContextStruct())).
		Id(b.naming.FetchContextAddPrefetchReverseMethod(e, by)).
		Params(
			Add(id, typeToJen(by.ID.Type)),
			Add(rid, libID),
		).
		BlockFunc(func(body *Group) {
			ok := Id("ok")

			body.If(
				List(_u, ok).Op(":=").Add(ctx).Dot(b.naming.FetchContextByIDReversedField(e, by)).Index(rid),
				Op("!").Add(ok),
			).Block(
				ctx.Clone().Dot(b.naming.FetchContextByIDReversedField(e, by)).Index(rid).Op("=").Make(b.byIDSet(by)),
			)
			body.Line()

			body.Add(ctx).Dot(b.naming.FetchContextByIDReversedField(e, by)).Index(rid).Index(id).Op("=").Struct().Values()
			body.Line()

			body.If(
				List(_u, ok).Op(":=").Add(ctx).Dot(b.naming.FetchContextSeenReversedField(e, by)).Index(id),
				ok.Clone().Op("||").Add(ctx).Dot(b.naming.StateField(e)).Index(
					ctx.Clone().Dot(b.naming.StateReversedField(e, by)).Index(id),
				).Dot("Fetched").Call(),
			).Block(Return())
			body.Line()

			body.Add(ctx).Dot(b.naming.FetchContextSeenReversedField(e, by)).Index(id).Op("=").Struct().Values()
			pending := ctx.Clone().Dot(b.naming.FetchContextPendingReversedField(e, by))

			body.Add(pending).Op("=").Append(pending, id)
		}).Line()
}

func (b *fetchContextBuilder) getMethod(e *schema.Entity) Code {
	//	func (ctx *<struct>) <method>(rid depquery.ID) iter.Seq[<entityType>] {
	//		return func(yield func(<entityType>) bool) {
	//			for id := range ctx.<byIDField>[rid] {
	//				e := ctx.<stateField>[id]
	//
	//				if e.Fetched() && !yield(e) {
	//					return
	//				}
	//			}
	//		}
	//	}
	ctx := Id("ctx")
	rid := Id("rid")
	yield := Id("yield")

	return Func().Params(ctx.Clone().Op("*").Id(b.naming.FetchContextStruct())).
		Id(b.naming.FetchContextGetMethod(e)).Params(Add(rid, libID)).Add(iterSeq).Types(typeToJen(e.Type)).
		Block(Return(
			Func().Params(yield.Clone().Func().Params(typeToJen(e.Type)).Bool()).BlockFunc(func(g *Group) {
				id := Id("id")
				g.For(id.Clone().Op(":=").Range().Add(ctx).Dot(b.naming.FetchContextByIDField(e)).Index(rid)).
					BlockFunc(func(g *Group) {
						entity := Id("e")
						g.Add(entity).Op(":=").Add(ctx).Dot(b.naming.StateField(e)).Index(id)
						g.Line()

						g.If(entity.Clone().Dot("Fetched").Call().Op("&&").
							Op("!").Add(yield).Call(entity.Clone().Dot("Value").Call())).
							Block(Return())
					})
			}),
		))
}

func (b *fetchContextBuilder) addMethodOverride(e *schema.Entity, rr []reverseRelation) Code {
	//	func (ctx *<struct>) <method>(e <entityType>) {
	//		ctx.<stateStruct>.<addMethod>(e)
	//
	//		<mapReversedPid> ...
	//	}
	ctx := Id("ctx")
	entity := Id("e")

	return Func().Params(ctx.Clone().Op("*").Id(b.naming.FetchContextStruct())).Id(b.naming.StateAddMethod(e)).
		Params(Add(entity, typeToJen(e.Type))).
		BlockFunc(func(body *Group) {
			body.Add(ctx).Dot(b.naming.StateStruct()).Dot(b.naming.StateAddMethod(e)).Call(entity)
			body.Line()

			for _, r := range rr {
				body.Add(b.mapReversePid(e, r, "ctx", "e"))
			}
		})
}

func (b *fetchContextBuilder) mapReversePid(e *schema.Entity, r reverseRelation, rcvParam, entityParam string) Code {
	//  {
	//		for pid, s := range <rcv>.<reversedByIdField> {
	//			if _, ok := s[<entityParam>.<revMember>]; ok {
	//				if _, ok := <rcv>.<byIdField>[pid]; !ok {
	//					<rcv>.<byIdField>[pid] = make(<byIdSet>)
	//				}
	//				<rcv>.<byIdField>[pid][<entityParam>.<idMember>] = struct{}{}
	//			}
	//		}
	//	}
	rcv := Id(rcvParam)
	pid := Id("pid")
	s := Id("s")
	entity := Id(entityParam)

	return Block(For(List(pid, s).Op(":=").Range().Add(rcv).Dot(b.naming.FetchContextByIDReversedField(e, r.Ref))).
		BlockFunc(func(bl *Group) {
			_u := Id("_")
			ok := Id("ok")
			byIDIndexed := rcv.Clone().Dot(b.naming.FetchContextByIDField(e)).Index(pid)

			bl.If(List(_u, ok).Op(":=").Add(s).Index(Add(entity, memberToJen(r.Member))), ok).Block(
				If(List(_u, ok).Op(":=").Add(byIDIndexed), Op("!").Add(ok)).Block(
					byIDIndexed.Clone().Op("=").Make(b.byIDSet(e)),
				),

				byIDIndexed.Clone().Index(Add(entity, memberToJen(e.ID))).Op("=").Struct().Values(),
			)
		}))
}

func (b *fetchContextBuilder) flush() Code {
	// 	func (ctx *<struct>) <method>() {
	//		ctx.<pendingField> = ctx.<pendingField>[:0] ...
	//	}
	ctx := Id("ctx")

	return Func().Params(ctx.Clone().Op("*").Id(b.naming.FetchContextStruct())).
		Id(b.naming.FetchContextFlushMethod()).Params().
		BlockFunc(func(body *Group) {
			for _, e := range b.refed {
				body.Add(b.flushField(ctx.Clone().Dot(b.naming.FetchContextPendingField(e))))
			}

			for _, r := range b.refedReversed {
				body.Add(b.flushField(ctx.Clone().Dot(b.naming.FetchContextPendingReversedField(r.Entity, r.By))))
			}
		})
}

func (b *fetchContextBuilder) flushField(pendingField Code) Code {
	return Add(pendingField).Op("=").Add(pendingField).Index(Empty(), Lit(0))
}
