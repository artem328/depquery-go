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
		methods = append(methods, r.generateBuilderMethodSignature(b, m, Null()))
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
	funcVar := Id("f")

	builder := r.generateRootBuilderConstructorCall(rb.GetID(), r.generateBuildContextConstructorCall(), Lit(0))

	r.f.Add(block(
		Func().Id(r.naming.Builder.Entrypoint[rb.GetID()]).Params(
			r.generateBuilderCallbackParam(builderIfaceType, funcVar),
		).Params(
			r.generateCompilerInterfaceType(builderIfaceType, r.types[e.Type]),
		).Block(
			Return(r.generateCompilerWithMethodCall(builder, funcVar)),
		),
	))
}

func (r *Renderer) generateBuilderMethodSignature(b plan.Builder, m plan.BuilderMethod, funcVar Code) Code {
	method := Id(r.naming.Builder.Method[m.GetID()])

	switch mm := m.(type) {
	case plan.EnableBuilderMethod:
		method.Params()
	case plan.DeepBuilderMethod:
		method.Params(r.generateBuilderCallbackParam(Id(r.naming.Builder.Interface[mm.ChildBuilder]), funcVar))
	case plan.VariantBuilderMethod:
		method.Params(r.generateBuilderCallbackParam(Id(r.naming.Builder.Interface[mm.ChildBuilder]), funcVar))
	case plan.NestedBuilderMethod:
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

	rcv := Id("b")

	if rb, ok := b.(plan.RootBuilder); ok {
		r.renderBuilderCompilerImplementation(rb, rcv)
	}

	for _, m := range b.GetMethods() {
		r.renderBuilderMethod(b, m, rcv)
	}

	if rb, ok := b.(plan.RootBuilder); ok {
		r.renderBuilderCandidateImplementation(rb)
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
			Id(r.naming.Builder.FieldParentID).Add(libID),
		)

		if bb.IsRelationBuilder {
			fields = append(fields, Id(r.naming.Builder.FieldID).Add(libID))
		}

		if bb.IsNestedBuilder {
			fields = append(fields, Id(r.naming.Builder.FieldNestedID).Add(libID))
		}

		if bb.IsRelationBuilder {
			fields = append(fields, Id(r.naming.Builder.FieldRelations).Uint64())
		}

		if bb.IsNestedBuilder {
			fields = append(fields, Id(r.naming.Builder.FieldNested).Uint64())
		}
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
	case plan.NestedChildBuilder:
		builderName = r.naming.Nested.ChildBuilder[cbb.Nested]
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
	var fields []Code

	fields = append(fields,
		Id(r.naming.Builder.FieldContext).Op(":").Add(ctxArg),
		Id(r.naming.Builder.FieldParentID).Op(":").Add(parentIdArg),
	)

	if b.IsRelationBuilder {
		fields = append(fields, Id(r.naming.Builder.FieldID).Op(":").Add(r.generateBuildContextNextIDMethodCall(ctxArg)))
	}

	if b.IsNestedBuilder {
		fields = append(fields, Id(r.naming.Builder.FieldNestedID).Op(":").Add(r.generateBuildContextNextIDMethodCall(ctxArg)))
	}

	builder := Op("&").Id(r.naming.Builder.Struct[b.GetID()]).Add(valuesMultiline(fields...))

	builderVar := Id("b")

	var body []Code

	body = append(body, Add(builderVar).Op(":=").Add(builder))
	body = append(body, Empty())
	if b.IsRelationBuilder {
		body = append(body, r.generateBuildContextAddCandidateMethodCall(ctxArg, Add(builderVar).Dot(r.naming.Builder.FieldID), Id(r.naming.Candidate.RelationStruct[b.ID]).Values(builderVar)))
	}
	if b.IsNestedBuilder {
		body = append(body, r.generateBuildContextAddCandidateMethodCall(ctxArg, Add(builderVar).Dot(r.naming.Builder.FieldNestedID), Id(r.naming.Candidate.NestedStruct[b.ID]).Values(builderVar)))
	}
	body = append(body, Empty(), Return(builderVar))

	return body
}

func (r *Renderer) generateVariantBuilderConstructorBody(b plan.VariantBuilder, parentArg Code) []Code {
	return []Code{
		Return(Op("&").Id(r.naming.Builder.Struct[b.GetID()]).Values(
			Id(r.naming.Builder.Struct[b.Parent]).Op(":").Add(parentArg),
		)),
	}
}

func (r *Renderer) renderBuilderCompilerImplementation(b plan.RootBuilder, rcv Code) {
	e := r.plan.Model.Entities[b.Entity]

	funcVar := Id("f")
	typeParamBuilder := Id(r.naming.Builder.Interface[b.GetID()])
	typeParamRoot := r.types[e.Type]
	opts := Id("opts")

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateCompilerWithMethodSignature(typeParamBuilder, typeParamRoot, funcVar)).Block(
			r.generateCompilerWithMethodBody(rcv, funcVar)...,
		),
	))

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateCompilerBuilderMethodSignature(typeParamBuilder)).Block(r.generateCompilerBuilderMethodBody(rcv)...),
	))

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateCompilerCompileMethodSignature(typeParamRoot, opts)).Block(
			r.generateCompilerCompileMethodBody(typeParamRoot, opts, Id(r.naming.FetchContext.ConstructorForEntity[b.FetchContextRoot]), Add(rcv).Dot(r.naming.Builder.FieldContext))...),
	))
}

