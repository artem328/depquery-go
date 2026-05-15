package semantic

import (
	"fmt"
	"go/types"
	"iter"

	"github.com/artem328/depquery-go/internal/gen/schema"
)

type memberKey struct {
	Entity EntityID
	Name   string
}

type memberEntry struct {
	ID   MemberID
	Type types.Type
}

type memberResolver struct {
	Struct   *types.Struct
	IsStruct bool
	Methods  func() iter.Seq[*types.Func]
	TypeName *types.TypeName
}

func (r *Resolver) findOrCreateMember(entity string, t types.Type, name schema.Value[string]) (MemberID, types.Type) {
	key := memberKey{
		Entity: r.seenEntities[entity].ID,
		Name:   name.V,
	}

	if m, ok := r.seenMembers[key]; ok {
		return m.ID, m.Type
	}

	named, ok := t.(*types.Named)
	if !ok {
		r.eerr(entity, fmt.Sprintf("expected named type to find its member `%s` (%s is not)", name.V, t.String()), name.Definition)
		return 0, nil
	}

	typ := named.Underlying()

	for {
		if a, ok := typ.(*types.Alias); ok {
			typ = a.Underlying()

			continue
		}

		break
	}

	var mr memberResolver

	mr.TypeName = named.Obj()

	switch tt := typ.(type) {
	case *types.Struct:
		mr.Struct = tt
		mr.IsStruct = true
		mr.Methods = func() iter.Seq[*types.Func] {
			return structMethodIter(named)
		}
	case *types.Interface:
		mr.Methods = tt.Methods
	default:
		r.eerr(entity, fmt.Sprintf("member `%s` cannot be found in %s since it is neither struct nor interface", name.V, mr.TypeName.Type().String()), name.Definition)

		return 0, nil
	}

	mk, mt, ok := r.resolveMemberType(entity, mr, name)
	if !ok {
		return 0, nil
	}

	rt := r.decomposeType(mt)

	id := MemberID(len(r.model.Members))
	r.seenMembers[key] = memberEntry{
		ID:   id,
		Type: mt,
	}
	r.model.Members = append(r.model.Members, Member{
		Name: name.V,
		Type: r.findOrCreateTypeRecursively(rt),
		Kind: mk,
	})

	return id, mt
}

func (r *Resolver) resolveMemberType(entity string, mr memberResolver, name schema.Value[string]) (MemberKind, types.Type, bool) {
	if t, valid, found := r.resolveMethodMemberType(entity, mr.Methods(), name); valid {
		return MemberKindMethod, t, true
	} else if found {
		return 0, nil, false
	}

	if mr.IsStruct {
		if t, valid, _ := r.resolveFieldMemberType(entity, mr.Struct, name, mr.TypeName); valid {
			return MemberKindField, t, true
		}
	}

	r.eerr(entity, fmt.Sprintf("member `%s` is not found in %s", name.V, mr.TypeName.Type().String()), name.Definition)

	return 0, nil, false
}

func (r *Resolver) resolveMethodMemberType(entity string, methods iter.Seq[*types.Func], name schema.Value[string]) (_ types.Type, valid, found bool) {
	var invalid bool

	for method := range methods {
		if method.Name() != name.V {
			continue
		}

		if !method.Exported() {
			r.eerr(
				entity,
				fmt.Sprintf("method %s of %s is not exported", method.Name(), method.Signature().Recv().Type().String()),
				name.Definition,
			)
			invalid = true
		}

		res := method.Signature().Results()
		if res.Len() != 1 {
			r.eerr(
				entity,
				fmt.Sprintf(
					"method %s of %s must return exactly one value. got %d",
					method.Name(),
					method.Signature().Recv().Type().String(),
					res.Len(),
				),
				name.Definition,
			)

			invalid = true
		}

		params := method.Signature().Params()
		if params.Len() > 0 {
			r.eerr(
				entity,
				fmt.Sprintf(
					"method %s of %s must not accept any parameters. got %d",
					method.Name(),
					method.Signature().Recv().Type().String(),
					params.Len(),
				),
				name.Definition,
			)

			invalid = true
		}

		if invalid {
			return nil, false, true
		}

		return res.At(0).Type(), true, true
	}

	return nil, false, false
}

func (r *Resolver) resolveFieldMemberType(entity string, s *types.Struct, name schema.Value[string], loc *types.TypeName) (_ types.Type, valid, found bool) {
	for field := range s.Fields() {
		if field.Name() != name.V {
			continue
		}

		if !field.Exported() {
			r.eerr(entity, fmt.Sprintf("field %s of %s is not exported", field.Name(), loc.Type().String()), name.Definition)

			return nil, false, true
		}

		return field.Type(), true, true
	}

	return nil, false, false
}

func structMethodIter(named *types.Named) iter.Seq[*types.Func] {
	iters := [...]func() iter.Seq[*types.Selection]{
		func() iter.Seq[*types.Selection] {
			return types.NewMethodSet(named).Methods()
		},
		func() iter.Seq[*types.Selection] {
			return types.NewMethodSet(types.NewPointer(named)).Methods()
		},
	}

	return func(yield func(*types.Func) bool) {
		for _, it := range iters {
			for method := range it() {
				if f, ok := method.Obj().(*types.Func); ok && !yield(f) {
					return
				}
			}
		}
	}
}
