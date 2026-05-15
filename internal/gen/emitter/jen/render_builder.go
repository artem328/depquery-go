package jen

import (
	"fmt"

	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderBuilderInterfaces() {
	for _, b := range r.plan.Builders {
		r.renderBuilderInterface(b)
	}
}

func (r *Renderer) renderBuilderInterface(b plan.Builder) {
	r.renderBuilderInterfaceDefinition(b)

	r.renderBuilderEntrypoint(b)
}

func (r *Renderer) renderBuilderInterfaceDefinition(b plan.Builder) {
	methods := make([]Code, 0, len(b.GetMethods()))
	for _, m := range b.GetMethods() {
		methods = append(methods, r.generateBuilderMethodSignature(b, m, ""))
	}

	r.f.Add(block(
		Type().Id(r.naming.Builder.Interface[b.GetID()]).Interface(methods...),
	))
}

func (r *Renderer) renderBuilderEntrypoint(b plan.Builder) {
	rb, ok := b.(plan.RootBuilder)
	if !ok {
		return
	}

	e := r.plan.Model.Entities[rb.Entity]

	builderIfaceType := Id(r.naming.Builder.Interface[rb.GetID()])
	const funcVar = "f"

	builder := r.generateRootBuilderConstructorCall(rb.GetID(), r.generateBuildContextConstructorCall(), Lit(0))

	r.f.Add(block(
		Func().Id(r.naming.Builder.Entrypoint[rb.GetID()]).Params(
			r.generateBuilderCallbackParam(builderIfaceType, funcVar),
		).Params(
			r.generateCompilerInterfaceType(builderIfaceType, r.types[e.Type]),
		).Block(
			Return(r.generateCompilerWithMethodCall(builder, Id(funcVar))),
		),
	))
}

func (r *Renderer) generateBuilderMethodSignature(b plan.Builder, m plan.BuilderMethod, funcVar string) Code {
	method := Id(r.naming.Builder.Method[m.GetID()])

	switch mm := m.(type) {
	case plan.EnableBuilderMethod:
		method.Params()
	case plan.DeepBuilderMethod:
		method.Params(r.generateBuilderCallbackParam(Id(r.naming.Builder.Interface[mm.ChildBuilder]), funcVar))
	case plan.VariantBuilderMethod:
		method.Params(r.generateBuilderCallbackParam(Id(r.naming.Builder.Interface[mm.ChildBuilder]), funcVar))
	default:
		panic(fmt.Errorf("unknown builder method type: %T", m))
	}

	return method.Params(Id(r.naming.Builder.Interface[b.GetID()]))
}

func (r *Renderer) renderBuilderImplementations() {
	for _, b := range r.plan.Builders {
		r.renderBuilderImplementation(b)
	}
}

func (r *Renderer) renderBuilderImplementation(b plan.Builder) {
	r.renderBuilderStructDefinition(b)
	r.renderBuilderConstructor(b)

	rcvVar := "b"

	if rb, ok := b.(plan.RootBuilder); ok {
		r.renderBuilderCompilerImplementation(rb, rcvVar)
		r.renderBuilderCandidateImplementation(rb, rcvVar)
	}

	for _, m := range b.GetMethods() {
		r.renderBuilderMethod(b, m, rcvVar)
	}
}

func (r *Renderer) renderBuilderStructDefinition(b plan.Builder) {
	fields := r.generateBuilderBaseFieldsDefinition(b)

	if len(fields) > 0 {
		fields = append(fields, Empty())
	}

	for _, cb := range b.GetChildBuilders() {
		fields = append(fields, r.generateBuilderChildBuilderFieldDefinition(cb))
	}

	r.f.Add(block(
		Type().Id(r.naming.Builder.Struct[b.GetID()]).Struct(fields...),
	))
}

func (r *Renderer) generateBuilderBaseFieldsDefinition(b plan.Builder) []Code {
	var fields []Code

	switch bb := b.(type) {
	case plan.RootBuilder:
		fields = append(fields,
			Id(r.naming.Builder.FieldContext).Op("*").Id(r.naming.BuildContext.Type),
			Id(r.naming.Builder.FieldID).Add(libID),
			Id(r.naming.Builder.FieldParentID).Add(libID),
			Id(r.naming.Builder.FieldRelations).Id("uint64"),
		)
	case plan.VariantBuilder:
		fields = append(fields,
			Op("*").Id(r.naming.Builder.Struct[bb.Parent]),
		)
	default:
		panic(fmt.Errorf("unknown builder type: %T", b))
	}

	return fields
}

func (r *Renderer) generateBuilderChildBuilderFieldDefinition(cb plan.ChildBuilder) Code {
	var builderName string
	switch cbb := cb.(type) {
	case plan.RegularChildBuilder:
		builderName = r.naming.Builder.ChildBuilder[cbb.Builder]
	case plan.RelationChildBuilder:
		builderName = r.naming.Relation.ChildBuilder[cbb.Relation]
	default:
		panic(fmt.Errorf("unknown child builder type: %T", cb))
	}

	return Id(builderName).Id(r.naming.Builder.Interface[cb.GetBuilderID()])
}

func (r *Renderer) renderBuilderConstructor(b plan.Builder) {
	params, args := r.generateBuilderConstructorParams(b)

	r.f.Add(block(
		Func().Id(r.naming.Builder.Constructor[b.GetID()]).Params(params...).Params(Op("*").Id(r.naming.Builder.Struct[b.GetID()])).Block(
			r.generateBuilderConstructorBody(b, args)...,
		),
	))
}

func (r *Renderer) generateRootBuilderConstructorCall(bid plan.BuilderID, ctxArg, parentIdArg Code) Code {
	return Id(r.naming.Builder.Constructor[bid]).Call(ctxArg, parentIdArg)
}

func (r *Renderer) generateVariantBuilderConstructorCall(bid plan.BuilderID, parentArg Code) Code {
	return Id(r.naming.Builder.Constructor[bid]).Call(parentArg)
}

func (r *Renderer) generateBuilderConstructorParams(b plan.Builder) ([]Code, []Code) {
	switch bb := b.(type) {
	case plan.RootBuilder:
		argCtx := Id("ctx")
		argParentId := Id("pid")

		return []Code{
			Add(argCtx).Op("*").Id(r.naming.BuildContext.Type),
			Add(argParentId).Add(libID),
		}, []Code{argCtx, argParentId}
	case plan.VariantBuilder:
		argParentBuilder := Id("parent")
		return []Code{
			Add(argParentBuilder).Op("*").Id(r.naming.Builder.Struct[bb.Parent]),
		}, []Code{argParentBuilder}
	default:
		panic(fmt.Errorf("unknown builder type: %T", bb))
	}
}

func (r *Renderer) generateBuilderConstructorBody(b plan.Builder, args []Code) []Code {
	switch bb := b.(type) {
	case plan.RootBuilder:
		if len(args) != 2 {
			panic(fmt.Errorf("unexpected number of arguments for root builder constructor: %d, expected 2", len(args)))
		}

		return r.generateRootBuilderConstructorBody(bb, args[0], args[1])
	case plan.VariantBuilder:
		if len(args) != 1 {
			panic(fmt.Errorf("unexpected number of arguments for variant builder constructor: %d, expected 1", len(args)))
		}

		return r.generateVariantBuilderConstructorBody(bb, args[0])
	default:
		panic(fmt.Errorf("unknown builder type: %T", bb))
	}
}

func (r *Renderer) generateRootBuilderConstructorBody(b plan.RootBuilder, ctxArg, parentIdArg Code) []Code {
	builder := Op("&").Id(r.naming.Builder.Struct[b.GetID()]).Values(
		Id(r.naming.Builder.FieldContext).Op(":").Add(ctxArg),
		Id(r.naming.Builder.FieldID).Op(":").Add(r.generateBuildContextNextIDMethodCall(ctxArg)),
		Id(r.naming.Builder.FieldParentID).Op(":").Add(parentIdArg),
	)

	builderVar := Id("b")

	return []Code{
		Add(builderVar).Op(":=").Add(builder),
		Add(r.generateBuildContextAddCandidateMethodCall(ctxArg, Add(builderVar).Dot(r.naming.Builder.FieldID), builderVar)),
		Empty(),
		Return(builderVar),
	}
}

func (r *Renderer) generateVariantBuilderConstructorBody(b plan.VariantBuilder, parentArg Code) []Code {
	return []Code{
		Return(Op("&").Id(r.naming.Builder.Struct[b.GetID()]).Values(
			Id(r.naming.Builder.Struct[b.Parent]).Op(":").Add(parentArg),
		)),
	}
}

func (r *Renderer) renderBuilderCompilerImplementation(b plan.RootBuilder, rcvVar string) {
	e := r.plan.Model.Entities[b.Entity]

	const funcVar = "f"
	rcv := Id(rcvVar)
	f := Id(funcVar)
	typeParamBuilder := Id(r.naming.Builder.Interface[b.GetID()])
	typeParamRoot := r.types[e.Type]

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateCompilerWithMethodSignature(typeParamBuilder, typeParamRoot, funcVar)).Block(
			r.generateCompilerWithMethodBody(rcv, f)...,
		),
	))

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateCompilerBuilderMethodSignature(typeParamBuilder)).Block(r.generateCompilerBuilderMethodBody(rcv)...),
	))

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateCompilerCompileMethodSignature(typeParamRoot)).Block(r.generateCompilerCompileMethodBody(typeParamRoot, Id(r.naming.FetchContext.ConstructorForEntity[b.FetchContextRoot]), Add(rcv).Dot(r.naming.Builder.FieldContext))...),
	))
}

