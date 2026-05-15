package jen

import (
	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderFetchContextImplementation() {
	r.renderFetchContextStruct()
	r.renderFetchContextConstructor()
	r.renderFetchContextInstanceConstructors()

	rcv := Id("ctx")

	r.renderFetchContextEnqueueMethods(rcv)
	r.renderFetchContextStateOverrides(rcv)
	r.renderFetchContextParentGetterMethods(rcv)
	r.renderFetchContextFlushMethod(rcv)
}

func (r *Renderer) renderFetchContextStruct() {
	r.f.Add(block(
		Type().Id(r.naming.FetchContext.Struct).Struct(
			r.generateFetchContextStructFields()...,
		),
	))
}

func (r *Renderer) generateFetchContextStructFields() []Code {
	var fields []Code

	fields = append(fields, Id(r.naming.State.Struct))

	if len(r.plan.FetchParents) > 0 {
		fields = append(fields, Empty())

		for _, fp := range r.plan.FetchParents {
			fields = append(fields, Id(r.naming.FetchContext.FieldByParentID[fp.ID]).Add(r.generateFetchContextParentFieldType(fp)))
		}
	}

	if len(r.plan.FetchChildren) > 0 {
		fields = append(fields, Empty())

		for _, fc := range r.plan.FetchChildren {
			fields = append(fields, Id(r.naming.FetchContext.FieldChildSeen[fc.ID]).Add(r.generateFetchContextSeenFieldType(fc)))
		}
	}

	if len(r.plan.FetchChildren) > 0 {
		fields = append(fields, Empty())

		for _, fc := range r.plan.FetchChildren {
			fields = append(fields, Id(r.naming.FetchContext.FieldPending[fc.ID]).Add(r.generateFetchContextPendingFieldType(fc)))
		}
	}

	return fields
}

func (r *Renderer) generateFetchContextParentFieldType(fp plan.FetchParent) Code {
	return Map(libID).Add(r.generateFetchContextParentLevelSetType(fp))
}

func (r *Renderer) generateFetchContextParentLevelSetType(fp plan.FetchParent) Code {
	e := r.plan.Model.Entities[fp.Entity]
	if fp.Reversed {
		e = r.plan.Model.Entities[fp.ReversedByEntity]
	}

	im := r.plan.Model.Members[e.IDMember]

	return Map(r.types[im.Type]).Struct()
}

func (r *Renderer) generateFetchContextSeenFieldType(fc plan.FetchChild) Code {
	e := r.plan.Model.Entities[fc.Entity]
	if fc.Reversed {
		e = r.plan.Model.Entities[fc.ReversedByEntity]
	}

	im := r.plan.Model.Members[e.IDMember]

	return Map(r.types[im.Type]).Struct()
}

func (r *Renderer) generateFetchContextPendingFieldType(fc plan.FetchChild) Code {
	e := r.plan.Model.Entities[fc.Entity]
	if fc.Reversed {
		e = r.plan.Model.Entities[fc.ReversedByEntity]
	}

	im := r.plan.Model.Members[e.IDMember]

	return Index().Add(r.types[im.Type])
}

func (r *Renderer) renderFetchContextConstructor() {
	var inits []Code

	inits = append(inits, Id(r.naming.State.Struct).Op(":").Add(r.generateStateConstructorCall()))

	for _, fp := range r.plan.FetchParents {
		inits = append(inits, Id(r.naming.FetchContext.FieldByParentID[fp.ID]).Op(":").Make(r.generateFetchContextParentFieldType(fp)))
	}

	for _, fc := range r.plan.FetchChildren {
		inits = append(inits, Id(r.naming.FetchContext.FieldChildSeen[fc.ID]).Op(":").Make(r.generateFetchContextSeenFieldType(fc)))
	}

	r.f.Add(block(
		Func().Id(r.naming.FetchContext.Constructor).Params().Params(Op("*").Id(r.naming.FetchContext.Struct)).Block(
			Return(Op("&").Id(r.naming.FetchContext.Struct).Add(valuesMultiline(inits...))),
		),
	))
}

func (r *Renderer) generateFetchContextConstructorCall() Code {
	return Id(r.naming.FetchContext.Constructor).Call()
}

func (r *Renderer) generateFetchContextConstructorForEntityType(typeParamRoot Code) Code {
	return Func().Add(r.generateFetchContextConstructorForEntitySignatureParams(Null(), typeParamRoot))
}

func (r *Renderer) generateFetchContextConstructorForEntityCall(rcv, iter Code) Code {
	return Add(rcv).Call(iter)
}

func (r *Renderer) generateFetchContextConstructorForEntitySignatureParams(iterParam, typeParamRoot Code) Code {
	return Params(Add(iterParam, iterSeq).Types(typeParamRoot)).Params(Op("*").Id(r.naming.FetchContext.Struct))
}

func (r *Renderer) generateFetchContextConstructorForEntitySignature(ef plan.FetchContextRoot, iterParam Code) Code {
	e := r.plan.Model.Entities[ef.Entity]

	return Id(r.naming.FetchContext.ConstructorForEntity[ef.ID]).Add(r.generateFetchContextConstructorForEntitySignatureParams(iterParam, r.types[e.Type]))
}

func (r *Renderer) renderFetchContextInstanceConstructors() {
	for _, ef := range r.plan.FetchContextRoots {
		r.renderFetchContextInstanceConstructor(ef)
	}
}

func (r *Renderer) renderFetchContextInstanceConstructor(fcr plan.FetchContextRoot) {
	e := r.plan.Model.Entities[fcr.Entity]

	iter := Id("i")
	fetchContext := Id("ctx")
	zeroLevel := Id("zero")
	entity := Id("x")

	r.f.Add(block(
		Func().Add(r.generateFetchContextConstructorForEntitySignature(fcr, iter)).Block(
			Add(fetchContext).Op(":=").Add(r.generateFetchContextConstructorCall()),

			Empty(),

			Add(zeroLevel).Op(":=").Make(r.generateFetchContextParentLevelSetType(r.plan.FetchParents[fcr.FetchParent])),
			For(Add(entity).Op(":=").Range().Add(iter)).Block(
				Add(zeroLevel).Index(r.members.Member(entity, e.IDMember)).Op("=").Struct().Values(),
				Add(fetchContext).Dot(r.naming.State.StateContainerField[fcr.StateContainer]).Index(r.members.Member(entity, e.IDMember)).Op("=").Add(libFetched).Call(entity),
			),

			Empty(),

			Add(fetchContext).Dot(r.naming.FetchContext.FieldByParentID[fcr.FetchParent]).Index(Lit(0)).Op("=").Add(zeroLevel),

			Empty(),

			Return(fetchContext),
		),
	))
}

func (r *Renderer) generateFetchContextMethodBase(rcv Code) Code {
	return Func().Params(Add(rcv).Op("*").Id(r.naming.FetchContext.Struct))
}

func (r *Renderer) renderFetchContextEnqueueMethods(rcv Code) {
	for _, ef := range r.plan.EntityFetches {
		r.renderFetchContextEnqueueMethod(ef, rcv)
	}
}

func (r *Renderer) generateFetchContextEnqueueMethodSignature(ef plan.EntityFetch, idParam, relationIDParam Code) Code {
	e := r.plan.Model.Entities[ef.Entity]
	if ef.Reversed {
		c := r.plan.FetchChildren[ef.Child]
		e = r.plan.Model.Entities[c.ReversedByEntity]
	}

	im := r.plan.Model.Members[e.IDMember]

	return Id(r.naming.FetchContext.EnqueueMethod[ef.ID]).Params(Add(idParam, r.types[im.Type]), Add(relationIDParam, libID))
}

func (r *Renderer) generateFetchContextEnqueueMethodCall(efid plan.EntityFetchID, rcv, idArg, entityLevelIDArg Code) Code {
	return Add(rcv).Dot(r.naming.FetchContext.EnqueueMethod[efid]).Call(idArg, entityLevelIDArg)
}

func (r *Renderer) generateFetchContextEnqueueMethodBody(ef plan.EntityFetch, rcv, idArg, relationIDArg Code) []Code {
	var statements []Code

	ok := Id("ok")
	swallow := Id("_")

	if ef.IsParent {
		p := r.plan.FetchParents[ef.Parent]
		byParentIDSet := Add(rcv).Dot(r.naming.FetchContext.FieldByParentID[ef.Parent]).Index(relationIDArg)

		statements = append(statements,
			If(List(swallow, ok).Op(":=").Add(byParentIDSet), Op("!").Add(ok)).Block(
				Add(byParentIDSet).Op("=").Add(Make(r.generateFetchContextParentLevelSetType(p))),
			),
			Empty(),
			Add(byParentIDSet).Index(idArg).Op("=").Struct().Values(),
		)
	}

	if len(statements) > 0 {
		statements = append(statements, Empty())
	}

	var maybeFetchedID Code
	if ef.Reversed {
		maybeFetchedID = Add(rcv).Dot(r.naming.State.StateContainerReversedByField[ef.ReversedStateContainer]).Index(idArg)
	} else {
		maybeFetchedID = idArg
	}

	statements = append(statements,
		If(
			List(swallow, ok).Op(":=").Add(rcv).Dot(r.naming.FetchContext.FieldChildSeen[ef.Child]).Index(idArg),
			Add(ok).Op("||").Add(rcv).Dot(r.naming.State.StateContainerField[ef.StateContainer]).Index(maybeFetchedID).Dot("Fetched").Call(),
		).Block(
			Return(),
		),
		Empty(),
		Add(rcv).Dot(r.naming.FetchContext.FieldChildSeen[ef.Child]).Index(idArg).Op("=").Struct().Values(),
		appendSlice(Add(rcv).Dot(r.naming.FetchContext.FieldPending[ef.Child]), idArg),
	)

	return statements
}

func (r *Renderer) renderFetchContextEnqueueMethod(ef plan.EntityFetch, rcv Code) {
	id := Id("id")
	relationID := Id("rid")

	r.f.Add(block(
		Add(r.generateFetchContextMethodBase(rcv), r.generateFetchContextEnqueueMethodSignature(ef, id, relationID)).Block(r.generateFetchContextEnqueueMethodBody(ef, rcv, id, relationID)...),
	))
}

func (r *Renderer) renderFetchContextStateOverrides(rcv Code) {
	for _, rfp := range r.plan.ReversedFetchParents {
		r.renderFetchContextStateOverride(rfp, rcv)
	}
}

func (r *Renderer) renderFetchContextStateOverride(rfp plan.ReversedFetchParent, rcv Code) {
	fp := r.plan.FetchParents[rfp.FetchParent]
	sc := r.plan.StateContainers[rfp.StateContainer]

	entity := Id("e")

	var statements []Code

	statements = append(statements, r.generateStateAdderMethodCall(sc.ID, Add(rcv).Dot(r.naming.State.Struct), entity))

	for _, fpr := range rfp.ReversedFetchParents {
		statements = append(statements, Empty())
		statements = append(statements, r.generateFetchContextReversedByParentUpdates(fp, fpr, rcv, entity)...)
	}

	r.f.Add(block(
		Add(r.generateFetchContextMethodBase(rcv), r.generateStateAdderMethodSignature(sc, entity)).Block(statements...),
	))
}

func (r *Renderer) generateFetchContextReversedByParentUpdates(fp plan.FetchParent, fpr plan.FetchParentReverse, rcv, entityArg Code) []Code {
	parentID := Id("pid")
	reversedParentSet := Id("s")
	swallow := Id("_")
	ok := Id("ok")
	parentSet := Add(rcv).Dot(r.naming.FetchContext.FieldByParentID[fp.ID]).Index(parentID)
	e := r.plan.Model.Entities[fp.Entity]
	fprsc := r.plan.ReversedStateContainers[fpr.StateContainer]

	return []Code{
		For(List(parentID, reversedParentSet).Op(":=").Range().Add(rcv).Dot(r.naming.FetchContext.FieldByParentID[fpr.FetchParent])).Block(
			If(List(swallow, ok).Op(":=").Add(reversedParentSet).Index(r.members.Member(entityArg, fprsc.HolderEntityMember)), Op("!").Add(ok)).Block(
				If(List(swallow, ok).Op(":=").Add(parentSet), Op("!").Add(ok)).Block(
					Add(parentSet).Op("=").Make(r.generateFetchContextParentLevelSetType(fp)),
				),
			),
			Empty(),
			Add(parentSet).Index(r.members.Member(entityArg, e.IDMember)).Op("=").Struct().Values(),
		),
	}
}

func (r *Renderer) renderFetchContextParentGetterMethods(rcv Code) {
	for _, pfg := range r.plan.ParentFetchGetters {
		r.renderFetchContextParentGetterMethod(pfg, rcv)
	}
}

func (r *Renderer) renderFetchContextParentGetterMethod(pfg plan.ParentFetchGetter, rcv Code) {
	levelID := Id("lid")

	r.f.Add(block(
		Add(r.generateFetchContextMethodBase(rcv), r.generateFetchContextParentGetterMethodSignature(pfg, levelID)).Block(
			r.generateFetchContextParentGetterMethodBody(pfg, rcv, levelID)...,
		),
	))
}

func (r *Renderer) generateFetchContextParentGetterMethodSignature(pfg plan.ParentFetchGetter, levelIDArg Code) Code {
	fp := r.plan.FetchParents[pfg.FetchParent]
	e := r.plan.Model.Entities[fp.Entity]

	return Id(r.naming.FetchContext.ParentGetterMethod[fp.ID]).Params(Add(levelIDArg, libID)).Params(Add(iterSeq).Types(r.types[e.Type]))
}

func (r *Renderer) generateFetchContextParentGetterMethodBody(pfg plan.ParentFetchGetter, rcv, levelIDArg Code) []Code {
	fp := r.plan.FetchParents[pfg.FetchParent]
	e := r.plan.Model.Entities[fp.Entity]

	yield := Id("yield")
	id := Id("id")
	entity := Id("e")

	return []Code{
		Return(
			Func().Params(Add(yield, Func().Params(r.types[e.Type]).Params(Bool()))).Block(
				For(Add(id).Op(":=").Range().Add(rcv).Dot(r.naming.FetchContext.FieldByParentID[fp.ID]).Index(levelIDArg)).Block(
					Add(entity).Op(":=").Add(rcv).Dot(r.naming.State.StateContainerField[pfg.StateContainer]).Index(id),
					Empty(),
					If(Add(entity).Dot("Fetched").Call().Op("&&").Op("!").Add(yield).Call(Add(entity).Dot("Value").Call())).Block(
						Return(),
					),
				),
			),
		),
	}
}

func (r *Renderer) generateFetchContextParentGetterMethodCall(fpid plan.FetchParentID, rcv, parentIDArg Code) Code {
	return Add(rcv).Dot(r.naming.FetchContext.ParentGetterMethod[fpid]).Call(parentIDArg)
}

func (r *Renderer) renderFetchContextFlushMethod(rcv Code) {
	statements := make([]Code, 0, len(r.plan.FetchChildren))

	for _, fc := range r.plan.FetchChildren {
		field := Add(rcv).Dot(r.naming.FetchContext.FieldPending[fc.ID])
		statements = append(statements, Add(field).Op("=").Add(field).Index(Empty(), Lit(0)))
	}

	r.f.Add(block(
		Add(r.generateFetchContextMethodBase(rcv), Id(r.naming.FetchContext.FlushMethod)).Params().Block(statements...),
	))
}

func (r *Renderer) generateFetchContextFlushMethodCall(rcv Code) Code {
	return Add(rcv).Dot(r.naming.FetchContext.FlushMethod).Call()
}
