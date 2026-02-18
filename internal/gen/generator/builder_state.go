package generator

import (
	. "github.com/dave/jennifer/jen"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (i *nameIndex) StateInterface() string {
	return "State"
}

func (i *nameIndex) StateStruct() string {
	return "state"
}

func (i *nameIndex) StateConstructor() string {
	return "newState"
}

func (i *nameIndex) StateField(e *schema.Entity) string {
	return i.getOrCreate(nKey("StateField", e.Name, "", ""), func() string {
		return sanitizeID(e.Name, sanitizeUnexported)
	})
}

func (i *nameIndex) StateReversedField(e, by *schema.Entity) string {
	return i.getOrCreate(nKey("StateReversedField", e.Name, "", by.Name), func() string {
		return sanitizeID(e.Name, sanitizeUnexported) + "By" + sanitizeID(by.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) StateMethod(e *schema.Entity) string {
	return i.getOrCreate(nKey("StateMethod", e.Name, "", ""), func() string {
		return sanitizeID(e.Name, sanitizeExported)
	})
}

func (i *nameIndex) StateReversedMethod(e, by *schema.Entity) string {
	return i.getOrCreate(nKey("StateReversedMethod", e.Name, "", by.Name), func() string {
		return sanitizeID(e.Name, sanitizeExported) + "By" + sanitizeID(by.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) StateShouldMethod(e *schema.Entity) string {
	return i.getOrCreate(nKey("StateShouldMethod", e.Name, "", ""), func() string {
		return "Should" + sanitizeID(e.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) StateShouldReversedMethod(e, by *schema.Entity) string {
	return i.getOrCreate(nKey("StateShouldReversedMethod", e.Name, "", by.Name), func() string {
		return "Should" + sanitizeID(e.Name, sanitizeRawCapitalized) + "By" + sanitizeID(by.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) StateAddMethod(e *schema.Entity) string {
	return i.getOrCreate(nKey("StateAddMethod", e.Name, "", ""), func() string {
		return "add" + sanitizeID(e.Name, sanitizeRawCapitalized)
	})
}

type stateBuilder struct {
	naming           *nameIndex
	entities         []*schema.Entity
	revRefed         []refedReversed
	reverseRelations map[*schema.Entity][]reverseRelation
	iface            Code
	impl             Code
}

func newStateBuilder(
	naming *nameIndex,
	entities []*schema.Entity,
	revRefed []refedReversed,
	reverseRelations map[*schema.Entity][]reverseRelation,
) *stateBuilder {
	return &stateBuilder{
		naming:           naming,
		entities:         entities,
		revRefed:         revRefed,
		reverseRelations: reverseRelations,
	}
}

func (b *stateBuilder) Builders() []builder {
	return []builder{
		builderFunc(b.buildIface),
		builderFunc(b.buildImpl),
	}
}

func (b *stateBuilder) Interface() Code {
	return b.iface
}

func (b *stateBuilder) Implementation() Code {
	return b.impl
}

func (b *stateBuilder) buildIface() {
	//	type <interface> interface {
	//		<methodSignature>
	//		...
	//	}
	b.iface = Type().Id(b.naming.StateInterface()).InterfaceFunc(func(i *Group) {
		for _, e := range b.entities {
			i.Add(b.methodSignature(e, ""))
			i.Add(b.shouldMethodSignature(e, ""))
		}

		for _, r := range b.revRefed {
			i.Add(b.methodReversedSignature(r.Entity, r.By, ""))
			i.Add(b.shouldReversedMethodSignature(r.Entity, r.By, ""))
		}
	})
}

func (b *stateBuilder) buildImpl() {
	// 	type <struct> struct {
	//		<field>...
	//
	//		<revField>...
	//	}
	//
	//	<constructor>
	//
	//  <methodImpl>...
	//
	//	<addMethod>...
	b.impl = Type().Id(b.naming.StateStruct()).StructFunc(func(s *Group) {
		for _, e := range b.entities {
			s.Add(b.field(e))
		}

		for i, r := range b.revRefed {
			f := s.Null()
			if i == 0 {
				f.Line()
			}

			f.Add(b.fieldRev(r.Entity, r.By))
		}
	}).Line().
		Line().Add(b.constructor()).Line().
		Line().Do(func(s *Statement) {
		for _, e := range b.entities {
			s.Line().Add(b.methodImpl(e)).Line()
			s.Line().Add(b.shouldMethodImpl(e)).Line()
		}

		for _, r := range b.revRefed {
			s.Line().Add(b.methodReversedImpl(r.Entity, r.By)).Line()
			s.Line().Add(b.shouldReversedMethodImpl(r.Entity, r.By)).Line()
		}
	}).
		Do(func(s *Statement) {
			for _, e := range b.entities {
				s.Line().Add(b.addMethod(e)).Line()
			}
		})
}

func (b *stateBuilder) methodSignature(e *schema.Entity, idParam string) Code {
	// <method>(<idParam> <idType>) (<entityType>, bool)
	return Id(b.naming.StateMethod(e)).
		Params(Id(idParam).Add(typeToJen(e.ID.Type))).
		Params(typeToJen(e.Type), Bool())
}

func (b *stateBuilder) methodImpl(e *schema.Entity) Code {
	// id is a parameter name inside methodSignature
	//
	//	func(s <struct>) <methodSignature> {
	//		return s.<field>[id].Maybe()
	//	}
	return Func().Params(Id("s").Id(b.naming.StateStruct())).Add(b.methodSignature(e, "id")).Block(
		Return(Id("s").Dot(b.naming.StateField(e)).Index(Id("id")).Dot("Maybe").Call()),
	)
}

func (b *stateBuilder) shouldMethodSignature(e *schema.Entity, idParam string) Code {
	// <method>(<idParam> <idType>) (<entityType>, error)
	return Id(b.naming.StateShouldMethod(e)).
		Params(Id(idParam).Add(typeToJen(e.ID.Type))).
		Params(typeToJen(e.Type), Error())
}

func (b *stateBuilder) shouldMethodImpl(e *schema.Entity) Code {
	// id is a parameter name inside methodSignature
	//
	//	func(s <struct>) <shouldMethodSignature> {
	//		var (
	//			e <entityType>
	//			ok bool
	//		)
	//
	//		if e, ok = s.<field>[id].Maybe(); ok {
	//			return e, nil
	//		}
	//
	//		return e, depquery.NewNotInStateError("<name>", id)
	//	}
	return Func().Params(Id("s").Id(b.naming.StateStruct())).
		Add(b.shouldMethodSignature(e, "id")).
		BlockFunc(func(body *Group) {
			entity := Id("e")
			id := Id("id")
			ok := Id("ok")

			body.Var().Defs(
				Add(entity, typeToJen(e.Type)),
				ok.Clone().Bool(),
			)
			body.Line()

			body.If(
				List(entity, ok).Op("=").Id("s").Dot(b.naming.StateField(e)).Index(id).Dot("Maybe").Call(),
				ok,
			).Block(Return(List(entity, Nil())))
			body.Line()

			body.Return(List(entity, libNotInStateErrCtor.Clone().Call(Lit(e.Name), id)))
		})
}

func (b *stateBuilder) methodReversedSignature(e, by *schema.Entity, idParam string) Code {
	// <method>(<idParam> <idType>) (<entityType>, bool)
	return Id(b.naming.StateReversedMethod(e, by)).
		Params(Id(idParam).Add(typeToJen(by.ID.Type))).
		Params(typeToJen(e.Type), Bool())
}

func (b *stateBuilder) methodReversedImpl(e, by *schema.Entity) Code {
	// id is a parameter name inside methodSignature
	//
	//	func(s <struct>) <methodSignature> {
	//		return s.<method>(s.<revField>[id])
	//	}
	return Func().Params(Id("s").Id(b.naming.StateStruct())).Add(b.methodReversedSignature(e, by, "id")).Block(Return(
		Id("s").Dot(b.naming.StateMethod(e)).Call(
			Id("s").Dot(b.naming.StateReversedField(e, by)).Index(Id("id")),
		),
	))
}

func (b *stateBuilder) shouldReversedMethodSignature(e, by *schema.Entity, idParam string) Code {
	// <method>(<idParam> <idType>) (<entityType>, error)
	return Id(b.naming.StateShouldReversedMethod(e, by)).
		Params(Id(idParam).Add(typeToJen(by.ID.Type))).
		Params(typeToJen(e.Type), Error())
}

func (b *stateBuilder) shouldReversedMethodImpl(e, by *schema.Entity) Code {
	// id is a parameter name inside methodSignature
	//
	//	func(s <struct>) <shouldMethodSignature> {
	//		return s.<method>(s.<revField>[id])
	//	}
	return Func().Params(Id("s").
		Id(b.naming.StateStruct())).
		Add(b.shouldReversedMethodSignature(e, by, "id")).
		Block(Return(
			Id("s").Dot(b.naming.StateShouldMethod(e)).Call(
				Id("s").Dot(b.naming.StateReversedField(e, by)).Index(Id("id")),
			),
		))
}

func (b *stateBuilder) addMethod(e *schema.Entity) Code {
	//	func (s <struct>) <addMethod>(e <entityType>) {
	//		s.<field>[e.<idMember>] = <Fetched>(e)
	//
	//		<reverseMapping>...
	//	}
	return Func().Params(Id("s").Id(b.naming.StateStruct())).Id(b.naming.StateAddMethod(e)).
		Params(Id("e").Add(typeToJen(e.Type))).
		BlockFunc(func(body *Group) {
			body.Id("s").Dot(b.naming.StateField(e)).Index(Id("e").Add(memberToJen(e.ID))).Op("=").Add(libFetched).Call(Id("e"))

			rr, ok := b.reverseRelations[e]
			if !ok {
				return
			}

			body.Line()

			for _, r := range rr {
				body.Add(b.reverseMapping(e, r, "e"))
			}
		}).Line()
}

func (b *stateBuilder) reverseMapping(e *schema.Entity, r reverseRelation, entityParam string) Code {
	// s.<reverseField>[<entityParam>.<reverseMember>] = <entityParam>.<idMember>
	entity := Id(entityParam)

	return Id("s").Dot(b.naming.StateReversedField(e, r.Ref)).Index(Add(entity, memberToJen(r.Member))).Op("=").
		Add(entity, memberToJen(e.ID))
}

func (b *stateBuilder) field(e *schema.Entity) Code {
	// <field> <fieldType>
	return Id(b.naming.StateField(e)).Add(b.fieldType(e))
}

func (b *stateBuilder) fieldType(e *schema.Entity) Code {
	// map[<idType>]MaybeFetched[<entityType>]
	return Map(typeToJen(e.ID.Type)).Add(libMaybeFetched).Types(typeToJen(e.Type))
}

func (b *stateBuilder) fieldRev(e, by *schema.Entity) Code {
	// <field> <fieldType>
	return Id(b.naming.StateReversedField(e, by)).Add(b.fieldRevType(e, by))
}

func (b *stateBuilder) fieldRevType(e, by *schema.Entity) Code {
	// map[<byIdType>]<entityIdType>
	return Map(typeToJen(by.ID.Type)).Add(typeToJen(e.ID.Type))
}

func (b *stateBuilder) constructor() Code {
	// 	func <stateConstructor>() <struct> {
	//		return <struct>{
	//			<field>: make(<fieldType>), ...
	//
	//			<revField>: make(<refFieldType>), ...
	//  	}
	// 	}
	return Func().Id(b.naming.StateConstructor()).Params().Id(b.naming.StateStruct()).Block(
		Return(
			Id(b.naming.StateStruct()).CustomFunc(multilineValuesOpts, func(v *Group) {
				for _, e := range b.entities {
					v.Id(b.naming.StateField(e)).Op(":").Make(b.fieldType(e))
				}

				for i, r := range b.revRefed {
					f := v.Null()
					if i == 0 {
						f.Line()
					}

					f.Id(b.naming.StateReversedField(r.Entity, r.By)).Op(":").Make(b.fieldRevType(r.Entity, r.By))
				}
			}),
		),
	)
}