func (r *Renderer) renderBuilderCandidateImplementation(b plan.RootBuilder) {
	if b.IsRelationBuilder {
		r.renderBuilderRelationsCandidateImplementation(b)
	}

	if b.IsNestedBuilder {
		r.renderBuilderNestedCandidateImplementation(b)
	}
}

func (r *Renderer) renderBuilderRelationsCandidateImplementation(b plan.RootBuilder) {
	const builderField = "b"

	r.f.Add(block(
		r.generateBuilderRelationsCandidateStruct(b, builderField),
	))

	rcv := Id("c")

	r.f.Add(block(
		Add(r.generateBuilderCandidateMethodBase(r.naming.Candidate.RelationStruct[b.ID], rcv), r.generateCandidateCandidateMethodSignature()).Block(
			r.generateCandidateRelationCandidateMethodBody(Add(rcv).Dot(builderField))...,
		),
	))

	r.f.Add(block(
		Add(r.generateBuilderCandidateMethodBase(r.naming.Candidate.RelationStruct[b.ID], rcv), r.generateCandidateResolverMethodSignature()).Block(
			r.generateCandidateRelationsResolverMethodBody(b.EntityResolver, Add(rcv).Dot(builderField))...,
		),
	))
}

func (r *Renderer) renderBuilderNestedCandidateImplementation(b plan.RootBuilder) {
	const builderField = "b"

	r.f.Add(block(
		r.generateBuilderNestedCandidateStruct(b, builderField),
	))

	rcv := Id("c")

	r.f.Add(block(
		Add(r.generateBuilderCandidateMethodBase(r.naming.Candidate.NestedStruct[b.ID], rcv), r.generateCandidateCandidateMethodSignature()).Block(
			r.generateCandidateNestedCandidateMethodBody(Add(rcv).Dot(builderField))...,
		),
	))

	r.f.Add(block(
		Add(r.generateBuilderCandidateMethodBase(r.naming.Candidate.NestedStruct[b.ID], rcv), r.generateCandidateResolverMethodSignature()).Block(
			r.generateCandidateNestedResolverMethodBody(b.NestedResolver, Add(rcv).Dot(builderField))...,
		),
	))
}

func (r *Renderer) generateBuilderRelationsCandidateStruct(b plan.RootBuilder, builderField string) Code {
	return Type().Id(r.naming.Candidate.RelationStruct[b.ID]).Struct(
		Id(builderField).Op("*").Id(r.naming.Builder.Struct[b.ID]),
	)
}

func (r *Renderer) generateBuilderNestedCandidateStruct(b plan.RootBuilder, builderField string) Code {
	return Type().Id(r.naming.Candidate.NestedStruct[b.ID]).Struct(
		Id(builderField).Op("*").Id(r.naming.Builder.Struct[b.ID]),
	)
}

func (r *Renderer) generateBuilderCandidateMethodBase(structName string, rcv Code) Code {
	return Func().Params(Add(rcv).Id(structName))
}

func (r *Renderer) renderBuilderMethod(b plan.Builder, m plan.BuilderMethod, rcv Code) {
	childRcv := Id("f")

	r.f.Add(block(
		Add(r.generateBuilderMethodBase(b, rcv), r.generateBuilderMethodSignature(b, m, childRcv)).Block(
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
	case plan.NestedBuilderMethod:
		return r.generateBuilderNestedMethodBody(mm, rcv, childRcv)
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

func (r *Renderer) generateBuilderNestedMethodBody(m plan.NestedBuilderMethod, rcv, childRcv Code) []Code {
	return []Code{
		Add(rcv).Dot(r.naming.Builder.FieldNested).Op("|=").Id(r.naming.Nested.ConstantName[m.Nested]),
		Empty(),
		If(Add(rcv).Dot(r.naming.Nested.ChildBuilder[m.Nested]).Op("==").Nil()).Block(
			Add(rcv).Dot(r.naming.Nested.ChildBuilder[m.Nested]).Op("=").Add(r.generateRootBuilderConstructorCall(m.ChildBuilder, Add(rcv).Dot(r.naming.Builder.FieldContext), Add(rcv).Dot(r.naming.Builder.FieldNestedID))),
		),
		Empty(),
		If(Add(childRcv).Op("!=").Nil()).Block(
			Add(childRcv).Call(Add(rcv).Dot(r.naming.Nested.ChildBuilder[m.Nested])),
		),
		Empty(),
		Return(rcv),
	}
}

func (r *Renderer) generateBuilderCallbackParam(builderType Code, funcParam Code) Code {
	return Add(funcParam).Func().Params(builderType)
}

func (r *Renderer) generateBuilderMethodBase(b plan.Builder, rcv Code) Code {
	return Func().Params(Add(rcv).Op("*").Id(r.naming.Builder.Struct[b.GetID()]))
}