func (r *Renderer) renderBuilderCandidateImplementation(b plan.RootBuilder, rcvVar string) {
	rcv := Id(rcvVar)

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateCandidateCandidateMethodSignature()).Block(
			r.generateCandidateCandidateMethodBody(rcv)...,
		),
	))

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateCandidateResolverMethodSignature()).Block(
			r.generateCandidateResolverMethodBody(b.EntityResolver, rcv)...,
		),
	))
}

func (r *Renderer) renderBuilderMethod(b plan.Builder, m plan.BuilderMethod, rcvVar string) {
	const funcVar = "f"

	rcv := Id(rcvVar)
	childRcv := Id(funcVar)

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateBuilderMethodSignature(b, m, funcVar)).Block(
			r.generateBuilderMethodBody(m, rcv, childRcv)...,
		),
	))
}

func (r *Renderer) generateBuilderMethodBody(m plan.BuilderMethod, rcv, childRcv Code) []Code {
	switch mm := m.(type) {
	case plan.EnableBuilderMethod:
		return r.generateBuilderEnableMethodBody(mm, rcv)
	case plan.DeepBuilderMethod:
		return r.generateBuilderDeepMethodBody(mm, rcv, childRcv)
	case plan.VariantBuilderMethod:
		return r.generateBuilderVariantMethodBody(mm, rcv, childRcv)
	default:
		panic(fmt.Errorf("unknown builder method type: %T", m))
	}
}

