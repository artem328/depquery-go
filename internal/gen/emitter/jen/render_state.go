package jen

import (
	"github.com/artem328/depquery-go/internal/gen/plan"
	. "github.com/dave/jennifer/jen"
)

func (r *Renderer) renderStateInterface() {
	r.f.Add(block(
		Type().Id(r.naming.State.Interface).Interface(r.generateStateMethodSignatures()...),
	))
}

func (r *Renderer) generateStateMethodSignatures() []Code {
	var methods []Code

	for _, sc := range r.plan.StateContainers {
		methods = append(methods, r.generateStateResolveMethodSignature(sc, Null()))
		methods = append(methods, r.generateStateShouldResolveMethodSignature(sc, Null()))
	}

	for _, rsc := range r.plan.ReversedStateContainers {
		methods = append(methods, r.generateStateResolveReversedByMethodSignature(rsc, Null()))
		methods = append(methods, r.generateStateShouldResolveReversedByMethodSignature(rsc, Null()))
	}

	return methods
}

func (r *Renderer) generateStateResolveMethodSignature(sc plan.StateContainer, idParam Code) Code {
	return r.generateStateGenericResolveMethodSignature(sc, r.naming.State.ResolveMethod, idParam, Bool())
}

func (r *Renderer) generateStateShouldResolveMethodSignature(sc plan.StateContainer, idParam Code) Code {
	return r.generateStateGenericResolveMethodSignature(sc, r.naming.State.ShouldResolveMethod, idParam, Error())
}

func (r *Renderer) generateStateGenericResolveMethodSignature(sc plan.StateContainer, naming []string, idParam, secondReturnParam Code) Code {
	e := r.plan.Model.Entities[sc.Entity]
	im := r.plan.Model.Members[e.IDMember]

	return Id(naming[sc.ID]).Params(Add(idParam, r.types[im.Type])).Params(r.types[e.Type], secondReturnParam)
}

func (r *Renderer) generateStateResolveReversedByMethodSignature(rsc plan.ReversedStateContainer, idParam Code) Code {
	return r.generateStateGenericResolveReversedByMethodSignature(rsc, r.naming.State.ResolveReversedByMethod, idParam, Bool())
}

func (r *Renderer) generateStateShouldResolveReversedByMethodSignature(rsc plan.ReversedStateContainer, idParam Code) Code {
	return r.generateStateGenericResolveReversedByMethodSignature(rsc, r.naming.State.ShouldResolveReversedByMethod, idParam, Error())
}

func (r *Renderer) generateStateGenericResolveReversedByMethodSignature(rsc plan.ReversedStateContainer, naming []string, idParam, secondReturnParam Code) Code {
	sc := r.plan.StateContainers[rsc.StateContainerID]
	e := r.plan.Model.Entities[sc.Entity]
	rm := r.plan.Model.Members[rsc.HolderEntityMember]

	return Id(naming[rsc.ID]).Params(Add(idParam, r.types[rm.Type])).Params(r.types[e.Type], secondReturnParam)
}

func (r *Renderer) renderStateImplementation() {
	r.renderStateStruct()
	r.renderStateConstructor()
	r.renderStateMethods()
}

func (r *Renderer) renderStateStruct() {
	var fields []Code

	for _, sc := range r.plan.StateContainers {
		fields = append(fields, r.generateStateContainerField(sc))
	}

	for _, sc := range r.plan.StateContainers {
		if len(sc.ReversedBy) == 0 {
			continue
		}

		if len(fields) > 0 {
			fields = append(fields, Empty())
		}

		for _, rscid := range sc.ReversedBy {
			rsc := r.plan.ReversedStateContainers[rscid]
			fields = append(fields, r.generateStateContainerReversedByField(rsc))
		}
	}

	r.f.Add(block(
		Type().Id(r.naming.State.Struct).Struct(fields...),
	))
}

func (r *Renderer) generateStateContainerField(sc plan.StateContainer) Code {
	return Id(r.naming.State.StateContainerField[sc.ID]).Add(r.generateStateContainerFieldType(sc))
}

func (r *Renderer) generateStateContainerFieldType(sc plan.StateContainer) Code {
	e := r.plan.Model.Entities[sc.Entity]
	im := r.plan.Model.Members[e.IDMember]

	return Map(r.types[im.Type]).Add(libMaybeFetched).Types(r.types[e.Type])
}

func (r *Renderer) generateStateContainerReversedByField(rsc plan.ReversedStateContainer) Code {
	return Id(r.naming.State.StateContainerReversedByField[rsc.ID]).Add(r.generateStateContainerReversedByFieldType(rsc))
}

func (r *Renderer) generateStateContainerReversedByFieldType(rsc plan.ReversedStateContainer) Code {
	sc := r.plan.StateContainers[rsc.StateContainerID]
	e := r.plan.Model.Entities[sc.Entity]
	hem := r.plan.Model.Members[rsc.HolderEntityMember]
	im := r.plan.Model.Members[e.IDMember]

	return Map(r.types[hem.Type]).Add(r.types[im.Type])
}

func (r *Renderer) renderStateConstructor() {
	var inits []Code
	for _, sc := range r.plan.StateContainers {
		inits = append(inits, Id(r.naming.State.StateContainerField[sc.ID]).Op(":").Make(r.generateStateContainerFieldType(sc)))
	}

	for _, rsc := range r.plan.ReversedStateContainers {
		inits = append(inits, Id(r.naming.State.StateContainerReversedByField[rsc.ID]).Op(":").Make(r.generateStateContainerReversedByFieldType(rsc)))
	}

	r.f.Add(block(
		Func().Id(r.naming.State.Constructor).Params().Params(Id(r.naming.State.Struct)).Block(
			Return(Id(r.naming.State.Struct).Add(valuesMultiline(inits...))),
		),
	))
}

