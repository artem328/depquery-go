package generator

import (
	"sync"

	. "github.com/dave/jennifer/jen"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

func (i *nameIndex) BuilderInterface(e *schema.Entity) string {
	return i.getOrCreate(nKey("BuilderInterface", e.Name, "", ""), func() string {
		return sanitizeID(e.Name, sanitizeExported) + "Builder"
	})
}

func (i *nameIndex) BuilderStruct(e *schema.Entity) string {
	return i.getOrCreate(nKey("BuilderStruct", e.Name, "", ""), func() string {
		return sanitizeID(e.Name, sanitizeUnexported) + "Builder"
	})
}

func (i *nameIndex) BuilderVariantInterface(e *schema.Entity, v schema.EntityVariant) string {
	return i.getOrCreate(nKey("BuilderVariantInterface", e.Name, v.Name, ""), func() string {
		return sanitizeID(e.Name, sanitizeExported) + sanitizeID(v.Name, sanitizeRawCapitalized) + "Builder"
	})
}

func (i *nameIndex) BuilderVariantStruct(e *schema.Entity, v schema.EntityVariant) string {
	return i.getOrCreate(nKey("BuilderVariantStruct", e.Name, v.Name, ""), func() string {
		return sanitizeID(e.Name, sanitizeUnexported) + "V" + sanitizeID(v.Name, sanitizeRawCapitalized) + "Builder"
	})
}

func (i *nameIndex) BuilderWithMethod(r schema.Relation) string {
	return i.getOrCreate(nKey("BuilderWithMethod", "", "", r.Name), func() string {
		return "With" + sanitizeID(r.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) BuilderDeepMethod(r schema.Relation) string {
	return i.getOrCreate(nKey("BuilderDeepMethod", "", "", r.Name), func() string {
		return sanitizeID(r.Name, sanitizeExported)
	})
}

func (i *nameIndex) BuilderVariantMethod(v schema.EntityVariant) string {
	return i.getOrCreate(nKey("BuilderVariantMethod", "", v.Name, ""), func() string {
		return "If" + sanitizeID(v.Name, sanitizeRawCapitalized)
	})
}

func (i *nameIndex) BuilderConstructor(e *schema.Entity) string {
	return i.getOrCreate(nKey("BuilderConstructor", e.Name, "", ""), func() string {
		return "new" + sanitizeID(e.Name, sanitizeRawCapitalized) + "Builder"
	})
}

func (i *nameIndex) BuilderVariantConstructor(e *schema.Entity, v schema.EntityVariant) string {
	return i.getOrCreate(nKey("BuilderVariantConstructor", e.Name, v.Name, ""), func() string {
		return "new" + sanitizeID(e.Name, sanitizeRawCapitalized) +
			"V" + sanitizeID(v.Name, sanitizeRawCapitalized) +
			"Builder"
	})
}

func (i *nameIndex) BuilderCompiler(e *schema.Entity) string {
	return i.getOrCreate(nKey("BuilderCompiler", e.Name, "", ""), func() string {
		return sanitizeID(e.Name, sanitizeExported)
	})
}

func (i *nameIndex) BuilderNestedBuilderField(r schema.Relation) string {
	return i.getOrCreate(nKey("BuilderNestedBuilderField", "", "", r.Name), func() string {
		return sanitizeID(r.Name, sanitizeUnexported)
	})
}

func (i *nameIndex) BuilderNestedVariantBuilderField(v schema.EntityVariant) string {
	return i.getOrCreate(nKey("BuilderNestedVariantBuilderField", "", v.Name, ""), func() string {
		return "v" + sanitizeID(v.Name, sanitizeRawCapitalized)
	})
}

type builderBuilder struct {
	naming   *nameIndex
	entities []*schema.Entity
	ifacesMu sync.Mutex
	ifaces   map[*schema.Entity][]Code
	implsMu  sync.Mutex
	impls    map[*schema.Entity][]Code
}

func newBuilderBuilder(naming *nameIndex, entities []*schema.Entity) *builderBuilder {
	return &builderBuilder{
		naming:   naming,
		entities: entities,
		ifaces:   make(map[*schema.Entity][]Code),
		impls:    make(map[*schema.Entity][]Code),
	}
}

func (b *builderBuilder) Builders() []builder {
	builders := make([]builder, 0, len(b.entities)+len(b.entities))

	for _, e := range b.entities {
		builders = append(
			builders,
			builderFunc(func() {
				b.buildIface(e)
			}),
			builderFunc(func() {
				b.buildImplementation(e)
			}),
		)
	}

	return builders
}

func (b *builderBuilder) Interface() Code {
	return Do(func(s *Statement) {
		for _, e := range b.entities {
			s.Add(b.ifaces[e]...)
		}
	})
}

func (b *builderBuilder) Implementation() Code {
	return Do(func(s *Statement) {
		for _, e := range b.entities {
			s.Add(b.impls[e]...)
		}
	})
}

func (b *builderBuilder) buildIface(e *schema.Entity) {
	needCompiler := false

	if i := b.buildEntityInterface(e); i != nil {
		b.addIface(e, Add(i).Line().Line())

		needCompiler = true
	}

	for _, v := range e.Variants {
		if i := b.buildVariantBuilderInterface(e, v); i != nil {
			b.addIface(e, Add(i).Line().Line())

			needCompiler = true
		}
	}

	if needCompiler {
		b.addIface(e, Add(b.buildCompiler(e)).Line().Line())
	}
}

func (b *builderBuilder) addIface(e *schema.Entity, code Code) {
	b.ifacesMu.Lock()
	defer b.ifacesMu.Unlock()

	b.ifaces[e] = append(b.ifaces[e], code)
}

func (b *builderBuilder) buildImplementation(e *schema.Entity) {
	if i := b.buildBuilder(e); i != nil {
		b.addImpl(e, Add(i).Line().Line())
	}

	for _, v := range e.Variants {
		if i := b.buildVariantBuilder(e, v); i != nil {
			b.addImpl(e, Add(i).Line().Line())
		}
	}
}

func (b *builderBuilder) addImpl(e *schema.Entity, code Code) {
	b.implsMu.Lock()
	defer b.implsMu.Unlock()

	b.impls[e] = append(b.impls[e], code)
}

func (b *builderBuilder) buildEntityInterface(e *schema.Entity) Code {
	added := false
	i := Type().Id(b.naming.BuilderInterface(e)).InterfaceFunc(func(i *Group) {
		for _, r := range e.Relations {
			i.Add(b.withMethodSignature(r, b.naming.BuilderInterface(e)))

			if len(r.Entity.Relations) > 0 {
				i.Add(b.deepMethodSignature(r, "", b.naming.BuilderInterface(e)))
			}

			added = true
		}

		for _, v := range e.Variants {
			if len(v.Relations) > 0 {
				i.Add(b.variantMethodSignature(e, v, "", b.naming.BuilderInterface(e)))

				added = true
			}
		}
	})

	if added {
		return i
	}

	return nil
}

func (b *builderBuilder) buildVariantBuilderInterface(e *schema.Entity, v schema.EntityVariant) Code {
	added := false
	i := Type().Id(b.naming.BuilderVariantInterface(e, v)).InterfaceFunc(func(i *Group) {
		for _, r := range v.Relations {
			i.Add(b.withMethodSignature(r, b.naming.BuilderVariantInterface(e, v)))

			if len(r.Entity.Relations) > 0 {
				i.Add(b.deepMethodSignature(r, "", b.naming.BuilderVariantInterface(e, v)))
			}

			added = true
		}
	})

	if added {
		return i
	}

	return nil
}

func (b *builderBuilder) buildCompiler(e *schema.Entity) Code {
	//	func <compiler>(f func(<iface>)) <compilerIface>[<iface>, <entityType>] {
	//		return <constr>(<buildContextConstr>(), 0).<compilerWith>(f)
	//	}
	return Func().Id(b.naming.BuilderCompiler(e)).
		Params(Id("f").Func().Params(Id(b.naming.BuilderInterface(e)))).
		Id(b.naming.CompilerInterface()).Types(Id(b.naming.BuilderInterface(e)), typeToJen(e.Type)).
		Block(Return(
			Id(b.naming.BuilderConstructor(e)).Call(
				Id(b.naming.BuildContextConstructor()).Call(),
				Lit(0),
			).Dot(b.naming.CompilerWithMethod()).Call(Id("f")),
		))
}

func (b *builderBuilder) buildBuilder(e *schema.Entity) Code {
	var (
		methods             = make([]Code, 0)
		nestedBuilderFields = make([]Code, 0)
		nestedVariantFields = make([]Code, 0)
	)

	rcv := Id("b").Op("*").Id(b.naming.BuilderStruct(e))

	for _, r := range e.Relations {
		//	func (b *<struct>) <withMethodSig> {
		//		<withMethodBody>
		//	}
		m := Func().Params(rcv).Add(b.withMethodSignature(r, b.naming.BuilderInterface(e))).
			Block(b.withMethodBody("b", b.naming.ResolverRelConst(e, r))...).
			Line().Line()
		methods = append(methods, m)

		if len(r.Entity.Relations) == 0 {
			continue
		}

		//	func (b *<struct>) <deepMethodSig> {
		//		<deepMethodBody>
		//	}
		m = Func().Params(rcv).Add(b.deepMethodSignature(r, "f", b.naming.BuilderInterface(e))).
			Block(b.deepMethodBody("b", "f", r)...).
			Line().Line()
		methods = append(methods, m)

		// <nestedBuilderField> <iface>
		nb := Id(b.naming.BuilderNestedBuilderField(r)).Id(b.naming.BuilderInterface(r.Entity))
		nestedBuilderFields = append(nestedBuilderFields, nb)
	}

	for _, v := range e.Variants {
		if len(v.Relations) == 0 {
			continue
		}

		//	func (b *<struct>) <variantMethodSig> {
		//		<variantMethodBody>
		//	}
		m := Func().Params(rcv).Add(b.variantMethodSignature(e, v, "f", b.naming.BuilderInterface(e))).
			Block(b.variantMethodBody("b", "f", e, v)...).
			Line().Line()
		methods = append(methods, m)

		// <nestedVariantBuilderField> <variantIface>
		nb := Id(b.naming.BuilderNestedVariantBuilderField(v)).Id(b.naming.BuilderVariantInterface(e, v))
		nestedVariantFields = append(nestedVariantFields, nb)
	}

	if len(methods) == 0 {
		return nil
	}

	return Add(b.buildStruct(e, nestedBuilderFields, nestedVariantFields), Line().Line()).
		Add(b.buildRestImpl(e, rcv)...).
		Add(methods...)
}

func (b *builderBuilder) buildStruct(e *schema.Entity, nestedBuilderFields, nestedVariantFields []Code) Code {
	//	type <struct> struct {
	//		_ctx *<buildContext>
	//		_id depquery.ID
	//		_parentID depquery.ID
	//		_relations uint64
	//
	//		<nestedBuilder> ...
	//
	//		<variantBuilder> ...
	//	}
	return Type().Id(b.naming.BuilderStruct(e)).StructFunc(func(s *Group) {
		s.Id("_ctx").Op("*").Id(b.naming.BuildContextStruct())
		s.Id("_id").Add(libID)
		s.Id("_parentID").Add(libID)
		s.Id("_relations").Uint64()

		if len(nestedBuilderFields) > 0 {
			s.Line()

			for _, f := range nestedBuilderFields {
				s.Add(f)
			}
		}

		if len(nestedVariantFields) > 0 {
			s.Line()

			for _, f := range nestedVariantFields {
				s.Add(f)
			}
		}
	})
}

func (b *builderBuilder) buildRestImpl(e *schema.Entity, rcv Code) []Code {
	return []Code{
		b.buildConstructor(e),

		Line().Line(),

		// Compiler Impl

		//	func (b *<struct>) <compilerWith>(f func(<iface>) <compilerIface>[<iface>, <entityType] {
		//		if f != nil {
		//			f(b)
		//		}
		//
		//		return b
		//	}
		Func().Params(rcv).Id(b.naming.CompilerWithMethod()).
			Params(Id("f").Func().Params(Id(b.naming.BuilderInterface(e)))).
			Id(b.naming.CompilerInterface()).Types(Id(b.naming.BuilderInterface(e)), typeToJen(e.Type)).
			Block(
				If(Id("f").Op("!=").Nil()).Block(Id("f").Call(Id("b"))),
				Line(),
				Return(Id("b")),
			),

		Line().Line(),

		//	func (b *<struct>) <compilerBuilder>() <iface> {
		//		return b
		//	}
		Func().Params(rcv).Id(b.naming.CompilerBuilderMethod()).Params().
			Id(b.naming.BuilderInterface(e)).
			Block(Return(Id("b"))),

		Line().Line(),

		//	func (b *<struct>) <compilerCompile>() Plan[<entityType>] {
		//		return <planStruct>{
		//			ctxConstructor: <entityFetchConstr>,
		//			resolvers: b._ctx.plan(depquery.TopologicalPlanner{})
		//		}
		//	}
		Func().Params(rcv).Id(b.naming.CompilerCompileMethod()).Params().
			Id(b.naming.PlanInterface()).Types(typeToJen(e.Type)).
			Block(Return(
				Id(b.naming.PlanStruct()).Types(typeToJen(e.Type)).CustomFunc(multilineValuesOpts, func(s *Group) {
					s.Id("ctxConstructor").Op(":").Id(b.naming.FetchContextEntityConstructor(e))
					s.Id("resolvers").Op(":").
						Id("b").Dot("_ctx").Dot(b.naming.BuildContextPlanMethod()).Call(Qual(libPkg, "TopologicalPlanner").Values())
				}),
			)),

		Line().Line(),

		// End Compiler Impl

		// candidate Impl

		Func().Params(rcv).Id(b.naming.BuildContextCandidateCandidateMethod()).Params().Add(libCandidate).Block(
			Return(libCandidate.Clone().Values(
				Id("ID").Op(":").Id("b").Dot("_id"),
				Id("ParentID").Op(":").Id("b").Dot("_parentID"),
				Id("Relations").Op(":").Id("b").Dot("_relations"),
			)),
		),

		Line().Line(),

		Func().Params(rcv).Id(b.naming.BuildContextCandidateResolverMethod()).Params().Id(b.naming.ResolverFunc()).Block(
			Return(Id(b.naming.ResolverName(e)).Call(
				Id("b").Dot("_id"),
				Id("b").Dot("_parentID"),
				Id("b").Dot("_relations"),
			)),
		),

		Line().Line(),

		// End candidate Impl
	}
}

func (b *builderBuilder) buildVariantBuilder(e *schema.Entity, v schema.EntityVariant) Code {
	var (
		methods             = make([]Code, 0)
		nestedBuilderFields = make([]Code, 0)
	)

	rcvParam := Params(Id("b").Op("*").Id(b.naming.BuilderVariantStruct(e, v)))

	for _, r := range v.Relations {
		//	func (b *<struct>) <withMethodSig> {
		//		<withMethodBody>
		//	}
		m := Func().Add(rcvParam).Add(b.withMethodSignature(r, b.naming.BuilderVariantInterface(e, v))).
			Block(b.withMethodBody("b", b.naming.ResolverVariantRelConst(e, v, r))...).
			Line().Line()
		methods = append(methods, m)

		if len(r.Entity.Relations) == 0 {
			continue
		}

		//	func (b *<struct>) <deepMethodSig> {
		//		<deepMethodBody>
		//	}
		m = Func().Add(rcvParam).Add(b.deepMethodSignature(r, "f", b.naming.BuilderVariantInterface(e, v))).
			Block(b.deepMethodBody("b", "f", r)...).
			Line().Line()
		methods = append(methods, m)

		// <nestedBuilderField> <iface>
		nb := Id(b.naming.BuilderNestedBuilderField(r)).Id(b.naming.BuilderInterface(r.Entity))
		nestedBuilderFields = append(nestedBuilderFields, nb)
	}

	if len(methods) == 0 {
		return nil
	}

	//	type <struct> struct {
	//		*<parentBuilder>
	//
	//		<nestedBuilder> ...
	//	}
	s := Type().Id(b.naming.BuilderVariantStruct(e, v)).StructFunc(func(s *Group) {
		s.Op("*").Id(b.naming.BuilderStruct(e))

		if len(nestedBuilderFields) > 0 {
			s.Line()

			for _, f := range nestedBuilderFields {
				s.Add(f)
			}
		}
	})

	return Add(s, Line().Line(), b.buildVariantConstructor(e, v), Line().Line()).Add(methods...)
}

func (b *builderBuilder) buildConstructor(e *schema.Entity) Code {
	return Func().Id(b.naming.BuilderConstructor(e)).
		Params(
			Id("ctx").Op("*").Id(b.naming.BuildContextStruct()),
			Id("pid").Add(libID),
		).
		Op("*").Id(b.naming.BuilderStruct(e)).
		BlockFunc(func(body *Group) {
			body.Id("b").Op(":=").Op("&").Id(b.naming.BuilderStruct(e)).Values(
				Id("_ctx").Op(":").Id("ctx"),
				Id("_id").Op(":").Id("ctx").Dot("nextID").Call(),
				Id("_parentID").Op(":").Id("pid"),
			)
			body.Id("ctx").Dot(b.naming.BuildContextAddCandidateMethod()).Call(Id("b").Dot("_id"), Id("b"))
			body.Line()

			body.Line().Return(Id("b"))
		})
}

func (b *builderBuilder) buildVariantConstructor(e *schema.Entity, v schema.EntityVariant) Code {
	return Func().Id(b.naming.BuilderVariantConstructor(e, v)).
		Params(Id("parent").Op("*").Id(b.naming.BuilderStruct(e))).
		Op("*").Id(b.naming.BuilderVariantStruct(e, v)).
		Block(
			Return().Op("&").Id(b.naming.BuilderVariantStruct(e, v)).Values(Id(b.naming.BuilderStruct(e)).Op(":").Id("parent")),
		)
}

func (b *builderBuilder) withMethodSignature(r schema.Relation, builder string) Code {
	return Id(b.naming.BuilderWithMethod(r)).Params().Id(builder)
}

func (b *builderBuilder) withMethodBody(rcv, constName string) []Code {
	return []Code{
		Id(rcv).Dot("_relations").Op("|=").Id(constName),
		Line(),
		Return(Id("b")),
	}
}

func (b *builderBuilder) deepMethodSignature(r schema.Relation, funcParam, builder string) Code {
	return Id(b.naming.BuilderDeepMethod(r)).
		Params(Id(funcParam).Func().Params(Id(b.naming.BuilderInterface(r.Entity)))).
		Id(builder)
}

func (b *builderBuilder) deepMethodBody(rcv, funcParam string, r schema.Relation) []Code {
	return []Code{
		Id(rcv).Dot(b.naming.BuilderWithMethod(r)).Call(),
		Line(),

		If(Id(funcParam).Op("==").Nil()).Block(Return(Id("b"))),
		Line(),

		If(Id(rcv).Dot(b.naming.BuilderNestedBuilderField(r)).Op("==").Nil()).Block(
			Id(rcv).Dot(b.naming.BuilderNestedBuilderField(r)).Op("=").
				Id(b.naming.BuilderConstructor(r.Entity)).Call(Id(rcv).Dot("_ctx"), Id(rcv).Dot("_id")),
		),
		Line(),

		Id(funcParam).Call(Id(rcv).Dot(b.naming.BuilderNestedBuilderField(r))),
		Line(),

		Return(Id(rcv)),
	}
}

func (b *builderBuilder) variantMethodSignature(
	e *schema.Entity,
	v schema.EntityVariant,
	funcParam, builder string,
) Code {
	return Id(b.naming.BuilderVariantMethod(v)).
		Params(Id(funcParam).Func().Params(Id(b.naming.BuilderVariantInterface(e, v)))).
		Id(builder)
}

func (b *builderBuilder) variantMethodBody(rcv, funcParam string, e *schema.Entity, v schema.EntityVariant) []Code {
	return []Code{
		If(Id(funcParam).Op("==").Nil()).Block(Return(Id(rcv))),
		Line(),

		If(Id(rcv).Dot(b.naming.BuilderNestedVariantBuilderField(v)).Op("==").Nil()).Block(
			Id("b").Dot(b.naming.BuilderNestedVariantBuilderField(v)).Op("=").
				Id(b.naming.BuilderVariantConstructor(e, v)).Call(Id(rcv)),
		),
		Line(),

		Id(funcParam).Call(Id(rcv).Dot(b.naming.BuilderNestedVariantBuilderField(v))),
		Line(),

		Return(Id(rcv)),
	}
}