func (r *Renderer) generateBuilderEnableMethodBody(m plan.EnableBuilderMethod, rcv Code) []Code {
	return []Code{
		Add(rcv).Dot(r.naming.Builder.FieldRelations).Op("|=").Id(r.naming.Relation.ConstantName[m.Relation]),
		Empty(),
		Return(rcv),
	}
}

func (r *Renderer) generateBuilderEnableMethodCall(mid plan.BuilderMethodID, rcv Code) Code {
	return Add(rcv).Dot(r.naming.Builder.Method[mid]).Call()
}

func (r *Renderer) generateBuilderDeepMethodBody(m plan.DeepBuilderMethod, rcv, childRcv Code) []Code {
	return []Code{
		r.generateBuilderEnableMethodCall(m.EnableMethodID, rcv),
		Empty(),
		If(Add(childRcv).Op("==").Nil()).Block(
			Return(rcv),
		),
		Empty(),
		If(Add(rcv).Dot(r.naming.Relation.ChildBuilder[m.Relation]).Op("==").Nil()).Block(
			Add(rcv).Dot(r.naming.Relation.ChildBuilder[m.Relation]).Op("=").Add(r.generateRootBuilderConstructorCall(m.ChildBuilder, Add(rcv).Dot(r.naming.Builder.FieldContext), Add(rcv).Dot(r.naming.Builder.FieldID))),
		),
		Empty(),
		Add(childRcv).Call(Add(rcv).Dot(r.naming.Relation.ChildBuilder[m.Relation])),
		Empty(),
		Return(rcv),
	}
}

func (r *Renderer) generateBuilderVariantMethodBody(m plan.VariantBuilderMethod, rcv, childRcv Code) []Code {
	return []Code{
		If(Add(childRcv).Op("==").Nil()).Block(
			Return(rcv),
		),
		Empty(),
		If(Add(rcv).Dot(r.naming.Builder.ChildBuilder[m.ChildBuilder]).Op("==").Nil()).Block(
			Add(rcv).Dot(r.naming.Builder.ChildBuilder[m.ChildBuilder]).Op("=").Add(r.generateVariantBuilderConstructorCall(m.ChildBuilder, rcv)),
		),
		Empty(),
		Add(childRcv).Call(Add(rcv).Dot(r.naming.Builder.ChildBuilder[m.ChildBuilder])),
		Empty(),
		Return(rcv),
	}
}

func (r *Renderer) generateBuilderCallbackParam(builderType Code, funcVar string) Code {
	return Id(funcVar).Func().Params(builderType)
}

func (r *Renderer) generateBuilderMethodBase(b plan.Builder, rcv Code) Code {
	return Func().Params(Add(rcv).Op("*").Id(r.naming.Builder.Struct[b.GetID()]))
}