func (r *Renderer) generateStateConstructorCall() Code {
	return Id(r.naming.State.Constructor).Call()
}

func (r *Renderer) renderStateMethods() {
	rcv := Id("s")
	id := Id("id")
	entity := Id("e")

	r.renderStateGetterMethods(rcv, id)
	r.renderStateAdderMethods(rcv, entity)
}

func (r *Renderer) renderStateGetterMethods(rcv, idArg Code) {
	for _, sc := range r.plan.StateContainers {
		r.f.Add(block(
			Add(r.generateStateMethodBase(rcv), r.generateStateResolveMethodSignature(sc, idArg)).Block(r.generateStateResolveMethodBody(sc, rcv, idArg)...),
		))

		r.f.Add(block(
			Add(r.generateStateMethodBase(rcv), r.generateStateShouldResolveMethodSignature(sc, idArg)).Block(r.generateStateShouldResolveMethodBody(sc, rcv, idArg)...),
		))
	}

	for _, rsc := range r.plan.ReversedStateContainers {
		r.f.Add(block(
			Add(r.generateStateMethodBase(rcv), r.generateStateResolveReversedByMethodSignature(rsc, idArg)).Block(r.generateStateResolveReversedByMethodBody(rsc, rcv, idArg)...),
		))

		r.f.Add(block(
			Add(r.generateStateMethodBase(rcv), r.generateStateShouldResolveReversedByMethodSignature(rsc, idArg)).Block(r.generateStateShouldResolveReversedByMethodBody(rsc, rcv, idArg)...),
		))
	}
}

func (r *Renderer) renderStateAdderMethods(rcv, entityArg Code) {
	for _, sc := range r.plan.StateContainers {
		r.f.Add(block(
			Add(r.generateStateMethodBase(rcv), r.generateStateAdderMethodSignature(sc, entityArg)).Block(
				r.generateStateAdderMethodAdders(sc, rcv, entityArg)...,
			),
		))
	}
}

func (r *Renderer) generateStateAdderMethodSignature(sc plan.StateContainer, entityParam Code) Code {
	e := r.plan.Model.Entities[sc.Entity]

	return Id(r.naming.State.AdderMethod[sc.ID]).Params(Add(entityParam, r.types[e.Type]))
}

func (r *Renderer) generateStateAdderMethodCall(scid plan.StateContainerID, rcv, entityArg Code) Code {
	return Add(rcv).Dot(r.naming.State.AdderMethod[scid]).Call(entityArg)
}

func (r *Renderer) generateStateAdderMethodAdders(sc plan.StateContainer, rcv Code, entityArg Code) []Code {
	e := r.plan.Model.Entities[sc.Entity]

	adders := make([]Code, 0, 1+1+len(sc.ReversedBy))
	adders = append(adders, Add(rcv).Dot(r.naming.State.StateContainerField[sc.ID]).Index(r.members.Member(entityArg, e.IDMember)).Op("=").Add(libFetched).Call(entityArg))

	if len(sc.ReversedBy) == 0 {
		return adders
	}

	adders = append(adders, Empty())
	for _, rscid := range sc.ReversedBy {
		rsc := r.plan.ReversedStateContainers[rscid]
		adders = append(adders, Add(rcv).Dot(r.naming.State.StateContainerReversedByField[rsc.ID]).Index(r.members.Member(entityArg, rsc.HolderEntityMember)).Op("=").Add(entityArg, r.members[e.IDMember]))
	}

	return adders
}

func (r *Renderer) generateStateMethodBase(rcv Code) Code {
	return Func().Params(Add(rcv).Id(r.naming.State.Struct))
}

func (r *Renderer) generateStateResolveMethodBody(sc plan.StateContainer, rcv, idArg Code) []Code {
	return []Code{
		Return(Add(rcv).Dot(r.naming.State.StateContainerField[sc.ID]).Index(idArg).Dot("Maybe").Call()),
	}
}

func (r *Renderer) generateStateShouldResolveMethodBody(sc plan.StateContainer, rcv, idArg Code) []Code {
	e := r.plan.Model.Entities[sc.Entity]

	entity := Id("e")
	ok := Id("ok")

	return []Code{
		Var().Defs(
			Add(entity, r.types[e.Type]),
			Add(ok, Bool()),
		),
		Empty(),
		If(List(entity, ok).Op("=").Add(rcv).Dot(r.naming.State.StateContainerField[sc.ID]).Index(idArg).Dot("Maybe").Call(), ok).Block(
			Return(entity, Nil()),
		),
		Empty(),
		Return(entity, Add(libNotInStateErrorConstructor).Call(Lit(e.Name), idArg)),
	}
}

func (r *Renderer) generateStateResolveReversedByMethodBody(rsc plan.ReversedStateContainer, rcv, idArg Code) []Code {
	return []Code{
		Return(
			Add(rcv).Dot(r.naming.State.ResolveMethod[rsc.StateContainerID]).Call(
				Add(rcv).Dot(r.naming.State.StateContainerReversedByField[rsc.ID]).Index(idArg),
			),
		),
	}
}

func (r *Renderer) generateStateShouldResolveReversedByMethodBody(rsc plan.ReversedStateContainer, rcv, idArg Code) []Code {
	return []Code{
		Return(
			Add(rcv).Dot(r.naming.State.ShouldResolveMethod[rsc.StateContainerID]).Call(
				Add(rcv).Dot(r.naming.State.StateContainerReversedByField[rsc.ID]).Index(idArg),
			),
		),
	}
}
