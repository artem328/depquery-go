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
	r.renderFetchContextAddNestedMethods(rcv)
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

	if len(r.plan.SyntheticStateContainers) > 0 {
		fields = append(fields, Empty())

		for _, ssc := range r.plan.SyntheticStateContainers {
			fields = append(fields, Id(r.naming.FetchContext.FieldSyntheticState[ssc.ID]).Add(r.generateFetchContextSyntheticStateFieldType(ssc)))
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
	var idType Code

	e := r.plan.Model.Entities[fp.Entity]
	if e.Synthetic {
		idType = Uint64()
	} else if fp.Reversed {
		re := r.plan.Model.Entities[fp.ReversedByEntity]
		idType = r.types[r.plan.Model.Members[re.IDMember].Type]
	} else {
		idType = r.types[r.plan.Model.Members[e.IDMember].Type]
	}

	return Map(idType).Struct()
}

func (r *Renderer) generateFetchContextSyntheticStateFieldType(ssc plan.SyntheticStateContainer) Code {
	e := r.plan.Model.Entities[ssc.Entity]

	return Map(Uint64()).Add(r.types[e.Type])
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

	for _, ssc := range r.plan.SyntheticStateContainers {
		inits = append(inits, Id(r.naming.FetchContext.FieldSyntheticState[ssc.ID]).Op(":").Make(r.generateFetchContextSyntheticStateFieldType(ssc)))
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
	iter := Id("i")
	fetchContext := Id("ctx")
	zeroLevel := Id("zero")

	var statements []Code

	statements = append(statements,
		Add(fetchContext).Op(":=").Add(r.generateFetchContextConstructorCall()),
		Empty(),
	)
	statements = append(statements, r.generateFetchContextInstanceConstructorZeroLevel(fcr, zeroLevel, iter, fetchContext)...)
	statements = append(statements,
		Empty(),
		Add(fetchContext).Dot(r.naming.FetchContext.FieldByParentID[fcr.FetchParent]).Index(Lit(0)).Op("=").Add(zeroLevel),
		Empty(),
		Return(fetchContext),
	)

	r.f.Add(block(
		Func().Add(r.generateFetchContextConstructorForEntitySignature(fcr, iter)).Block(statements...),
	))
}

func (r *Renderer) generateFetchContextInstanceConstructorZeroLevel(fcr plan.FetchContextRoot, zeroLevel, iter, fetchContext Code) []Code {
	e := r.plan.Model.Entities[fcr.Entity]

	entity := Id("x")
	entityID := Id("xid")
	counter := Id("_sid")

	var statements []Code

	var (
		entityIDDef     Code
		entityIDPostDef Code
		stateEntity     Code
		stateField      string
	)
	if fcr.Synthetic {
		ssc := r.plan.SyntheticStateContainers[fcr.SyntheticStateContainer]

		statements = append(statements,
			Const().Id(r.naming.FetchContext.SyntheticNamespaceConst).Op("=").Lit(ssc.IDNamespace),
			Empty(),
			Var().Add(counter).Int(),
		)
		entityIDDef = Add(libSyntheticID).Call(Id(r.naming.FetchContext.SyntheticNamespaceConst), Add(libIBytes).Call(counter))
		entityIDPostDef = Add(counter).Op("++")
		stateField = r.naming.FetchContext.FieldSyntheticState[fcr.SyntheticStateContainer]
		stateEntity = entity
	} else {
		entityIDDef = r.members.Member(entity, e.IDMember)
		entityIDPostDef = Null()
		stateField = r.naming.State.StateContainerField[fcr.StateContainer]
		stateEntity = Add(libFetched).Call(entity)
	}

	statements = append(statements,
		Add(zeroLevel).Op(":=").Make(r.generateFetchContextParentLevelSetType(r.plan.FetchParents[fcr.FetchParent])),
		For(Add(entity).Op(":=").Range().Add(iter)).Block(
			Add(entityID).Op(":=").Add(entityIDDef),
			Add(entityIDPostDef),
			Add(zeroLevel).Index(entityID).Op("=").Struct().Values(),
			Add(fetchContext).Dot(stateField).Index(entityID).Op("=").Add(stateEntity),
		),
	)

	return statements
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
		statements = append(statements, r.generateFetchContextParentLevelSetInit(p, rcv, relationIDArg, idArg)...)
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

func (r *Renderer) generateFetchContextParentLevelSetInit(fp plan.FetchParent, rcv, parentIDArg, idArg Code) []Code {
	byParentIDSet := Add(rcv).Dot(r.naming.FetchContext.FieldByParentID[fp.ID]).Index(parentIDArg)

	ok := Id("ok")
	swallow := Id("_")

	return []Code{
		If(List(swallow, ok).Op(":=").Add(byParentIDSet), Op("!").Add(ok)).Block(
			Add(byParentIDSet).Op("=").Add(Make(r.generateFetchContextParentLevelSetType(fp))),
		),
		Empty(),
		Add(byParentIDSet).Index(idArg).Op("=").Struct().Values(),
	}
}

func (r *Renderer) renderFetchContextAddNestedMethods(rcv Code) {
	for _, nef := range r.plan.NestedEntityFetches {
		if nef.Synthetic {
			r.renderFetchContextAddNestedSyntheticMethod(nef, rcv)
		} else {
			r.renderFetchContextAddNestedMethod(nef, rcv)
		}
	}
}

func (r *Renderer) renderFetchContextAddNestedSyntheticMethod(nef plan.NestedEntityFetch, rcv Code) {
	nested := Id("n")
	parentLevelID := Id("pid")
	id := Id("id")
	swallow := Id("_")
	ok := Id("ok")

	var body []Code

	body = append(body, r.generateFetchContextParentLevelSetInit(r.plan.FetchParents[nef.Parent], rcv, parentLevelID, id)...)

	nestedLocation := Add(rcv).Dot(r.naming.FetchContext.FieldSyntheticState[nef.SyntheticStateContainerID]).Index(id)
	body = append(body,
		Empty(),
		If(List(swallow, ok).Op(":=").Add(nestedLocation), Op("!").Add(ok)).Block(
			Add(nestedLocation).Op("=").Add(nested),
		),
	)

	r.f.Add(block(
		Add(r.generateFetchContextMethodBase(rcv), r.generateFetchContextAddNestedSyntheticMethodSignature(nef, nested, parentLevelID, id)).Block(body...),
	))
}

func (r *Renderer) generateFetchContextAddNestedSyntheticMethodSignature(nef plan.NestedEntityFetch, nestedParam, entityLevelIDParam, idParam Code) Code {
	e := r.plan.Model.Entities[nef.Entity]

	return Id(r.naming.FetchContext.AddNestedMethod[nef.ID]).Params(Add(nestedParam, r.types[e.Type]), Add(entityLevelIDParam, libID), Add(idParam, Uint64())).Params()
}

func (r *Renderer) generateFetchContextAddNestedSyntheticMethodCall(nefid plan.NestedEntityFetchID, rcv, nestedArg, entityLevelIDArg, idArg Code) Code {
	return Add(rcv).Dot(r.naming.FetchContext.AddNestedMethod[nefid]).Call(nestedArg, entityLevelIDArg, idArg)
}

func (r *Renderer) renderFetchContextAddNestedMethod(nef plan.NestedEntityFetch, rcv Code) {
	e := r.plan.Model.Entities[nef.Entity]

	id := Id("id")
	nested := Id("n")
	entityLevelID := Id("eid")
	maybeFetched := Id("mf")
	ok := Id("ok")

	var body []Code

	body = append(body,
		Add(id).Op(":=").Add(r.members.Member(nested, e.IDMember)),
		Empty(),
	)
	body = append(body, r.generateFetchContextParentLevelSetInit(r.plan.FetchParents[nef.Parent], rcv, entityLevelID, id)...)

	nestedLocation := Add(rcv).Dot(r.naming.State.StateContainerField[nef.StateContainer]).Index(id)
	body = append(body,
		Empty(),
		If(List(maybeFetched, ok).Op(":=").Add(nestedLocation), Op("!").Add(ok).Op("||").Op("!").Add(maybeFetched).Dot("Fetched").Call()).Block(
			Add(nestedLocation).Op("=").Add(libFetched).Call(nested),
		),
	)

	r.f.Add(block(
		Add(r.generateFetchContextMethodBase(rcv), r.generateFetchContextAddNestedMethodSignature(nef, nested, entityLevelID)).Block(body...),
	))
}

func (r *Renderer) generateFetchContextAddNestedMethodSignature(nef plan.NestedEntityFetch, nestedParam, entityLevelIDParam Code) Code {
	e := r.plan.Model.Entities[nef.Entity]

	return Id(r.naming.FetchContext.AddNestedMethod[nef.ID]).Params(Add(nestedParam, r.types[e.Type]), Add(entityLevelIDParam, libID)).Params()
}

func (r *Renderer) generateFetchContextAddNestedMethodCall(nefid plan.NestedEntityFetchID, rcv Code, nestedArg, entityLevelIDArg Code) Code {
	return Add(rcv).Dot(r.naming.FetchContext.AddNestedMethod[nefid]).Call(nestedArg, entityLevelIDArg)
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
	return Id(r.naming.FetchContext.ParentGetterMethod[pfg.ID]).Params(Add(levelIDArg, libID)).Params(r.generateFetchContextParentGetterReturnParam(pfg))
}

func (r *Renderer) generateFetchContextParentGetterReturnParam(pfg plan.ParentFetchGetter) Code {
	fp := r.plan.FetchParents[pfg.FetchParent]
	e := r.plan.Model.Entities[fp.Entity]

	if pfg.Synthetic {
		return Add(iterSeq2).Types(Uint64(), r.types[e.Type])
	}

	return Add(iterSeq).Types(r.types[e.Type])
}

func (r *Renderer) generateFetchContextParentGetterMethodBody(pfg plan.ParentFetchGetter, rcv, levelIDArg Code) []Code {
	fp := r.plan.FetchParents[pfg.FetchParent]
	e := r.plan.Model.Entities[fp.Entity]

	yield := Id("yield")
	id := Id("id")
	entity := Id("e")

	var (
		yieldParams      []Code
		contextFieldName string
		yieldBlock       Code
	)

	if pfg.Synthetic {
		yieldParams = []Code{Uint64(), r.types[e.Type]}
		contextFieldName = r.naming.FetchContext.FieldSyntheticState[pfg.SyntheticStateContainer]
		yieldBlock = If(Op("!").Add(yield).Call(id, entity)).Block(
			Return(),
		)
	} else {
		yieldParams = []Code{r.types[e.Type]}
		contextFieldName = r.naming.State.StateContainerField[pfg.StateContainer]
		yieldBlock = If(Add(entity).Dot("Fetched").Call().Op("&&").Op("!").Add(yield).Call(Add(entity).Dot("Value").Call())).Block(
			Return(),
		)
	}

	return []Code{
		Return(
			Func().Params(Add(yield, Func().Params(yieldParams...).Params(Bool()))).Block(
				For(Add(id).Op(":=").Range().Add(rcv).Dot(r.naming.FetchContext.FieldByParentID[fp.ID]).Index(levelIDArg)).Block(
					Add(entity).Op(":=").Add(rcv).Dot(contextFieldName).Index(id),
					Empty(),
					yieldBlock,
				),
			),
		),
	}
}

func (r *Renderer) generateFetchContextParentGetterMethodCall(pfgid plan.ParentFetchGetterID, rcv, parentIDArg Code) Code {
	return Add(rcv).Dot(r.naming.FetchContext.ParentGetterMethod[pfgid]).Call(parentIDArg)
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
